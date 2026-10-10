package ssl

import (
	"context"
	"crypto/tls"
	"embed"
	"net"
	"net/http"

	"github.com/enteresanlikk/go-toolset/internal/platform"
)

type tlsProber struct{ dialer *net.Dialer }

func (p tlsProber) Probe(ctx context.Context, host, port string) (tls.ConnectionState, error) {
	d := tls.Dialer{NetDialer: p.dialer, Config: &tls.Config{ServerName: host, InsecureSkipVerify: true}}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return tls.ConnectionState{}, err
	}
	defer conn.Close()
	return conn.(*tls.Conn).ConnectionState(), nil
}

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: Service{Prober: tlsProber{d.Dialer}}}
	return platform.Module{
		Name:        "ssl",
		Title:       "SSL Checker",
		Description: "TLS certificate check: validity, days left, version, chain (port default 443)",
		UI:          ui,
		Views: []platform.View{
			{Path: "/ssl", Dir: "check", Title: "Check"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/ssl", Example: "/api/ssl?host=google.com", Handler: h.Check},
		},
	}
}

type Handler struct{ Svc Service }

func (h Handler) Check(w http.ResponseWriter, r *http.Request) {
	rep, err := h.Svc.Check(r.Context(), r.URL.Query().Get("host"))
	platform.Reply(w, rep, err)
}
