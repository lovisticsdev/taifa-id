package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-id/internal/config"
	"taifa-id/internal/membership"
	"taifa-id/internal/organization"
	"taifa-id/internal/person"
	"taifa-id/internal/platform/clock"
	"taifa-id/internal/platform/httpserver"
	"taifa-id/internal/platform/postgres"
)

type App struct {
	cfg        config.Config
	logger     *slog.Logger
	dbPool     *pgxpool.Pool
	httpServer *http.Server
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	if logger == nil {
		logger = slog.Default()
	}

	var dbPool *pgxpool.Pool
	if cfg.Database.DSN != "" {
		pool, err := postgres.Open(context.Background(), postgres.Config{
			DSN:            cfg.Database.DSN,
			MinConns:       cfg.Database.MinConns,
			MaxConns:       cfg.Database.MaxConns,
			ConnectTimeout: cfg.Database.ConnectTimeout,
		})
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}

		dbPool = pool
	}

	router := chi.NewRouter()

	router.Use(httpserver.CorrelationIDMiddleware)
	router.Use(httpserver.RequestLoggerMiddleware(logger))
	router.Use(httpserver.RecovererMiddleware(logger))

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpserver.WriteJSON(w, r, http.StatusOK, map[string]any{
			"status":      "ok",
			"service":     cfg.ServiceName,
			"environment": cfg.Environment,
		})
	})

	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusOK
		status := "ok"
		databaseStatus := "not_configured"

		if dbPool != nil {
			pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			if err := postgres.Ping(pingCtx, dbPool); err != nil {
				statusCode = http.StatusServiceUnavailable
				status = "degraded"
				databaseStatus = "unavailable"
			} else {
				databaseStatus = "ok"
			}
		}

		httpserver.WriteJSON(w, r, statusCode, map[string]any{
			"status":  status,
			"service": cfg.ServiceName,
			"dependencies": map[string]string{
				"database": databaseStatus,
			},
		})
	})

	if dbPool != nil {
		registerDomainRoutes(router, dbPool)
	}

	server := httpserver.New(httpserver.Config{
		Addr:         cfg.HTTP.Addr,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}, router)

	return &App{
		cfg:        cfg,
		logger:     logger,
		dbPool:     dbPool,
		httpServer: server,
	}, nil
}

func registerDomainRoutes(router chi.Router, dbPool *pgxpool.Pool) {
	realClock := clock.NewRealClock()

	personRepository := person.NewRepository(dbPool)
	personService := person.NewService(dbPool, personRepository, realClock)
	personHandler := person.NewHandler(personService)
	person.RegisterRoutes(router, personHandler)

	organizationRepository := organization.NewRepository(dbPool)
	organizationService := organization.NewService(dbPool, organizationRepository, realClock)
	organizationHandler := organization.NewHandler(organizationService)
	organization.RegisterRoutes(router, organizationHandler)

	membershipRepository := membership.NewRepository(dbPool)
	membershipService := membership.NewService(dbPool, membershipRepository, realClock)
	membershipHandler := membership.NewHandler(membershipService)
	membership.RegisterRoutes(router, membershipHandler)
}

func (a *App) Run(ctx context.Context) error {
	serverErrors := make(chan error, 1)

	go func() {
		a.logger.Info(
			"starting HTTP server",
			"service", a.cfg.ServiceName,
			"environment", a.cfg.Environment,
			"addr", a.cfg.HTTP.Addr,
		)

		err := a.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}

		serverErrors <- nil
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		if a.dbPool != nil {
			a.dbPool.Close()
			a.logger.Info("postgres pool closed")
		}

		a.logger.Info("HTTP server stopped cleanly")
		return nil

	case err := <-serverErrors:
		if a.dbPool != nil {
			a.dbPool.Close()
		}

		return err
	}
}
