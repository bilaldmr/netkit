package encode

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

type Result struct {
	Output string `json:"output"`
	Binary bool   `json:"binary,omitempty"`
}

func Base64(input, mode string, urlSafe bool) (Result, error) {
	enc := base64.StdEncoding
	if urlSafe {
		enc = base64.URLEncoding
	}
	switch mode {
	case "", "encode":
		return Result{Output: enc.EncodeToString([]byte(input))}, nil
	case "decode":
		b, err := decodeBase64(strings.TrimSpace(input))
		if err != nil {
			return Result{}, errors.New("invalid base64: " + err.Error())
		}
		if !utf8.Valid(b) {
			return Result{Output: hex.EncodeToString(b), Binary: true}, nil
		}
		return Result{Output: string(b)}, nil
	}
	return Result{}, errors.New("mode must be encode or decode")
}

func decodeBase64(s string) ([]byte, error) {
	s = strings.NewReplacer("\n", "", "\r", "", " ", "").Replace(s)
	var firstErr error
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding} {
		b, err := enc.DecodeString(s)
		if err == nil {
			return b, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return nil, firstErr
}

type URLResult struct {
	Query string `json:"query,omitempty"`
	Path  string `json:"path,omitempty"`
	Text  string `json:"text,omitempty"`
}

func URL(input, mode string) (URLResult, error) {
	switch mode {
	case "", "encode":
		return URLResult{Query: url.QueryEscape(input), Path: url.PathEscape(input)}, nil
	case "decode":
		s, err := url.QueryUnescape(input)
		if err != nil {
			return URLResult{}, errors.New("invalid percent-encoding: " + err.Error())
		}
		return URLResult{Text: s}, nil
	}
	return URLResult{}, errors.New("mode must be encode or decode")
}

type JWT struct {
	Header    map[string]any `json:"header"`
	Payload   map[string]any `json:"payload"`
	Signature string         `json:"signature"`
	Algorithm string         `json:"algorithm,omitempty"`
	IssuedAt  string         `json:"issued_at,omitempty"`
	NotBefore string         `json:"not_before,omitempty"`
	ExpiresAt string         `json:"expires_at,omitempty"`
	Expired   *bool          `json:"expired,omitempty"`
}

func DecodeJWT(token string, now time.Time) (JWT, error) {
	parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(token), "Bearer ")), ".")
	if len(parts) != 3 {
		return JWT{}, errors.New("a JWT has 3 dot-separated parts")
	}
	var j JWT
	for i, dst := range []*map[string]any{&j.Header, &j.Payload} {
		b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[i], "="))
		if err != nil {
			return JWT{}, errors.New([]string{"header", "payload"}[i] + ": invalid base64url")
		}
		if err := json.Unmarshal(b, dst); err != nil {
			return JWT{}, errors.New([]string{"header", "payload"}[i] + ": invalid JSON")
		}
	}
	j.Signature = parts[2]
	j.Algorithm, _ = j.Header["alg"].(string)
	claim := func(k string) (time.Time, bool) {
		f, ok := j.Payload[k].(float64)
		return time.Unix(int64(f), 0).UTC(), ok
	}
	if t, ok := claim("iat"); ok {
		j.IssuedAt = t.Format(time.RFC3339)
	}
	if t, ok := claim("nbf"); ok {
		j.NotBefore = t.Format(time.RFC3339)
	}
	if t, ok := claim("exp"); ok {
		j.ExpiresAt = t.Format(time.RFC3339)
		expired := now.After(t)
		j.Expired = &expired
	}
	return j, nil
}

func Hash(input string) map[string]string {
	b := []byte(input)
	m5, s1, s256, s512 := md5.Sum(b), sha1.Sum(b), sha256.Sum256(b), sha512.Sum512(b)
	return map[string]string{
		"md5":    hex.EncodeToString(m5[:]),
		"sha1":   hex.EncodeToString(s1[:]),
		"sha256": hex.EncodeToString(s256[:]),
		"sha512": hex.EncodeToString(s512[:]),
	}
}
