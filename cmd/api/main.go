package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	taskRepo := postgresrepo.NewTaskRepository(pool)
	taskRecurrenceRepo := postgresrepo.NewTaskRecurrenceRepository(pool)
	taskRecurrenceDatesRepo := postgresrepo.NewTaskRecurrenceDateRepository(pool)
	taskOccurrenceRepo := postgresrepo.NewTaskOccurrenceRepository(pool)

	taskUsecase := task.NewService(
		taskRepo,
		taskRecurrenceRepo,
		taskRecurrenceDatesRepo,
	)

	taskOccurrenceUsecase := task.NewTaskOccurrenceService(
		taskOccurrenceRepo,
		taskRepo,
	)

	generatorService := task.NewGeneratorService(
		taskRepo,
		taskRecurrenceRepo,
		taskRecurrenceDatesRepo,
		taskOccurrenceRepo,
	)

	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	docsHandler := swaggerdocs.NewHandler()
	occurrenceHandler := httphandlers.NewTaskOccurrenceHandler(taskOccurrenceUsecase)
	router := transporthttp.NewRouter(taskHandler, occurrenceHandler, docsHandler)

	//first start
	go func() {
		if err := generatorService.GenerateRecent(ctx); err != nil {
			logger.Error("initial recent generation failed", "error", err)
		}
		runRecentGenerator(ctx, logger, generatorService)
	}()

	go runDailyGenerator(ctx, logger, generatorService)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8089"),
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func runRecentGenerator(ctx context.Context, logger *slog.Logger, generator *task.GeneratorService) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := generator.GenerateRecent(ctx); err != nil {
				logger.Error("generate recent occurrences", "error", err)
			} else {
				logger.Info("recent occurrence generation completed")
			}
		}
	}
}

func runDailyGenerator(ctx context.Context, logger *slog.Logger, generator *task.GeneratorService) {
	for {
		now := time.Now().UTC()
		nextRun := nextDailyRunUTC(now, 1, 0) // 01:00 UTC

		timer := time.NewTimer(time.Until(nextRun))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if err := generator.GenerateDaily(ctx); err != nil {
				logger.Error("generate daily occurrences", "error", err)
			} else {
				logger.Info("daily occurrence generation completed", "run_at", nextRun)
			}
		}
	}
}

func nextDailyRunUTC(now time.Time, hour, minute int) time.Time {
	run := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		hour,
		minute,
		0,
		0,
		time.UTC,
	)

	if !run.After(now) {
		run = run.Add(24 * time.Hour)
	}

	return run
}
