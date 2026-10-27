package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	config             *config.Config
	uiConfig           *server.UIConfig
	uiRouter           *router.UIRouter
	blogAPIRouter      *router.BlogAPIRouter
	auth               *auth.Auth
	core               *core.Core
	logger             *slog.Logger
	wg                 sync.WaitGroup
	db                 *sql.DB
	session            *databaseutils.SQLSession
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

	err = runMigrations(db, logger)
	if err != nil {
		logger.Error("Error running migrations", "error", err)
		os.Exit(1)
	}

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
	session := databaseutils.NewSession(db, 3*time.Second)
	core := core.NewCore(logger, session)
	userController := controller.NewUserController(
		core,
		logger, cfg)

	articleController := controller.NewArticleController(
		core, logger, cfg,
	)

	authInstance := auth.New(cfg)
	commentController := controller.NewCommentController(core, logger, cfg)
	tagController := controller.NewTagController(core, logger)
	profilerController := controller.NewProfileController(core, logger, authInstance)

	authUserMiddleware := middleware.NewAuthUserMiddleware(core, cfg, logger)

	app := &application{
		uiConfig: uiConfig,
		uiRouter: uiRouter,
		blogAPIRouter: router.NewBlogAPIRouter(userController, profilerController,
			articleController, commentController, tagController),
		auth:               authInstance,
		core:               core,
		logger:             logger,
		wg:                 sync.WaitGroup{},
		db:                 db,
		session:            session,
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

// runMigrations applies all database migrations to the specified database
func runMigrations(db *sql.DB, logger *slog.Logger) error {
	const migrationsPath = "/home/javad/Project/go-sample/golang-realworld-example-app/migrations"
	logger.Info("Running database migrations", "path", migrationsPath)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		(&url.URL{Scheme: "file", Path: migrationsPath}).String(),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	logger.Info("Migrations completed successfully")
	return nil
}
