package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jamesstocktonj1/ticker-provider/bindings/jamesstocktonj1/ticker/ticker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.wasmcloud.dev/provider"
)

var (
	tracer = otel.Tracer("ticker-provider")

	ErrTickerNotFound = errors.New("error ticker task not found")
)

type Ticker struct {
	provider *provider.WasmcloudProvider
	tasks    gocron.Scheduler
	taskList map[string]uuid.UUID
}

func CreateTicker() (*Ticker, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Ticker{
		tasks:    s,
		taskList: make(map[string]uuid.UUID),
	}, nil
}

func (t *Ticker) Start() error {
	_, span := tracer.Start(context.Background(), "Start")
	defer span.End()

	t.tasks.Start()
	return nil
}

func (t *Ticker) Shutdown() error {
	err := t.tasks.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

func (t *Ticker) TaskFunc(taskId string) error {
	ctx, span := tracer.Start(context.Background(), "TaskFunc")
	span.SetAttributes(attribute.String("task_id", taskId))
	defer span.End()

	t.provider.Logger.Info("task execute", "task_id", taskId)

	taskErr, err := ticker.Task(
		injectTraceHeader(ctx),
		t.provider.OutgoingRpcClient(taskId),
	)
	if err != nil || taskErr == nil {
		t.provider.Logger.Error("error: ticker.Task", "error", err, "task_id", taskId)
		span.RecordError(err)
		return err
	} else if taskErr.Discriminant() != ticker.TaskErrorNone {
		err := errors.New(taskErr.String())
		t.provider.Logger.Error("error: ticker.Task TaskError", "error", err, "task_id", taskId)
		span.RecordError(err)
		return err
	}

	return nil
}

func (t *Ticker) handlePutTargetLink(link provider.InterfaceLinkDefinition) error {
	t.provider.Logger.Info("handlePutTargetLink", "link", link)

	jobDef, err := newSchedulerJob(link.TargetConfig)
	if err != nil {
		return err
	}

	job, err := t.tasks.NewJob(
		jobDef,
		gocron.NewTask(t.TaskFunc, link.SourceID),
	)
	if err != nil {
		return err
	}
	t.taskList[link.SourceID] = job.ID()

	return nil
}

func (t *Ticker) handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	t.provider.Logger.Info("handleDelTargetLink", "link", link)

	taskId, ok := t.taskList[link.SourceID]
	if !ok {
		return ErrTickerNotFound
	}
	t.tasks.RemoveJob(taskId)

	return nil
}

func (t *Ticker) handleHealthCheck() string {
	h := provider.HealthCheckResponse{
		Healthy: true,
		Message: "healthy",
	}

	data, err := json.Marshal(&h)
	if err != nil {
		return "unhealthy"
	}
	return string(data)
}

func (t *Ticker) handleShutdown() error {
	return nil
}
