package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

const (
	VocabURL = "https://schema.org/version/latest/schemaorg-current-https.jsonld"
	MaxBody  = 5 << 20
)

var ldScript = regexp.MustCompile(`(?is)<script[^>]*type\s*=\s*["']?application/ld\+json["']?[^>]*>(.*?)</script>`)

type Issue struct {
	Level  string `json:"level"` // error or warning
	Path   string `json:"path"`
	Msg    string `json:"msg"`
	Source string `json:"source"` // schema.org or google
}

type Block struct {
	Script  int      `json:"script"`             // 1-based index of the ld+json script tag
	Path    string   `json:"path"`               // location inside the script, $.@graph[2] for graph entries
	InGraph bool     `json:"in_graph,omitempty"` // block is an entry of the script's @graph
	Types   []string `json:"types"`
	Data    any      `json:"data,omitempty"`
	Raw     string   `json:"raw,omitempty"`
	Issues  []Issue  `json:"issues"`
}

type Result struct {
	Final  string  `json:"final"`
	Status int     `json:"status"`
	Blocks []Block `json:"blocks"`
}

type Service struct {
	HTTP Doer

	mu    sync.Mutex
	vocab *Vocab
}

func (s *Service) Get(ctx context.Context, raw string) (Result, error) {
	if raw == "" {
		return Result{}, errors.New("url required")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return Result{}, errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Result{}, errors.New("only http and https urls are supported")
	}
	v, err := s.loadVocab(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("schema.org vocabulary: %w", err)
	}
	body, resp, err := s.fetch(ctx, u.String())
	if err != nil {
		return Result{}, err
	}
	res := Result{Final: resp.Request.URL.String(), Status: resp.StatusCode, Blocks: []Block{}}
	for i, m := range ldScript.FindAllSubmatch(body, -1) {
		for _, b := range Split(v.Check(m[1])) {
			b.Script = i + 1
			res.Blocks = append(res.Blocks, b)
		}
	}
	return res, nil
}

// ponytail: vocab is fetched once per process, restart to pick up a new schema.org release.
func (s *Service) loadVocab(ctx context.Context) (*Vocab, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vocab != nil {
		return s.vocab, nil
	}
	body, resp, err := s.fetch(ctx, VocabURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	v, err := ParseVocab(body)
	if err != nil {
		return nil, err
	}
	s.vocab = v
	return v, nil
}

func (s *Service) fetch(ctx context.Context, target string) ([]byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBody))
	return body, resp, err
}

type Vocab struct {
	Classes map[string][]string // class -> parents
	Props   map[string]Prop
	Members map[string]bool // enumeration members like InStock
	lower   map[string]string
}

type Prop struct {
	Domains      []string
	SupersededBy string
}

type ids []struct {
	ID string `json:"@id"`
}

// UnmarshalJSON accepts both a single {"@id"} object and an array of them.
func (x *ids) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '{' {
		b = append(append([]byte{'['}, b...), ']')
	}
	type plain ids
	return json.Unmarshal(b, (*plain)(x))
}

func (x ids) names() []string {
	var out []string
	for _, v := range x {
		out = append(out, term(v.ID))
	}
	return out
}

