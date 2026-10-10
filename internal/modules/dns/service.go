package dns

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
)

type Resolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	LookupCNAME(ctx context.Context, host string) (string, error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
}

type Service struct{ Resolver Resolver }

func (s Service) Lookup(ctx context.Context, domain, typ string) (map[string]any, error) {
	if domain == "" {
		return nil, errors.New("domain required")
	}
	r := s.Resolver
	lookups := map[string]func() (any, error){
		"A":     func() (any, error) { return r.LookupIP(ctx, "ip4", domain) },
		"AAAA":  func() (any, error) { return r.LookupIP(ctx, "ip6", domain) },
		"CNAME": func() (any, error) { return r.LookupCNAME(ctx, domain) },
		"MX":    func() (any, error) { return r.LookupMX(ctx, domain) },
		"NS":    func() (any, error) { return r.LookupNS(ctx, domain) },
		"TXT":   func() (any, error) { return r.LookupTXT(ctx, domain) },
	}
	if typ = strings.ToUpper(typ); typ != "" {
		f, ok := lookups[typ]
		if !ok {
			return nil, errors.New("unsupported type " + strconv.Quote(typ))
		}
		lookups = map[string]func() (any, error){typ: f}
	}
	out := map[string]any{}
	for t, f := range lookups {
		if v, err := f(); err != nil {
			out[t] = map[string]string{"error": err.Error()}
		} else {
			out[t] = v
		}
	}
	return out, nil
}
