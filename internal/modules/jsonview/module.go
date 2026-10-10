package jsonview

import (
	"embed"
	"errors"
	"io"
	"net/http"

	"github.com/enteresanlikk/go-toolkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: Service{HTTP: d.HTTP}}
	return platform.Module{
		Name:        "jsonview",
		Title:       "JSON Viewer",
		Description: "Pretty-print JSON from a URL, ?data= or a POST body",
		UI:          ui,
		Views: []platform.View{
			{Path: "/json", Dir: "viewer", Title: "Viewer"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/json", Example: "/api/json?url=https://api.github.com", Handler: h.FromQuery},
			{Pattern: "POST /api/json", Handler: h.Format},
		},
	}
}

type Handler struct{ Svc Service }

func (h Handler) FromQuery(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	switch {
	case q.Get("data") != "":
		writeFormatted(w, []byte(q.Get("data")), nil)
	case q.Get("url") != "":
		raw, err := h.Svc.Fetch(r.Context(), q.Get("url"))
		writeFormatted(w, raw, err)
	default:
		platform.Reply(w, nil, errors.New("url or data param (or POST body) required"))
	}
}

func (h Handler) Format(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, MaxSize))
	writeFormatted(w, raw, err)
}

func writeFormatted(w http.ResponseWriter, raw []byte, err error) {
	if err == nil {
		raw, err = Format(raw)
	}
	if err != nil {
		platform.Reply(w, nil, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}
