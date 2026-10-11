package ports

import (
	"embed"
	"net/http"

	"github.com/bilaldmr/netkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: Service{Resolver: d.Resolver, Dialer: d.Dialer}}
	return platform.Module{
		Name:        "ports",
		Title:       "Port Checker",
		Description: "Check which TCP ports of a public host are open (up to 20 per check)",
		UI:          ui,
		Views: []platform.View{
			{Path: "/ports", Dir: "check", Title: "Check"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/ports", Example: "/api/ports?host=github.com&ports=22,80,443", Handler: h.Check},
		},
	}
}

type Handler struct{ Svc Service }

func (h Handler) Check(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rep, err := h.Svc.Check(r.Context(), q.Get("host"), q.Get("ports"))
	platform.Reply(w, rep, err)
}