func ParseVocab(b []byte) (*Vocab, error) {
	var doc struct {
		Graph []struct {
			ID           string          `json:"@id"`
			Type         json.RawMessage `json:"@type"`
			SubClassOf   ids             `json:"rdfs:subClassOf"`
			Domain       ids             `json:"schema:domainIncludes"`
			SupersededBy ids             `json:"schema:supersededBy"`
		} `json:"@graph"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	v := &Vocab{Classes: map[string][]string{}, Props: map[string]Prop{}, Members: map[string]bool{}}
	for _, n := range doc.Graph {
		name := term(n.ID)
		types := strings.Join(strTypes(n.Type), " ")
		switch {
		case strings.Contains(types, "rdf:Property"):
			p := Prop{Domains: n.Domain.names()}
			if s := n.SupersededBy.names(); len(s) > 0 {
				p.SupersededBy = s[0]
			}
			v.Props[name] = p
		case strings.Contains(types, "rdfs:Class"):
			v.Classes[name] = n.SubClassOf.names()
		default:
			v.Members[name] = true
		}
	}
	v.lower = make(map[string]string, len(v.Props))
	for name := range v.Props {
		v.lower[strings.ToLower(name)] = name
	}
	if len(v.Classes) == 0 || len(v.Props) == 0 {
		return nil, errors.New("empty vocabulary")
	}
	return v, nil
}

// Check parses one JSON-LD script body and validates it against the vocabulary.
func (v *Vocab) Check(raw []byte) Block {
	b := Block{Issues: []Issue{}}
	src := strings.TrimSpace(string(raw))
	src = strings.TrimSuffix(strings.TrimPrefix(src, "<![CDATA["), "]]>")
	if err := json.Unmarshal([]byte(src), &b.Data); err != nil {
		// Some CMSs HTML-escape the payload.
		if err2 := json.Unmarshal([]byte(html.UnescapeString(src)), &b.Data); err2 != nil {
			b.Raw = src
			b.Issues = append(b.Issues, Issue{"error", "$", "invalid JSON: " + err.Error(), "schema.org"})
			return b
		}
	}
	roots, isArr := b.Data.([]any)
	if !isArr {
		roots = []any{b.Data}
	}
	for i, r := range roots {
		p := "$"
		if isArr {
			p = fmt.Sprintf("$[%d]", i)
		}
		m, ok := r.(map[string]any)
		if !ok {
			b.Issues = append(b.Issues, Issue{"error", p, "top level must be an object", "schema.org"})
			continue
		}
		if c, ok := m["@context"]; !ok {
			b.Issues = append(b.Issues, Issue{"warning", p, "missing @context", "schema.org"})
		} else if s, _ := json.Marshal(c); !strings.Contains(string(s), "schema.org") {
			b.Issues = append(b.Issues, Issue{"warning", p + ".@context", "@context does not reference schema.org", "schema.org"})
		}
		b.Types = append(b.Types, strTypes(m["@type"])...)
		if g, ok := m["@graph"].([]any); ok {
			for _, x := range g {
				if gm, ok := x.(map[string]any); ok {
					b.Types = append(b.Types, strTypes(gm["@type"])...)
				}
			}
		}
		v.walk(r, p, &b, true)
	}
	for i, t := range b.Types {
		b.Types[i] = term(t)
	}
	sort.Slice(b.Issues, func(i, j int) bool {
		x, y := b.Issues[i], b.Issues[j]
		return x.Path < y.Path || x.Path == y.Path && x.Msg < y.Msg
	})
	return b
}

// walk validates node; root is true for top-level and @graph entities.
// Split turns a checked {"@graph": [...]} script into one block per graph entry.
// Issues outside any entry, like a missing @context, go to the first block.
// ponytail: only a single top-level object is split, not arrays of @graph objects.
func Split(b Block) []Block {
	b.Path = "$"
	m, _ := b.Data.(map[string]any)
	g, _ := m["@graph"].([]any)
	if len(g) == 0 {
		return []Block{b}
	}
	out := make([]Block, len(g))
	for i, x := range g {
		out[i] = Block{Path: fmt.Sprintf("$.@graph[%d]", i), InGraph: true, Data: x, Issues: []Issue{}}
		if xm, ok := x.(map[string]any); ok {
			for _, t := range strTypes(xm["@type"]) {
				out[i].Types = append(out[i].Types, term(t))
			}
		}
	}
	for _, is := range b.Issues {
		i := 0
		if rest, ok := strings.CutPrefix(is.Path, "$.@graph["); ok {
			if end := strings.IndexByte(rest, ']'); end > 0 {
				if n, err := strconv.Atoi(rest[:end]); err == nil && n < len(out) {
					i = n
				}
			}
		}
		out[i].Issues = append(out[i].Issues, is)
	}
	return out
}

func (v *Vocab) walk(node any, path string, b *Block, root bool) {
	switch n := node.(type) {
	case []any:
		for i, x := range n {
			v.walk(x, fmt.Sprintf("%s[%d]", path, i), b, root)
		}
	case map[string]any:
		var types []string
		for _, t := range strTypes(n["@type"]) {
			if !isSchemaKey(t) {
				continue
			}
			t = term(t)
			if _, ok := v.Classes[t]; !ok {
				b.Issues = append(b.Issues, Issue{"error", path + ".@type", "unknown type " + t, "schema.org"})
				continue
			}
			types = append(types, t)
		}
		for k, val := range n {
			p := path + "." + k
			if k == "@graph" {
				v.walk(val, p, b, true)
				continue
			}
			if strings.HasPrefix(k, "@") || !isSchemaKey(k) {
				continue
			}
			name := term(k)
			// Action annotations like "query-input" (schema.org/docs/actions.html).
			name = strings.TrimSuffix(strings.TrimSuffix(name, "-input"), "-output")
			prop, ok := v.Props[name]
			switch {
			case !ok:
				msg := "unknown property " + name
				if s, ok := v.lower[strings.ToLower(name)]; ok {
					msg += ", did you mean " + s + "?"
				}
				b.Issues = append(b.Issues, Issue{"error", p, msg, "schema.org"})
			case prop.SupersededBy != "":
				b.Issues = append(b.Issues, Issue{"warning", p, name + " is superseded by " + prop.SupersededBy, "schema.org"})
			case len(types) > 0 && !v.applies(prop, types):
				b.Issues = append(b.Issues, Issue{"warning", p, name + " is not a property of " + strings.Join(types, ", "), "schema.org"})
			}
			v.walk(val, p, b, false)
		}
		v.rules(n, types, path, b, root)
	}
}

func (v *Vocab) applies(p Prop, types []string) bool {
	for _, t := range types {
		for _, d := range p.Domains {
			if v.isA(t, d, 0) {
				return true
			}
		}
	}
	return false
}

func (v *Vocab) isA(t, ancestor string, depth int) bool {
	if t == ancestor {
		return true
	}
	if depth > 32 {
		return false
	}
	for _, p := range v.Classes[t] {
		if v.isA(p, ancestor, depth+1) {
			return true
		}
	}
	return false
}

func strTypes(raw any) []string {
	switch t := raw.(type) {
	case string:
		return []string{t}
	case []any:
		var out []string
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case json.RawMessage:
		var x any
		json.Unmarshal(t, &x)
		return strTypes(x)
	}
	return nil
}

var prefixes = []string{"schema:", "https://schema.org/", "http://schema.org/"}

func term(s string) string {
	for _, p := range prefixes {
		s = strings.TrimPrefix(s, p)
	}
	return s
}

// isSchemaKey skips keys from other vocabularies (og:title, https://example.com/x).
func isSchemaKey(k string) bool {
	return !strings.Contains(k, ":") || term(k) != k
}
