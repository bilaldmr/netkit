package myip

import (
	"embed"
	"net/http"
	"net/netip"
	"strings"

	"github.com/enteresanlikk/netkit/internal/platform"
)

//go:embed ui
var ui embed.FS

type Config struct {
	V4URL string
	V6URL string
}

func New(cfg Config) platform.Factory {
	return func(platform.Deps) platform.Module {
		h := Handler{Probes: map[string]string{
			"v4": strings.TrimRight(cfg.V4URL, "/"),
			"v6": strings.TrimRight(cfg.V6URL, "/"),
		}}
		return platform.Module{
			Name:        "myip",
			Title:       "My IP",
			Description: "Your IPv4 and IPv6 addresses and user agent",
			UI:          ui,
			Views: []platform.View{
				{Path: "/myip", Dir: "info", Title: "Info"},
				{Path: "/myip/headers", Dir: "headers", Title: "Request headers"},
			},
			Routes: []platform.Route{
				{Pattern: "GET /api/myip", Example: "/api/myip", Handler: Get},
				{Pattern: "GET /api/myip/probes", Example: "/api/myip/probes", Handler: h.ProbeURLs},
				{Pattern: "GET /api/myip/headers", Example: "/api/myip/headers", Handler: Headers},
			},
		}
	}
}

type Handler struct {
	Probes map[string]string
}

func Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ip, version := platform.ClientIP(r), 0
	if a, err := netip.ParseAddr(ip); err == nil {
		a = a.Unmap()
		ip, version = a.String(), 4
		if a.Is6() {
			version = 6
		}
	}
	platform.Reply(w, map[string]any{"ip": ip, "version": version, "user_agent": r.UserAgent()}, nil)
}

func Headers(w http.ResponseWriter, r *http.Request) {
	headers := map[string]string{}
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}
	platform.Reply(w, map[string]any{
		"client_ip":   platform.ClientIP(r),
		"remote_addr": r.RemoteAddr,
		"method":      r.Method,
		"proto":       r.Proto,
		"host":        r.Host,
		"headers":     headers,
	}, nil)
}

func (h Handler) ProbeURLs(w http.ResponseWriter, r *http.Request) {
	platform.Reply(w, h.Probes, nil)
}
