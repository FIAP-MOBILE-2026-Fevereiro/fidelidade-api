package app

import (
	"context"
	"fmt"

	"github.com/felip/api-fidelidade/internal/auth"
	"github.com/felip/api-fidelidade/internal/config"
	"github.com/felip/api-fidelidade/internal/http/handler"
	"github.com/felip/api-fidelidade/internal/http/router"
	"github.com/felip/api-fidelidade/internal/platform/migrations"
	"github.com/felip/api-fidelidade/internal/platform/postgres"
	"github.com/felip/api-fidelidade/internal/service"
	"github.com/felip/api-fidelidade/internal/storage"
	"github.com/felip/api-fidelidade/internal/store"
	"github.com/gin-gonic/gin"
)

type App struct {
	config config.Config
	router *gin.Engine
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if cfg.AutoMigrate {
		if err := migrations.Up(cfg.DatabaseURL, cfg.MigrationsDir); err != nil {
			return nil, err
		}
	}

	pool, err := postgres.OpenPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	appStore := store.New(pool)
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.TokenTTL)
	fileStorage, err := storage.NewLocal(cfg.UploadDir, cfg.PublicBaseURL)
	if err != nil {
		return nil, err
	}

	userService := service.NewUserService(appStore.Queries, tokenManager, fileStorage)
	programService := service.NewProgramService(appStore.Queries)
	stampService := service.NewStampService(appStore)

	routerEngine := router.New(router.Dependencies{
		Config:         cfg,
		TokenManager:   tokenManager,
		UserHandler:    handler.NewUserHandler(userService),
		ProgramHandler: handler.NewProgramHandler(programService),
		StampHandler:   handler.NewStampHandler(stampService),
		QRHandler:      handler.NewQRHandler(stampService),
	})

	return &App{
		config: cfg,
		router: routerEngine,
	}, nil
}

func (a *App) Run() error {
	return a.router.Run(fmt.Sprintf(":%d", a.config.HTTPPort))
}
