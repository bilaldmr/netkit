package cidr

import (
	"embed"
	"net/http"

	"github.com/enteresanlikk/netkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(_ platform.Deps) platform.Module {
	return platform.Module{
		Name:        "cidr",
		Title:       "CIDR Calculator",
		Description: "Subnet range, mask, host count and whether an IP is inside it (IPv4 and IPv6)",
		UI:          ui,
		Views: []platform.View{
			{Path: "/cidr", Dir: "calc", Title: "Calculate"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/cidr", Example: "/api/cidr?cidr=192.168.1.0/24&ip=192.168.1.10", Handler: Get},
		},
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	info, err := Calculate(q.Get("cidr"), q.Get("ip"))
	platform.Reply(w, info, err)
}
