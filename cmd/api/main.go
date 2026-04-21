package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
		task.GeneratorConfig{
			RecentWindowMinutes: cfg.RecentWindowMinutes,
			RecentLookaheadDays: cfg.RecentLookaheadDays,
			DailyLookaheadDays:  cfg.DailyLookaheadDays,
		},
	)

	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	occurrenceHandler := httphandlers.NewTaskOccurrenceHandler(taskOccurrenceUsecase)
	docsHandler := swaggerdocs.NewHandler()
	router := transporthttp.NewRouter(taskHandler, occurrenceHandler, docsHandler)

	go func() {
		if err := generatorService.GenerateRecent(ctx); err != nil {
			logger.Error("initial recent generation failed", "error", err)
		}
		runRecentGenerator(ctx, logger, generatorService, cfg.RecentInterval)
	}()

	go runDailyGenerator(
		ctx,
		logger,
		generatorService,
		cfg.DailyRunHour,
		cfg.DailyRunMinute,
	)

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

	logger.Info(
		"http server started",
		"addr", cfg.HTTPAddr,
		"recent_interval", cfg.RecentInterval.String(),
		"recent_window_minutes", cfg.RecentWindowMinutes,
		"recent_lookahead_days", cfg.RecentLookaheadDays,
		"daily_lookahead_days", cfg.DailyLookaheadDays,
		"daily_run_hour", cfg.DailyRunHour,
		"daily_run_minute", cfg.DailyRunMinute,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type config struct {
	HTTPAddr            string
	DatabaseDSN         string
	RecentInterval      time.Duration
	RecentWindowMinutes int
	RecentLookaheadDays int
	DailyLookaheadDays  int
	DailyRunHour        int
	DailyRunMinute      int
}

func loadConfig() config {
	recentIntervalMinutes := envInt("GENERATOR_RECENT_INTERVAL_MINUTES", 5)

	cfg := config{
		HTTPAddr:            envOrDefault("HTTP_ADDR", ":8089"),
		DatabaseDSN:         envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
		RecentInterval:      time.Duration(recentIntervalMinutes) * time.Minute,
		RecentWindowMinutes: envInt("GENERATOR_RECENT_WINDOW_MINUTES", 10),
		RecentLookaheadDays: envInt("GENERATOR_RECENT_LOOKAHEAD_DAYS", 2),
		DailyLookaheadDays:  envInt("GENERATOR_DAILY_LOOKAHEAD_DAYS", 7),
		DailyRunHour:        envInt("GENERATOR_DAILY_RUN_HOUR", 1),
		DailyRunMinute:      envInt("GENERATOR_DAILY_RUN_MINUTE", 0),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	if cfg.DailyRunHour < 0 || cfg.DailyRunHour > 23 {
		panic(fmt.Errorf("GENERATOR_DAILY_RUN_HOUR must be between 0 and 23"))
	}

	if cfg.DailyRunMinute < 0 || cfg.DailyRunMinute > 59 {
		panic(fmt.Errorf("GENERATOR_DAILY_RUN_MINUTE must be between 0 and 59"))
	}

	if cfg.RecentWindowMinutes <= 0 {
		panic(fmt.Errorf("GENERATOR_RECENT_WINDOW_MINUTES must be positive"))
	}

	if cfg.RecentLookaheadDays < 0 {
		panic(fmt.Errorf("GENERATOR_RECENT_LOOKAHEAD_DAYS must be >= 0"))
	}

	if cfg.DailyLookaheadDays < 0 {
		panic(fmt.Errorf("GENERATOR_DAILY_LOOKAHEAD_DAYS must be >= 0"))
	}

	if cfg.RecentInterval <= 0 {
		panic(fmt.Errorf("GENERATOR_RECENT_INTERVAL_MINUTES must be positive"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Errorf("%s must be integer: %w", key, err))
	}

	return parsed
}

func runRecentGenerator(
	ctx context.Context,
	logger *slog.Logger,
	generator *task.GeneratorService,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
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

func runDailyGenerator(
	ctx context.Context,
	logger *slog.Logger,
	generator *task.GeneratorService,
	hour int,
	minute int,
) {
	for {
		now := time.Now().UTC()
		nextRun := nextDailyRunUTC(now, hour, minute)

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
