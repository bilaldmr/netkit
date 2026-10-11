package sitefiles

import (
	"embed"
	"net/http"

	"github.com/bilaldmr/netkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: Service{HTTP: d.HTTP}}
	return platform.Module{
		Name:        "sitefiles",
		Title:       "Site Files",
		Description: "robots.txt, sitemap.xml, security.txt and llms.txt of a site",
		UI:          ui,
		Views: []platform.View{
			{Path: "/site-files", Dir: "files", Title: "Files"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/site-files", Example: "/api/site-files?url=github.com", Handler: h.Get},
		},
	}
}

type Handler struct{ Svc Service }

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	files, err := h.Svc.Get(r.Context(), r.URL.Query().Get("url"))
	platform.Reply(w, map[string]any{"files": files}, err)
}
