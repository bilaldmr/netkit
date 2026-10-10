package redirects

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

const MaxHops = 10

type Hop struct {
	URL      string `json:"url"`
	Status   int    `json:"status"`
	Location string `json:"location,omitempty"`
	Server   string `json:"server,omitempty"`
}

type Trace struct {
	Final string `json:"final,omitempty"`
	Hops  []Hop  `json:"hops"`
	Error string `json:"error,omitempty"`
}

type Service struct{ HTTP Doer }

func (s Service) Trace(ctx context.Context, url string) (Trace, error) {
	if url == "" {
		return Trace{}, errors.New("url required")
	}
	if !strings.Contains(url, "://") {
		url = "http://" + url
	}
	var t Trace
	for range MaxHops {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return t, err
		}
		resp, err := s.HTTP.Do(req)
		if err != nil {
			t.Error = err.Error()
			return t, nil
		}
		resp.Body.Close()
		hop := Hop{URL: url, Status: resp.StatusCode, Server: resp.Header.Get("Server")}
		loc := resp.Header.Get("Location")
		if resp.StatusCode/100 != 3 || loc == "" {
			t.Hops, t.Final = append(t.Hops, hop), url
			return t, nil
		}
		next, err := req.URL.Parse(loc)
		if err != nil {
			t.Hops, t.Error = append(t.Hops, hop), "bad location: "+err.Error()
			return t, nil
		}
		hop.Location = next.String()
		t.Hops, url = append(t.Hops, hop), next.String()
	}
	t.Error = "too many redirects"
	return t, nil
}
