package main

import (
	"errors"
	"log/slog"
	"testing"

	gocronmocks "github.com/go-co-op/gocron/mocks/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.wasmcloud.dev/provider"
)

func TestCreateTicker(t *testing.T) {
	ticker, err := CreateTicker()
	assert.NoError(t, err)
	assert.NotNil(t, ticker)
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := gocronmocks.NewMockScheduler(ctrl)

	ticker := Ticker{
		tasks: s,
	}
	s.EXPECT().Start().Times(1)

	ticker.Start()
}

func TestShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := gocronmocks.NewMockScheduler(ctrl)

	ticker := Ticker{
		tasks: s,
	}
	s.EXPECT().Shutdown().Times(1)

	ticker.Shutdown()
}

func TestPutTargetLink(t *testing.T) {
	t.Run("valid link", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		s := gocronmocks.NewMockScheduler(ctrl)
		j := gocronmocks.NewMockJob(ctrl)

		ticker := Ticker{
			tasks:    s,
			taskList: make(map[string]*TickerTask),
			provider: &provider.WasmcloudProvider{
				Logger: slog.Default(),
			},
		}
		mockId := uuid.New()

		s.EXPECT().NewJob(
			gomock.Any(),
			gomock.Any(),
		).Return(j, nil).Times(1)
		j.EXPECT().ID().Return(mockId).Times(1)

		testLink := provider.InterfaceLinkDefinition{
			Name:     "default",
			SourceID: "my-id",
			TargetConfig: map[string]string{
				"period": "10s",
			},
		}

		err := ticker.handlePutTargetLink(testLink)
		assert.NoError(t, err)

		myJob, ok := ticker.taskList["default.my-id"]
		assert.True(t, ok)
		assert.Equal(t, mockId, myJob.ID)
	})

	t.Run("invalid config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		s := gocronmocks.NewMockScheduler(ctrl)

		ticker := Ticker{
			tasks:    s,
			taskList: make(map[string]*TickerTask),
			provider: &provider.WasmcloudProvider{
				Logger: slog.Default(),
			},
		}

		testLink := provider.InterfaceLinkDefinition{
			Name:     "default",
			SourceID: "my-id",
			TargetConfig: map[string]string{
				"period": "abcd",
			},
		}

		err := ticker.handlePutTargetLink(testLink)
		assert.Error(t, err)

		_, ok := ticker.taskList["default.my-id"]
		assert.False(t, ok)
	})

	t.Run("new job error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		s := gocronmocks.NewMockScheduler(ctrl)

		ticker := Ticker{
			tasks:    s,
			taskList: make(map[string]*TickerTask),
			provider: &provider.WasmcloudProvider{
				Logger: slog.Default(),
			},
		}
		testError := errors.New("test error")

		s.EXPECT().NewJob(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, testError).Times(1)

		testLink := provider.InterfaceLinkDefinition{
			Name:     "default",
			SourceID: "my-id",
			TargetConfig: map[string]string{
				"period": "10s",
			},
		}

		err := ticker.handlePutTargetLink(testLink)
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		_, ok := ticker.taskList["default.my-id"]
		assert.False(t, ok)
	})
}

func TestDelTargetLink(t *testing.T) {
	t.Run("valid delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		s := gocronmocks.NewMockScheduler(ctrl)

		ticker := Ticker{
			tasks: s,
			taskList: map[string]*TickerTask{
				"default.my-id": {
					Component: "my-component",
					ID:        uuid.New(),
					Type:      "cron",
				},
			},
			provider: &provider.WasmcloudProvider{
				Logger: slog.Default(),
			},
		}

		s.EXPECT().RemoveJob(
			gomock.Any(),
		).Return(nil).Times(1)

		testLink := provider.InterfaceLinkDefinition{
			Name:     "default",
			SourceID: "my-id",
			TargetConfig: map[string]string{
				"period": "10s",
			},
		}

		err := ticker.handleDelTargetLink(testLink)
		assert.NoError(t, err)

		_, ok := ticker.taskList["default.my-id"]
		assert.False(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		s := gocronmocks.NewMockScheduler(ctrl)

		ticker := Ticker{
			tasks:    s,
			taskList: make(map[string]*TickerTask),
			provider: &provider.WasmcloudProvider{
				Logger: slog.Default(),
			},
		}

		testLink := provider.InterfaceLinkDefinition{
			Name:     "default",
			SourceID: "my-id",
			TargetConfig: map[string]string{
				"period": "10s",
			},
		}

		err := ticker.handleDelTargetLink(testLink)
		assert.Error(t, err)
		assert.Equal(t, ErrTickerNotFound, err)
	})

	t.Run("remove error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		s := gocronmocks.NewMockScheduler(ctrl)

		ticker := Ticker{
			tasks: s,
			taskList: map[string]*TickerTask{
				"default.my-id": {
					Component: "my-component",
					ID:        uuid.New(),
					Type:      "cron",
				},
			},
			provider: &provider.WasmcloudProvider{
				Logger: slog.Default(),
			},
		}
		testError := errors.New("test error")

		s.EXPECT().RemoveJob(
			gomock.Any(),
		).Return(testError).Times(1)

		testLink := provider.InterfaceLinkDefinition{
			Name:     "default",
			SourceID: "my-id",
			TargetConfig: map[string]string{
				"period": "10s",
			},
		}

		err := ticker.handleDelTargetLink(testLink)
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		_, ok := ticker.taskList["default.my-id"]
		assert.True(t, ok)
	})
}
