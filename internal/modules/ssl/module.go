package ssl

import (
	"context"
	"crypto/tls"
	"embed"
	"net"
	"net/http"

	"github.com/bilaldmr/netkit/internal/platform"
)

type tlsProber struct{ dialer *net.Dialer }

func (p tlsProber) Probe(ctx context.Context, host, port string, cfg *tls.Config) (tls.ConnectionState, error) {
	cfg.ServerName, cfg.InsecureSkipVerify = host, true
	d := tls.Dialer{NetDialer: p.dialer, Config: cfg}
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
		Description: "TLS certificate check (validity, days left, chain) and supported TLS versions and ciphers (port default 443)",
		UI:          ui,
		Views: []platform.View{
			{Path: "/ssl", Dir: "check", Title: "Check"},
			{Path: "/ssl/ciphers", Dir: "ciphers", Title: "Ciphers"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/ssl", Example: "/api/ssl?host=google.com", Handler: h.Check},
			{Pattern: "GET /api/ssl/ciphers", Example: "/api/ssl/ciphers?host=google.com&compare=github.com", Handler: h.Ciphers},
		},
	}
}

type Handler struct{ Svc Service }

func (h Handler) Check(w http.ResponseWriter, r *http.Request) {
	rep, err := h.Svc.Check(r.Context(), r.URL.Query().Get("host"))
	platform.Reply(w, rep, err)
}

func (h Handler) Ciphers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	res, err := h.Svc.Ciphers(r.Context(), q.Get("host"), q.Get("compare"))
	platform.Reply(w, res, err)
}
