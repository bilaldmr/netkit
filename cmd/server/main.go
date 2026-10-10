package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enteresanlikk/go-toolset/internal/config"
	"github.com/enteresanlikk/go-toolset/internal/logger"
	"github.com/enteresanlikk/go-toolset/internal/modules/dns"
	"github.com/enteresanlikk/go-toolset/internal/modules/ipcheck"
	"github.com/enteresanlikk/go-toolset/internal/modules/jsonview"
	"github.com/enteresanlikk/go-toolset/internal/modules/myip"
	"github.com/enteresanlikk/go-toolset/internal/modules/redirects"
	"github.com/enteresanlikk/go-toolset/internal/modules/ssl"
	"github.com/enteresanlikk/go-toolset/internal/platform"
	"github.com/enteresanlikk/go-toolset/internal/server"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.Load()
	l := logger.New(cfg.Log)

	if err := run(l, cfg); err != nil {
		l.Error("startup failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(l *slog.Logger, cfg config.Config) error {
	routes := platform.NewServer(platform.NewDeps(),
		dns.New,
		ipcheck.New(cfg.IPCheck),
		jsonview.New,
		myip.New(cfg.MyIP),
		redirects.New,
		ssl.New,
	)
	httpServer := server.New(l, cfg.HTTPServer, routes)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() { serverErr <- httpServer.ListenAndServe() }()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		stop()
	}

	l.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		l.Error("Graceful shutdown failed",
			slog.String("error", err.Error()))
		return err
	}

	l.Info("Shutdown complete")
	return nil
}
