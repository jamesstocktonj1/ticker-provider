package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.wasmcloud.dev/provider"
	wrpcnats "wrpc.io/go/nats"
)

const (
	// Config Type
	configTypeKey     = "type"
	configTypeDefault = configTypeInterval

	configTypeInterval = "interval"
	configTypeCron     = "cron"
	configTypeStartup  = "startup"

	// Interval Config
	intervalConfigKey = "period"

	// Cron Config
	cronConfigKey    = "cron"
	cronSecConfigKey = "seconds"

	// Start Up Config
	delayConfigKey = "delay"
)

var (
	ErrInvalidJobType = errors.New("invalid config \"type\" specified")

	ErrMissingConfigValue = errors.New("missing config value")
)

func getJobKey(link provider.InterfaceLinkDefinition) string {
	return fmt.Sprintf("%s.%s", link.Name, link.SourceID)
}

func newSchedulerJob(config map[string]string) (gocron.JobDefinition, error) {
	if _, ok := config[configTypeKey]; !ok {
		config[configTypeKey] = configTypeDefault
	}

	switch config[configTypeKey] {
	case configTypeInterval:
		return newIntervalJob(config)
	case configTypeCron:
		return newCronJob(config)
	case configTypeStartup:
		return newStartupJob(config)
	default:
		return nil, ErrInvalidJobType
	}
}

func newIntervalJob(config map[string]string) (gocron.JobDefinition, error) {
	timeIntervalConfig, ok := config[intervalConfigKey]
	if !ok {
		return nil, fmt.Errorf("%w: key %s", ErrMissingConfigValue, intervalConfigKey)
	}

	timeInterval, err := time.ParseDuration(timeIntervalConfig)
	if err != nil {
		return nil, err
	}

	return gocron.DurationJob(timeInterval), nil
}

func newCronJob(config map[string]string) (gocron.JobDefinition, error) {
	cronConfig, ok := config[cronConfigKey]
	if !ok {
		return nil, fmt.Errorf("%w: key %s", ErrMissingConfigValue, cronConfigKey)
	}

	cronSeconds := false
	if secConfig, ok := config[cronSecConfigKey]; ok {
		cronSeconds = strings.Compare(secConfig, "true") == 0
	}

	return gocron.CronJob(cronConfig, cronSeconds), nil
}

func newStartupJob(config map[string]string) (gocron.JobDefinition, error) {
	delayConfig, ok := config[delayConfigKey]
	if !ok {
		return nil, fmt.Errorf("%w: key %s", ErrMissingConfigValue, delayConfigKey)
	}

	timeDelay, err := time.ParseDuration(delayConfig)
	if err != nil {
		return nil, err
	}

	return gocron.OneTimeJob(
		gocron.OneTimeJobStartDateTime(time.Now().Add(timeDelay)),
	), nil
}

func injectTraceHeader(_ctx context.Context) context.Context {
	carrier := nats.Header{}
	otel.GetTextMapPropagator().Inject(_ctx, NatsHeaderCarrier(carrier))
	return wrpcnats.ContextWithHeader(_ctx, carrier)
}
