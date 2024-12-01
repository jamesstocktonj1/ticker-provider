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

const (
	OtelName = "ticker-provider"
)

var (
	tracer = otel.Tracer(OtelName)

	ErrTickerNotFound = errors.New("error ticker task not found")
)

type Ticker struct {
	provider *provider.WasmcloudProvider
	tasks    gocron.Scheduler
	taskList map[string]*TickerTask
}

type TickerTask struct {
	Component string
	ID        uuid.UUID
	Type      string
}

func CreateTicker() (*Ticker, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Ticker{
		tasks:    s,
		taskList: make(map[string]*TickerTask),
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

func (t *Ticker) TaskFunc(task *TickerTask) error {
	ctx, span := tracer.Start(context.Background(), "TaskFunc")
	span.SetAttributes(
		attribute.String("id", task.ID.String()),
		attribute.String("component", task.Component),
		attribute.String("type", task.Type),
	)
	defer span.End()

	t.provider.Logger.Info("task execute", "id", task.ID.String(), "component", task.Component, "type", task.Type)

	taskErr, err := ticker.Task(
		injectTraceHeader(ctx),
		t.provider.OutgoingRpcClient(task.Component),
	)
	if err != nil || taskErr == nil {
		t.provider.Logger.Error("error: ticker.Task", "error", err, "id", task.ID.String())
		span.RecordError(err)
		return err
	} else if taskErr.Discriminant() != ticker.TaskErrorNone {
		err := errors.New(taskErr.String())
		t.provider.Logger.Error("error: ticker.Task TaskError", "error", err, "id", task.ID.String())
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

	jobKey := getJobKey(link)
	jobCtx := &TickerTask{
		Component: link.SourceID,
		Type:      link.TargetConfig[configTypeKey],
	}

	job, err := t.tasks.NewJob(
		jobDef,
		gocron.NewTask(t.TaskFunc, jobCtx),
	)
	if err != nil {
		return err
	}
	jobCtx.ID = job.ID()
	t.taskList[jobKey] = jobCtx

	return nil
}

func (t *Ticker) handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	t.provider.Logger.Info("handleDelTargetLink", "link", link)

	jobKey := getJobKey(link)
	taskId, ok := t.taskList[jobKey]
	if !ok {
		return ErrTickerNotFound
	}

	err := t.tasks.RemoveJob(taskId.ID)
	if err != nil {
		return err
	}

	delete(t.taskList, jobKey)
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
