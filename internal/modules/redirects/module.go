package redirects

import (
	"embed"
	"net/http"

	"github.com/enteresanlikk/go-toolset/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: Service{HTTP: d.NoFollow}}
	return platform.Module{
		Name:        "redirects",
		Title:       "Redirect Checker",
		Description: "Follow a URL's redirect chain hop by hop (max 10)",
		UI:          ui,
		Views: []platform.View{
			{Path: "/redirects", Dir: "trace", Title: "Trace"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/redirects", Example: "/api/redirects?url=github.com", Handler: h.Trace},
		},
	}
}

type Handler struct{ Svc Service }

func (h Handler) Trace(w http.ResponseWriter, r *http.Request) {
	tr, err := h.Svc.Trace(r.Context(), r.URL.Query().Get("url"))
	platform.Reply(w, tr, err)
}
