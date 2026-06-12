package main

import (
	"database/sql"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/golang-cz/devslog"
	_ "github.com/lib/pq"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/controller"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/middleware"
	"github.com/siahsang/blog/internal/router"
	"github.com/siahsang/blog/internal/server"
	"github.com/siahsang/blog/internal/utils/config"
	"github.com/siahsang/blog/internal/utils/databaseutils"
)

type application struct {
	config            *config.Config
	uiConfig          *server.UIConfig
	uiRouter          *router.UIRouter
	blogAPIRouter     *router.BlogAPIRouter
	auth              *auth.Auth
	core              *core.Core
	logger            *slog.Logger
	wg                sync.WaitGroup
	db                *sql.DB
	session           databaseutils.Session
	authUserMiddleware *middleware.AuthUserMiddleware
}

func main() {
	logger := configLogger()
	logger.Info("Starting application...")
	
	// Read database connection string from environment variable
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		// Fallback to default connection string
		dbDSN = "postgres://postgres:postgres@localhost/myblog?sslmode=disable"
		logger.Warn("DB_DSN environment variable not set, using default connection string")
	}
	
	db, err := databaseutils.OpenDBConnection(logger, dbDSN)
	if err != nil {
		logger.Error("Errors opening database connection", "error", err)
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Errors closing database connection", "error", err)
			os.Exit(1)
		}
	}()

	app, err := newApplication(db, logger)
	if err != nil {
		logger.Error("Error creating application", "error", err)
		os.Exit(1)
	}

	if err := app.serve(); err != nil {
		logger.Error("Error in starting server", "error", err)
		os.Exit(1)
	}
}

// newApplication creates and configures a new application instance
func newApplication(db *sql.DB, logger *slog.Logger) (*application, error) {
	cfg := &config.Config{}
	cfg.JWTSecret = os.Getenv("JWT_SECRET")

	uiConfig := &server.UIConfig{
		APIBaseURL: "/api",
	}

	uiRouter := router.NewUIRouter(logger)
	core := core.NewCore(db, logger, databaseutils.NewSQLTemplate(db, 3*time.Second))
	userController := controller.NewUserController(
		core,
		logger, cfg)

	authUserMiddleware := middleware.NewAuthUserMiddleware(core, cfg, logger)

	app := &application{
		uiConfig:           uiConfig,
		uiRouter:           uiRouter,
		blogAPIRouter:      router.NewBlogAPIRouter(userController),
		auth:               auth.New(cfg),
		core:               core,
		logger:             logger,
		wg:                 sync.WaitGroup{},
		db:                 db,
		session:            databaseutils.NewSession(db),
		config:             cfg,
		authUserMiddleware: authUserMiddleware,
	}

	return app, nil
}

func configLogger() *slog.Logger {
	handler := devslog.NewHandler(
		os.Stdout, &devslog.Options{
			HandlerOptions: &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
			NewLineAfterLog: false,
		})

	logger := slog.New(handler)
	return logger
}
