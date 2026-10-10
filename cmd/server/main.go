package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enteresanlikk/netkit/internal/config"
	"github.com/enteresanlikk/netkit/internal/logger"
	"github.com/enteresanlikk/netkit/internal/modules/cidr"
	"github.com/enteresanlikk/netkit/internal/modules/dns"
	"github.com/enteresanlikk/netkit/internal/modules/encode"
	"github.com/enteresanlikk/netkit/internal/modules/generate"
	"github.com/enteresanlikk/netkit/internal/modules/headers"
	"github.com/enteresanlikk/netkit/internal/modules/ipcheck"
	"github.com/enteresanlikk/netkit/internal/modules/jsonview"
	"github.com/enteresanlikk/netkit/internal/modules/myip"
	"github.com/enteresanlikk/netkit/internal/modules/ports"
	"github.com/enteresanlikk/netkit/internal/modules/redirects"
	"github.com/enteresanlikk/netkit/internal/modules/sitefiles"
	"github.com/enteresanlikk/netkit/internal/modules/ssl"
	"github.com/enteresanlikk/netkit/internal/modules/whois"
	"github.com/enteresanlikk/netkit/internal/platform"
	"github.com/enteresanlikk/netkit/internal/server"
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
