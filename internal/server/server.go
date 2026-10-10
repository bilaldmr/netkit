package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

type Config struct {
	Host string
	Port string
}

func (c *Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

type Server struct {
	l    *slog.Logger
	http *http.Server
}

func New(l *slog.Logger, config Config, handler http.Handler) *Server {
	return &Server{
		l: l,
		http: &http.Server{
			Addr:              config.Addr(),
			Handler:           handler,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,

			ErrorLog: slog.NewLogLogger(l.Handler(), slog.LevelError),
		},
	}
}

func (s *Server) ListenAndServe() error {
	s.l.Info("HTTP server started",
		slog.String("addr", s.http.Addr))

	if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
