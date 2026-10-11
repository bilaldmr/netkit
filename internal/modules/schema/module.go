package schema

import (
	"embed"
	"net/http"

	"github.com/bilaldmr/netkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: &Service{HTTP: d.HTTP}}
	return platform.Module{
		Name:        "schema",
		Title:       "Structured Data",
		Description: "JSON-LD blocks of a page validated against the schema.org vocabulary",
		UI:          ui,
		Views: []platform.View{
			{Path: "/schema", Dir: "validate", Title: "Validate"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/schema", Example: "/api/schema?url=github.com", Handler: h.Get},
		},
	}
}

type Handler struct{ Svc *Service }

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	res, err := h.Svc.Get(r.Context(), r.URL.Query().Get("url"))
	platform.Reply(w, res, err)
}
