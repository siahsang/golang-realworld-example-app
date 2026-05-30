package test_utils

import (
	"database/sql"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/controller"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/router"
	"github.com/siahsang/blog/internal/server"
	"github.com/siahsang/blog/internal/utils/config"
	"github.com/siahsang/blog/internal/utils/databaseutils"
)

// TestApplication holds a configured application instance for testing
type TestApplication struct {
	Config        *config.Config
	UIConfig      *server.UIConfig
	UIRouter      *router.UIRouter
	BlogAPIRouter *router.BlogAPIRouter
	Auth          *auth.Auth
	Core          *core.Core
	Logger        *slog.Logger
	WG            sync.WaitGroup
	DB            *sql.DB
	Session       databaseutils.Session
}

// NewTestApplication creates a new application instance configured for testing
func NewTestApplication(db *sql.DB, logger *slog.Logger) (*TestApplication, error) {
	cfg := &config.Config{}
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "test-secret-key"
	}

	uiConfig := &server.UIConfig{
		APIBaseURL: "/api",
	}

	uiRouter := router.NewUIRouter(logger)
	coreInstance := core.NewCore(db, logger, databaseutils.NewSQLTemplate(db, 3*time.Second))
	userController := controller.NewUserController(
		coreInstance,
		logger, cfg)

	app := &TestApplication{
		Config:        cfg,
		UIConfig:      uiConfig,
		UIRouter:      uiRouter,
		BlogAPIRouter: router.NewBlogAPIRouter(userController),
		Auth:          auth.New(cfg),
		Core:          coreInstance,
		Logger:        logger,
		WG:            sync.WaitGroup{},
		DB:            db,
		Session:       databaseutils.NewSession(db),
	}

	return app, nil
}

// GetTestDB returns the test database connection
func GetTestDB() *sql.DB {
	return testDB
}
