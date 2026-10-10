package ipcheck

import (
	"embed"
	"net/http"

	"github.com/enteresanlikk/go-toolset/internal/platform"
)

//go:embed ui
var ui embed.FS

type Config struct {
	IPInfoToken string
}

func New(cfg Config) platform.Factory {
	return func(d platform.Deps) platform.Module {
		h := Handler{Svc: Service{Resolver: d.Resolver, Locator: IPInfo{HTTP: d.HTTP, Token: cfg.IPInfoToken}}}
		return platform.Module{
			Name:        "ipcheck",
			Title:       "IP Checker",
			Description: "Classify an IP (version, public/private, ...), reverse DNS and location; defaults to caller's IP",
			UI:          ui,
			Views: []platform.View{
				{Path: "/ip", Dir: "check", Title: "Check"},
			},
			Routes: []platform.Route{
				{Pattern: "GET /api/ip", Example: "/api/ip?ip=8.8.8.8", Handler: h.Check},
			},
		}
	}
}

type Handler struct{ Svc Service }

func (h Handler) Check(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		ip = platform.ClientIP(r)
	}
	rep, err := h.Svc.Check(r.Context(), ip)
	platform.Reply(w, rep, err)
}
