package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService"
	"github.com/iwtcode/fanucService/internal/handlers"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/repository"
	"github.com/iwtcode/fanucService/internal/services/fanuc"
	"github.com/iwtcode/fanucService/internal/usecases"

	"go.uber.org/fx"
)

func New() *fx.App {
	return fx.New(
		fx.Provide(
			fanucService.LoadConfig,
			repository.NewRepository,
			fanuc.NewService,
			usecases.NewConnectionUsecase,
			handlers.NewConnectionHandler,
			handlers.NewRouter,
		),
		fx.Invoke(
			startServer,
			restoreConnections,
		),
	)
}

func restoreConnections(lifecycle fx.Lifecycle, service interfaces.FanucService) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go service.RestoreConnections()
			return nil
		},
	})
}

func startServer(lifecycle fx.Lifecycle, r *gin.Engine, cfg *fanucService.Config) {
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				fmt.Printf("Starting server on %s\n", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Printf("Server error: %v\n", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
