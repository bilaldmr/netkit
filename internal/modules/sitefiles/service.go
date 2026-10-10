package sitefiles

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

var Paths = []string{"/robots.txt", "/sitemap.xml", "/.well-known/security.txt", "/llms.txt"}

const MaxBody = 256 << 10

type File struct {
	Path        string `json:"path"`
	URL         string `json:"url"`
	Status      int    `json:"status,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Body        string `json:"body,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
	Error       string `json:"error,omitempty"`
}

type Service struct{ HTTP Doer }

func (s Service) Get(ctx context.Context, raw string) ([]File, error) {
	if raw == "" {
		return nil, errors.New("url required")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return nil, errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("only http and https urls are supported")
	}
	files := make([]File, len(Paths))
	var wg sync.WaitGroup
	for i, p := range Paths {
		wg.Go(func() { files[i] = s.fetch(ctx, u.Scheme+"://"+u.Host+p, p) })
	}
	wg.Wait()
	return files, nil
}

func (s Service) fetch(ctx context.Context, target, path string) File {
	f := File{Path: path, URL: target}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		f.Error = err.Error()
		return f
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		f.Error = err.Error()
		return f
	}
	defer resp.Body.Close()
	f.URL, f.Status, f.ContentType = resp.Request.URL.String(), resp.StatusCode, resp.Header.Get("Content-Type")
	if resp.StatusCode != http.StatusOK {
		return f
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, MaxBody+1))
	if err != nil {
		f.Error = err.Error()
	}
	f.Truncated = len(b) > MaxBody
	f.Body = string(b[:min(len(b), MaxBody)])
	return f
}
