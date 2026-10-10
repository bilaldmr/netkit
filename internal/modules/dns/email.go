package dns

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

var DKIMSelectors = []string{"default", "google", "selector1", "selector2", "k1", "s1", "s2", "dkim", "mail"}

type Policy struct {
	Record string            `json:"record,omitempty"`
	Tags   map[string]string `json:"tags,omitempty"`
	Issues []string          `json:"issues,omitempty"`
}

type DKIMKey struct {
	Selector string `json:"selector"`
	Record   string `json:"record"`
}

type Email struct {
	Domain       string    `json:"domain"`
	SPF          Policy    `json:"spf"`
	DMARC        Policy    `json:"dmarc"`
	DKIM         []DKIMKey `json:"dkim"`
	DKIMSelector []string  `json:"dkim_checked"`
}

func (s Service) Email(ctx context.Context, domain, selector string) (Email, error) {
	domain = strings.TrimSuffix(strings.TrimSpace(domain), ".")
	if domain == "" {
		return Email{}, errors.New("domain required")
	}
	e := Email{Domain: domain, DKIM: []DKIMKey{}, DKIMSelector: DKIMSelectors}
	if selector != "" {
		e.DKIMSelector = []string{selector}
	}

	spf := s.txt(ctx, domain, "v=spf1")
	switch len(spf) {
	case 0:
		e.SPF.Issues = []string{"no SPF record"}
	case 1:
		e.SPF = checkSPF(spf[0])
	default:
		e.SPF = Policy{Record: strings.Join(spf, "\n"), Issues: []string{"multiple SPF records; receivers treat this as an error"}}
	}

	if dmarc := s.txt(ctx, "_dmarc."+domain, "v=DMARC1"); len(dmarc) == 0 {
		e.DMARC.Issues = []string{"no DMARC record"}
	} else {
		e.DMARC = checkDMARC(dmarc[0])
	}

	for _, sel := range e.DKIMSelector {
		for _, rec := range s.txt(ctx, sel+"._domainkey."+domain, "") {
			if strings.Contains(rec, "p=") {
				e.DKIM = append(e.DKIM, DKIMKey{Selector: sel, Record: rec})
			}
		}
	}
	return e, nil
}

func (s Service) txt(ctx context.Context, name, prefix string) []string {
	recs, _ := s.Resolver.LookupTXT(ctx, name)
	var out []string
	for _, r := range recs {
		if len(r) >= len(prefix) && strings.EqualFold(r[:len(prefix)], prefix) {
			out = append(out, r)
		}
	}
	return out
}

func checkSPF(rec string) Policy {
	p := Policy{Record: rec}
	lookups, all, redirect := 0, "", false
	for _, term := range strings.Fields(rec)[1:] {
		mech := strings.ToLower(strings.TrimLeft(term, "+-~?"))
		name, _, _ := strings.Cut(mech, ":")
		name, _, _ = strings.Cut(name, "/")
		switch {
		case name == "all":
			all = term
		case strings.HasPrefix(mech, "redirect="):
			redirect = true
			lookups++
		case name == "include" || name == "a" || name == "mx" || name == "ptr" || name == "exists":
			lookups++
		}
	}
	switch strings.ToLower(all) {
	case "all", "+all":
		p.Issues = append(p.Issues, "+all lets any server send mail as this domain")
	case "?all":
		p.Issues = append(p.Issues, "?all is neutral; use ~all or -all")
	case "":
		if !redirect {
			p.Issues = append(p.Issues, "no all mechanism; add ~all or -all")
		}
	}
	if lookups > 10 {
		p.Issues = append(p.Issues, strconv.Itoa(lookups)+" DNS lookups; SPF allows at most 10")
	}
	return p
}

func checkDMARC(rec string) Policy {
	p := Policy{Record: rec, Tags: map[string]string{}}
	for _, part := range strings.Split(rec, ";") {
		if k, v, ok := strings.Cut(strings.TrimSpace(part), "="); ok {
			p.Tags[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
		}
	}
	switch strings.ToLower(p.Tags["p"]) {
	case "reject", "quarantine":
	case "none":
		p.Issues = append(p.Issues, "p=none only monitors; failing mail is still delivered")
	case "":
		p.Issues = append(p.Issues, "missing p= policy tag")
	default:
		p.Issues = append(p.Issues, "invalid policy p="+p.Tags["p"])
	}
	if p.Tags["rua"] == "" {
		p.Issues = append(p.Issues, "no rua= address; you will not receive aggregate reports")
	}
	return p
}
