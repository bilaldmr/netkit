package dns

import (
	"context"
	"embed"
	"net"
	"net/http"

	"github.com/enteresanlikk/go-toolkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: Service{Resolver: d.Resolver}, Dialer: d.Dialer}
	return platform.Module{
		Name:        "dns",
		Title:       "DNS Checker",
		Description: "DNS records (A, AAAA, CNAME, MX, NS, TXT) propagation across public resolvers, and SPF/DMARC/DKIM checks",
		UI:          ui,
		Views: []platform.View{
			{Path: "/dns", Dir: "lookup", Title: "Lookup"},
			{Path: "/dns/propagation", Dir: "propagation", Title: "Propagation"},
			{Path: "/dns/email", Dir: "email", Title: "Email security"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/dns", Example: "/api/dns?domain=google.com&type=MX&server=1.1.1.1", Handler: h.Lookup},
			{Pattern: "GET /api/dns/propagation", Example: "/api/dns/propagation?domain=github.com&type=A", Handler: h.Propagation},
			{Pattern: "GET /api/dns/email", Example: "/api/dns/email?domain=github.com&selector=google", Handler: h.Email},
		},
	}
}

type Handler struct {
	Svc    Service
	Dialer *net.Dialer
}

func (h Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	svc, server := h.Svc, q.Get("server")
	if server != "" {
		server = serverAddr(server)
		svc.Resolver = serverResolver(h.Dialer, server)
	}
	recs, err := svc.Lookup(r.Context(), q.Get("domain"), q.Get("type"))
	platform.Reply(w, map[string]any{"domain": q.Get("domain"), "server": server, "records": recs}, err)
}

func (h Handler) Propagation(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p, err := Propagate(r.Context(), q.Get("domain"), q.Get("type"), PublicServers, func(s Server) Resolver {
		return serverResolver(h.Dialer, serverAddr(s.Addr))
	})
	platform.Reply(w, p, err)
}

func (h Handler) Email(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	e, err := h.Svc.Email(r.Context(), q.Get("domain"), q.Get("selector"))
	platform.Reply(w, e, err)
}

func serverAddr(s string) string {
	if _, _, err := net.SplitHostPort(s); err == nil {
		return s
	}
	return net.JoinHostPort(s, "53")
}

func serverResolver(d *net.Dialer, addr string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return d.DialContext(ctx, network, addr)
		},
	}
}
