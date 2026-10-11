package headers

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
		Name:        "headers",
		Title:       "HTTP Headers",
		Description: "Response headers of a URL, grouped into security, permissions, cache and CORS",
		UI:          ui,
		Views: []platform.View{
			{Path: "/headers", Dir: "inspect", Title: "Inspect"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/headers", Example: "/api/headers?url=github.com", Handler: h.Get},
		},
	}
}

type Handler struct{ Svc Service }

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	res, err := h.Svc.Get(r.Context(), r.URL.Query().Get("url"))
	platform.Reply(w, res, err)
}
