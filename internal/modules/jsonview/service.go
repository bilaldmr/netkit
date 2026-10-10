package jsonview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

const MaxSize = 5 << 20

type Service struct{ HTTP Doer }

func (s Service) Fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, MaxSize))
}

func Format(raw []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		return nil, errors.New("invalid json: " + err.Error())
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}
