package ipcheck

import (
	"context"
	"net/netip"

	"github.com/enteresanlikk/go-toolset/internal/platform"
)

type ReverseResolver interface {
	LookupAddr(ctx context.Context, addr string) ([]string, error)
}

type Locator interface {
	Locate(ctx context.Context, ip string) (Location, error)
}

type Location struct {
	City     string  `json:"city,omitempty"`
	Region   string  `json:"region,omitempty"`
	Country  string  `json:"country,omitempty"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Org      string  `json:"org,omitempty"`
	Timezone string  `json:"timezone,omitempty"`
}

type Report struct {
	IP            string    `json:"ip"`
	Version       int       `json:"version"`
	Public        bool      `json:"public"`
	Private       bool      `json:"private"`
	Loopback      bool      `json:"loopback"`
	Multicast     bool      `json:"multicast"`
	LinkLocal     bool      `json:"link_local"`
	PTR           []string  `json:"ptr"`
	Location      *Location `json:"location,omitempty"`
	LocationError string    `json:"location_error,omitempty"`
}

type Service struct {
	Resolver ReverseResolver
	Locator  Locator
}

func (s Service) Check(ctx context.Context, raw string) (Report, error) {
	ip, err := netip.ParseAddr(raw)
	if err != nil {
		return Report{}, err
	}
	ip = ip.Unmap()
	ptr, _ := s.Resolver.LookupAddr(ctx, ip.String())
	version := 4
	if ip.Is6() {
		version = 6
	}
	rep := Report{
		IP:        ip.String(),
		Version:   version,
		Public:    platform.CheckPublic(ip.String()) == nil,
		Private:   ip.IsPrivate(),
		Loopback:  ip.IsLoopback(),
		Multicast: ip.IsMulticast(),
		LinkLocal: ip.IsLinkLocalUnicast(),
		PTR:       ptr,
	}

	if rep.Public {
		if loc, err := s.Locator.Locate(ctx, rep.IP); err != nil {
			rep.LocationError = err.Error()
		} else {
			rep.Location = &loc
		}
	}
	return rep, nil
}
