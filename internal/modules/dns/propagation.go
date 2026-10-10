package dns

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	Addr    string `json:"addr"`
}

var PublicServers = []Server{
	{"Google", "US", "8.8.8.8"},
	{"Cloudflare", "US", "1.1.1.1"},
	{"Quad9", "CH", "9.9.9.9"},
	{"OpenDNS", "US", "208.67.222.222"},
	{"Lumen (Level3)", "US", "4.2.2.2"},
	{"Comodo", "US", "8.26.56.26"},
	{"UltraDNS", "US", "156.154.70.1"},
	{"Control D", "CA", "76.76.2.0"},
	{"AdGuard", "CY", "94.140.14.14"},
	{"DNS.SB", "DE", "185.222.222.222"},
	{"Yandex", "RU", "77.88.8.8"},
	{"AliDNS", "CN", "223.5.5.5"},
	{"114DNS", "CN", "114.114.114.114"},
	{"KT", "KR", "168.126.63.1"},
}

type Answer struct {
	Server  Server   `json:"server"`
	Records []string `json:"records,omitempty"`
	Error   string   `json:"error,omitempty"`
	Agrees  bool     `json:"agrees"`
}

type Propagation struct {
	Domain     string   `json:"domain"`
	Type       string   `json:"type"`
	Majority   []string `json:"majority"`
	Consistent bool     `json:"consistent"`
	Answers    []Answer `json:"answers"`
}

func Propagate(ctx context.Context, domain, typ string, servers []Server, resolverFor func(Server) Resolver) (Propagation, error) {
	if domain == "" {
		return Propagation{}, errors.New("domain required")
	}
	if typ = strings.ToUpper(typ); typ == "" {
		typ = "A"
	}
	if !slices.Contains(Types, typ) {
		return Propagation{}, errors.New("unsupported type " + strconv.Quote(typ))
	}

	answers := make([]Answer, len(servers))
	var wg sync.WaitGroup
	for i, s := range servers {
		wg.Go(func() {
			c, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			recs, err := records(c, resolverFor(s), domain, typ)
			answers[i] = Answer{Server: s, Records: recs}
			if err != nil {
				answers[i].Error = err.Error()
			}
		})
	}
	wg.Wait()

	count := map[string]int{}
	best := ""
	for _, a := range answers {
		if a.Error != "" {
			continue
		}
		k := strings.Join(a.Records, "\n")
		if count[k]++; count[k] > count[best] {
			best = k
		}
	}
	p := Propagation{Domain: domain, Type: typ, Answers: answers, Consistent: true}
	if count[best] > 0 {
		p.Majority = strings.Split(best, "\n")
	}
	for i, a := range answers {
		answers[i].Agrees = a.Error == "" && count[best] > 0 && strings.Join(a.Records, "\n") == best
		p.Consistent = p.Consistent && answers[i].Agrees
	}
	return p, nil
}

var Types = []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT"}

func records(ctx context.Context, r Resolver, domain, typ string) ([]string, error) {
	var out []string
	var err error
	switch typ {
	case "A", "AAAA":
		network := map[string]string{"A": "ip4", "AAAA": "ip6"}[typ]
		ips, e := r.LookupIP(ctx, network, domain)
		for _, ip := range ips {
			out = append(out, ip.String())
		}
		err = e
	case "CNAME":
		var c string
		c, err = r.LookupCNAME(ctx, domain)
		out = []string{c}
	case "MX":
		mx, e := r.LookupMX(ctx, domain)
		for _, m := range mx {
			out = append(out, strconv.Itoa(int(m.Pref))+" "+m.Host)
		}
		err = e
	case "NS":
		ns, e := r.LookupNS(ctx, domain)
		for _, n := range ns {
			out = append(out, n.Host)
		}
		err = e
	case "TXT":
		out, err = r.LookupTXT(ctx, domain)
	default:
		return nil, errors.New("unsupported type " + strconv.Quote(typ))
	}
	if err != nil {
		return nil, err
	}
	slices.Sort(out)
	return out, nil
}
