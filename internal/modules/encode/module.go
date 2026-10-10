package encode

import (
	"embed"
	"net/http"
	"time"

	"github.com/enteresanlikk/go-toolset/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(_ platform.Deps) platform.Module {
	return platform.Module{
		Name:        "encode",
		Title:       "Encode / Decode",
		Description: "Base64, URL encoding, JWT decoding and hashes. Input is sent in the request body, never in the URL.",
		UI:          ui,
		Views: []platform.View{
			{Path: "/encode", Dir: "base64", Title: "Base64"},
			{Path: "/encode/url", Dir: "url", Title: "URL"},
			{Path: "/encode/jwt", Dir: "jwt", Title: "JWT"},
			{Path: "/encode/hash", Dir: "hash", Title: "Hash"},
		},
		Routes: []platform.Route{
			{Pattern: "POST /api/encode/base64", Example: `{"input":"hello","mode":"encode","url_safe":"false"}`, Handler: PostBase64},
			{Pattern: "POST /api/encode/url", Example: `{"input":"a b&c","mode":"encode"}`, Handler: PostURL},
			{Pattern: "POST /api/encode/jwt", Example: `{"token":"eyJ..."}`, Handler: PostJWT},
			{Pattern: "POST /api/encode/hash", Example: `{"input":"hello"}`, Handler: PostHash},
		},
	}
}

type request struct {
	Input   string `json:"input"`
	Mode    string `json:"mode"`
	URLSafe string `json:"url_safe"`
	Token   string `json:"token"`
}

func decode(w http.ResponseWriter, r *http.Request) (request, bool) {
	var req request
	if err := platform.DecodeJSON(r, &req); err != nil {
		platform.Reply(w, nil, err)
		return req, false
	}
	return req, true
}

func PostBase64(w http.ResponseWriter, r *http.Request) {
	if req, ok := decode(w, r); ok {
		res, err := Base64(req.Input, req.Mode, req.URLSafe == "true")
		platform.Reply(w, res, err)
	}
}

func PostURL(w http.ResponseWriter, r *http.Request) {
	if req, ok := decode(w, r); ok {
		res, err := URL(req.Input, req.Mode)
		platform.Reply(w, res, err)
	}
}

func PostJWT(w http.ResponseWriter, r *http.Request) {
	if req, ok := decode(w, r); ok {
		res, err := DecodeJWT(req.Token, time.Now())
		platform.Reply(w, res, err)
	}
}

func PostHash(w http.ResponseWriter, r *http.Request) {
	if req, ok := decode(w, r); ok {
		platform.Reply(w, Hash(req.Input), nil)
	}
}
