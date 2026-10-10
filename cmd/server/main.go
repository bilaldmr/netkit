package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enteresanlikk/go-toolkit/internal/config"
	"github.com/enteresanlikk/go-toolkit/internal/logger"
	"github.com/enteresanlikk/go-toolkit/internal/modules/cidr"
	"github.com/enteresanlikk/go-toolkit/internal/modules/dns"
	"github.com/enteresanlikk/go-toolkit/internal/modules/encode"
	"github.com/enteresanlikk/go-toolkit/internal/modules/generate"
	"github.com/enteresanlikk/go-toolkit/internal/modules/headers"
	"github.com/enteresanlikk/go-toolkit/internal/modules/ipcheck"
	"github.com/enteresanlikk/go-toolkit/internal/modules/jsonview"
	"github.com/enteresanlikk/go-toolkit/internal/modules/myip"
	"github.com/enteresanlikk/go-toolkit/internal/modules/ports"
	"github.com/enteresanlikk/go-toolkit/internal/modules/redirects"
	"github.com/enteresanlikk/go-toolkit/internal/modules/sitefiles"
	"github.com/enteresanlikk/go-toolkit/internal/modules/ssl"
	"github.com/enteresanlikk/go-toolkit/internal/modules/whois"
	"github.com/enteresanlikk/go-toolkit/internal/platform"
	"github.com/enteresanlikk/go-toolkit/internal/server"
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
		cidr.New,
		dns.New,
		encode.New,
		generate.New,
		headers.New,
		ipcheck.New(cfg.IPCheck),
		jsonview.New,
		myip.New(cfg.MyIP),
		ports.New,
		redirects.New,
		sitefiles.New,
		ssl.New,
		whois.New,
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
