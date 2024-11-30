package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSchedulerJob(t *testing.T) {

	t.Run("interval: valid config", func(t *testing.T) {
		cfg := map[string]string{
			"type":   "interval",
			"period": "10s",
		}

		job, err := newSchedulerJob(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, job)
	})

	t.Run("interval: invalid period", func(t *testing.T) {
		cfg := map[string]string{
			"type":   "interval",
			"period": "abcd",
		}

		job, err := newSchedulerJob(cfg)
		assert.Error(t, err)
		assert.Nil(t, job)
	})

	t.Run("interval: default type", func(t *testing.T) {
		cfg := map[string]string{
			"period": "10s",
		}

		job, err := newSchedulerJob(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, job)
	})

	t.Run("cron: valid config", func(t *testing.T) {
		cfg := map[string]string{
			"type": "cron",
			"cron": "* * * * *",
		}

		job, err := newSchedulerJob(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, job)
	})

	t.Run("cron: valid config, with seconds", func(t *testing.T) {
		cfg := map[string]string{
			"type":    "cron",
			"cron":    "* * * * * *",
			"seconds": "true",
		}

		job, err := newSchedulerJob(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, job)
	})

	t.Run("cron: missing cron", func(t *testing.T) {
		cfg := map[string]string{
			"type": "cron",
		}

		job, err := newSchedulerJob(cfg)
		assert.Error(t, err)
		assert.Nil(t, job)
	})
}
