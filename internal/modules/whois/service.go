package whois

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Summary struct {
	Query       string         `json:"query"`
	Kind        string         `json:"kind"`
	Server      string         `json:"server"`
	Name        string         `json:"name,omitempty"`
	Handle      string         `json:"handle,omitempty"`
	Status      []string       `json:"status,omitempty"`
	Registrar   string         `json:"registrar,omitempty"`
	Registrant  string         `json:"registrant,omitempty"`
	Registered  string         `json:"registered,omitempty"`
	Updated     string         `json:"updated,omitempty"`
	Expires     string         `json:"expires,omitempty"`
	DaysLeft    *int           `json:"days_left,omitempty"`
	Nameservers []string       `json:"nameservers,omitempty"`
	DNSSEC      *bool          `json:"dnssec,omitempty"`
	Range       string         `json:"range,omitempty"`
	Country     string         `json:"country,omitempty"`
	Raw         map[string]any `json:"raw"`
}

type Service struct {
	HTTP Doer

	mu     sync.Mutex
	boot   map[string][][2][]string
	loaded time.Time
}

const bootstrapURL = "https://data.iana.org/rdap/"

func (s *Service) Lookup(ctx context.Context, query string) (Summary, error) {
	q := strings.TrimSpace(strings.ToLower(query))
	q = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(q, "https://"), "http://"), "/")
	q, _, _ = strings.Cut(q, "/")
	if q == "" {
		return Summary{}, errors.New("query required: a domain, IP address or ASN")
	}

	var kind, path, base string
	var err error
	if ip, perr := netip.ParseAddr(q); perr == nil {
		ip = ip.Unmap()
		kind, path = "ip", "ip/"+ip.String()
		file := map[bool]string{true: "ipv4", false: "ipv6"}[ip.Is4()]
		base, err = s.server(ctx, file, func(key string) int {
			if p, err := netip.ParsePrefix(key); err == nil && p.Contains(ip) {
				return p.Bits()
			}
			return -1
		})
	} else if n, ok := asn(q); ok {
		kind, path = "autnum", "autnum/"+strconv.FormatUint(n, 10)
		base, err = s.server(ctx, "asn", func(key string) int {
			lo, hi, _ := strings.Cut(key, "-")
			a, _ := strconv.ParseUint(lo, 10, 32)
			b := a
			if hi != "" {
				b, _ = strconv.ParseUint(hi, 10, 32)
			}
			if n >= a && n <= b {
				return 1
			}
			return -1
		})
	} else {
		q = strings.TrimSuffix(q, ".")
		kind, path = "domain", "domain/"+q
		base, err = s.server(ctx, "dns", func(key string) int {
			if q == key || strings.HasSuffix(q, "."+key) {
				return strings.Count(key, ".") + 1
			}
			return -1
		})
	}
	if err != nil {
		return Summary{}, err
	}

	raw, err := s.fetch(ctx, base+path, "application/rdap+json")
	if err != nil {
		return Summary{}, err
	}
	sum := summarize(raw, time.Now())
	sum.Query, sum.Kind, sum.Server = q, kind, base
	return sum, nil
}

func asn(q string) (uint64, bool) {
	n, err := strconv.ParseUint(strings.TrimPrefix(q, "as"), 10, 32)
	return n, err == nil && strings.HasPrefix(q, "as")
}

func (s *Service) server(ctx context.Context, file string, score func(key string) int) (string, error) {
	services, err := s.bootstrap(ctx, file)
	if err != nil {
		return "", err
	}
	best, bestScore := "", -1
	for _, svc := range services {
		for _, key := range svc[0] {
			if sc := score(strings.ToLower(key)); sc > bestScore {
				best, bestScore = pickURL(svc[1]), sc
			}
		}
	}
	if best == "" {
		return "", errors.New("no RDAP server is registered for this query")
	}
	if !strings.HasSuffix(best, "/") {
		best += "/"
	}
	return best, nil
}

func pickURL(urls []string) string {
	for _, u := range urls {
		if strings.HasPrefix(u, "https://") {
			return u
		}
	}
	if len(urls) > 0 {
		return urls[0]
	}
	return ""
}

