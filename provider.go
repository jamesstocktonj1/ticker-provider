package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jamesstocktonj1/ticker-provider/bindings/jamesstocktonj1/ticker/ticker"
	"github.com/madflojo/tasks"
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
	tasks    *tasks.Scheduler
}

func CreateTicker() (*Ticker, error) {
	t := Ticker{
		tasks: tasks.New(),
	}

	return &t, nil
}

func (t *Ticker) Shutdown() error {
	t.tasks.Stop()
	return nil
}

func (t *Ticker) TaskFunc(task tasks.TaskContext) error {
	t.provider.Logger.Info("task execute", "task_id", task.ID())

	taskErr, err := ticker.Task(
		context.Background(),
		t.provider.OutgoingRpcClient(task.ID()),
	)
	if err != nil || taskErr == nil {
		t.provider.Logger.Error("error: ticker.Task", "error", err, "task_id", task.ID())
		return err
	} else if taskErr.Discriminant() != ticker.TaskErrorNone {
		t.provider.Logger.Error("error: ticker.Task TaskError", "error", err, "task_id", task.ID())
		return errors.New(taskErr.String())
	}

	return nil
}

func (t *Ticker) ErrorFunc(task tasks.TaskContext, err error) {
	t.provider.Logger.Error("task error", "task_id", task.ID(), "error", err)
	// TODO: should this do anything else?
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

	err = t.tasks.AddWithID(link.SourceID, &tasks.Task{
		Interval:               timeDuration,
		FuncWithTaskContext:    t.TaskFunc,
		ErrFuncWithTaskContext: t.ErrorFunc,
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *Ticker) handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	t.provider.Logger.Info("handleDelTargetLink", "link", link)
	task, err := t.tasks.Lookup(link.SourceID)
	if err != nil || task == nil {
		return ErrTickerNotFound
	}

	t.tasks.Del(link.Name)
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
