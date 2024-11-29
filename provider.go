package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jamesstocktonj1/ticker-provider/bindings/jamesstocktonj1/ticker/ticker"
	"go.wasmcloud.dev/provider"
)

const (
	IntervalConfigKey = "interval"
)

var (
	ErrIntervalNotSpecified  = errors.New("error time interval not configured")
	ErrIntervalFormatInvalid = errors.New("error time interval is not correctly formatted")
	ErrTickerNotFound        = errors.New("error ticker task not found")
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
	t.provider.Logger.Info("task execute", "task_id", taskId)

	taskErr, err := ticker.Task(
		context.Background(),
		t.provider.OutgoingRpcClient(taskId),
	)
	if err != nil || taskErr == nil {
		t.provider.Logger.Error("error: ticker.Task", "error", err, "task_id", taskId)
		return err
	} else if taskErr.Discriminant() != ticker.TaskErrorNone {
		t.provider.Logger.Error("error: ticker.Task TaskError", "error", err, "task_id", taskId)
		return errors.New(taskErr.String())
	}

	return nil
}

func (t *Ticker) handlePutTargetLink(link provider.InterfaceLinkDefinition) error {
	t.provider.Logger.Info("handlePutTargetLink", "link", link)
	timeInterval, ok := link.TargetConfig[IntervalConfigKey]
	if !ok {
		return ErrIntervalNotSpecified
	}

	timeDuration, err := time.ParseDuration(timeInterval)
	if err != nil {
		return ErrIntervalFormatInvalid
	}

	job, err := t.tasks.NewJob(
		gocron.DurationJob(timeDuration),
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
