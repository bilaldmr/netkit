package headers

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Result struct {
	Final   string            `json:"final"`
	Status  int               `json:"status"`
	Proto   string            `json:"proto"`
	Headers map[string]string `json:"headers"`
}

type Service struct{ HTTP Doer }

func (s Service) Get(ctx context.Context, url string) (Result, error) {
	if url == "" {
		return Result{}, errors.New("url required")
	}
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return Result{}, err
	}
	resp.Body.Close()
	h := make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		h[k] = strings.Join(v, "\n")
	}
	return Result{Final: resp.Request.URL.String(), Status: resp.StatusCode, Proto: resp.Proto, Headers: h}, nil
}
