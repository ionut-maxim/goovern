package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"

	"github.com/ionut-maxim/goovern/config"
	"github.com/ionut-maxim/goovern/telemetry"
	"github.com/ionut-maxim/goovern/worker"
)

func main() {
	ctx := context.Background()
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config")
	}

	var logger *slog.Logger
	var otelShutdown telemetry.Shutdown

	if cfg.Telemetry.Enabled {
		logger, otelShutdown, err = telemetry.Setup(
			ctx,
			cfg.Telemetry.ServiceName,
			cfg.Telemetry.ServiceVersion,
			cfg.Telemetry.OTELEndpoint,
		)
		if err != nil {
			log.Fatalf("failed to setup telemetry: %v", err)
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := otelShutdown(shutdownCtx); err != nil {
				slog.Error("failed to shutdown telemetry", "error", err)
			}
		}()
		logger.Info("OpenTelemetry initialized", "endpoint", cfg.Telemetry.OTELEndpoint)
	} else {
		logger = cfg.Log.New()
	}

	pool, db, err := newDB(cfg.DB.Url, logger)
	if err != nil {
		slog.Error("failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	wrk, err := worker.New(pool, db, logger)
	if err != nil {
		slog.Error("failed to create worker", "error", err)
		os.Exit(1)
	}

	if err = wrk.Start(workerCtx); err != nil {
		slog.Error("failed to start worker", "error", err)
		os.Exit(1)
	}
	s := startSSHServer(pool, db, 42069, logger, done)

	<-done

	logger.Info("shutting down gracefully")

	workerCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer func() { cancel() }()

	logger.Info("Stopping worker - jobs will be cancelled in 3 seconds")
	if err = wrk.Stop(shutdownCtx); err != nil {
		logger.Error("Could not stop worker", "error", err)
	}
	logger.Info("Closing database connection")
	pool.Close()
	logger.Info("Stopping SSH server")
	if err = s.Shutdown(shutdownCtx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		logger.Error("Could not stop server", "error", err)
	}
}
