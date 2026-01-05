package app

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService"
	"github.com/iwtcode/fanucService/internal/handlers"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/repository"
	"github.com/iwtcode/fanucService/internal/services/fanuc"
	"github.com/iwtcode/fanucService/internal/services/kafka"
	"github.com/iwtcode/fanucService/internal/usecases"
	"github.com/sirupsen/logrus"

	"go.uber.org/fx"
)

func New() *fx.App {
	return fx.New(
		fx.Provide(
			fanucService.LoadConfig,
			NewLogger,
			kafka.NewProducer,
			repository.NewRepository,
			fanuc.NewService,
			usecases.NewConnectionUsecase,
			usecases.NewRestoreUsecase,
			usecases.NewPollingUsecase,
			usecases.NewProgramUsecase,
			handlers.NewConnectionHandler,
			handlers.NewPollingHandler,
			handlers.NewProgramHandler,
			handlers.NewRouter,
		),
		fx.Invoke(
			startServer,
			restoreConnections,
			registerHooks,
		),
	)
}

func NewLogger(cfg *fanucService.Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if cfg.Logger.ServiceLevel == "off" || cfg.Logger.ServiceLevel == "none" {
		logger.SetOutput(io.Discard)
	} else {
		level, err := logrus.ParseLevel(cfg.Logger.ServiceLevel)
		if err != nil {
			level = logrus.InfoLevel
		}
		logger.SetLevel(level)
		logger.SetOutput(os.Stdout)
	}

	return logger
}

func registerHooks(lifecycle fx.Lifecycle, producer *kafka.Producer) {
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return producer.Close()
		},
	})
}

func restoreConnections(lifecycle fx.Lifecycle, usecase interfaces.RestoreUsecase) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			usecase.Restore()
			return nil
		},
	})
}

func startServer(lifecycle fx.Lifecycle, r *gin.Engine, cfg *fanucService.Config, logger *logrus.Logger) {
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Infof("Starting server on %s", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Errorf("Server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
