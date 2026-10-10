package whois

import (
	"embed"
	"net/http"

	"github.com/enteresanlikk/netkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(d platform.Deps) platform.Module {
	h := Handler{Svc: &Service{HTTP: d.HTTP}}
	return platform.Module{
		Name:        "whois",
		Title:       "WHOIS / RDAP",
		Description: "Registration data for a domain, IP address or ASN from the official registries (RDAP)",
		UI:          ui,
		Views: []platform.View{
			{Path: "/whois", Dir: "lookup", Title: "Lookup"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/whois", Example: "/api/whois?q=github.com", Handler: h.Lookup},
		},
	}
}

type Handler struct{ Svc *Service }

func (h Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	sum, err := h.Svc.Lookup(r.Context(), r.URL.Query().Get("q"))
	platform.Reply(w, sum, err)
}