func (s *Service) bootstrap(ctx context.Context, file string) ([][2][]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.boot == nil || time.Since(s.loaded) > 24*time.Hour {
		s.boot, s.loaded = map[string][][2][]string{}, time.Now()
	}
	if b, ok := s.boot[file]; ok {
		return b, nil
	}
	raw, err := s.fetch(ctx, bootstrapURL+file+".json", "application/json")
	if err != nil {
		return nil, errors.New("loading IANA bootstrap: " + err.Error())
	}
	var parsed struct {
		Services [][2][]string `json:"services"`
	}
	b, _ := json.Marshal(raw)
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, errors.New("parsing IANA bootstrap: " + err.Error())
	}
	s.boot[file] = parsed.Services
	return parsed.Services, nil
}

func (s *Service) fetch(ctx context.Context, url, accept string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("not found in the registry")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("registry answered " + resp.Status)
	}
	var out map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&out); err != nil {
		return nil, errors.New("invalid RDAP response: " + err.Error())
	}
	return out, nil
}

func summarize(raw map[string]any, now time.Time) Summary {
	s := Summary{Raw: raw}
	s.Name = str(raw, "ldhName")
	if s.Name == "" {
		s.Name = str(raw, "name")
	}
	s.Handle = str(raw, "handle")
	s.Country = str(raw, "country")
	for _, v := range list(raw, "status") {
		if st, ok := v.(string); ok {
			s.Status = append(s.Status, st)
		}
	}
	for _, e := range objects(raw, "events") {
		date := str(e, "eventDate")
		switch str(e, "eventAction") {
		case "registration":
			s.Registered = date
		case "last changed":
			s.Updated = date
		case "expiration":
			s.Expires = date
			if t, err := time.Parse(time.RFC3339, date); err == nil {
				d := int(t.Sub(now).Hours() / 24)
				s.DaysLeft = &d
			}
		}
	}
	walkEntities(objects(raw, "entities"), func(roles []string, name string) {
		for _, r := range roles {
			switch {
			case r == "registrar" && s.Registrar == "":
				s.Registrar = name
			case r == "registrant" && s.Registrant == "":
				s.Registrant = name
			}
		}
	})
	for _, ns := range objects(raw, "nameservers") {
		s.Nameservers = append(s.Nameservers, strings.ToLower(str(ns, "ldhName")))
	}
	if sec, ok := raw["secureDNS"].(map[string]any); ok {
		if signed, ok := sec["delegationSigned"].(bool); ok {
			s.DNSSEC = &signed
		}
	}
	if a, b := str(raw, "startAddress"), str(raw, "endAddress"); a != "" {
		s.Range = a + " - " + b
	}
	if a, ok := raw["startAutnum"].(float64); ok {
		b, _ := raw["endAutnum"].(float64)
		s.Range = "AS" + strconv.Itoa(int(a)) + " - AS" + strconv.Itoa(int(b))
	}
	return s
}

func walkEntities(entities []map[string]any, fn func(roles []string, name string)) {
	for _, e := range entities {
		var roles []string
		for _, r := range list(e, "roles") {
			if s, ok := r.(string); ok {
				roles = append(roles, s)
			}
		}
		fn(roles, vcardName(e))
		walkEntities(objects(e, "entities"), fn)
	}
}

func vcardName(e map[string]any) string {
	vc := list(e, "vcardArray")
	if len(vc) < 2 {
		return ""
	}
	props, _ := vc[1].([]any)
	var fn string
	for _, p := range props {
		f, _ := p.([]any)
		if len(f) < 4 {
			continue
		}
		switch f[0] {
		case "org":
			if v, ok := f[3].(string); ok && v != "" {
				return v
			}
		case "fn":
			fn, _ = f[3].(string)
		}
	}
	return fn
}

func str(m map[string]any, k string) string {
	s, _ := m[k].(string)
	return s
}

func list(m map[string]any, k string) []any {
	l, _ := m[k].([]any)
	return l
}

func objects(m map[string]any, k string) []map[string]any {
	var out []map[string]any
	for _, v := range list(m, k) {
		if o, ok := v.(map[string]any); ok {
			out = append(out, o)
		}
	}
	return out
}
