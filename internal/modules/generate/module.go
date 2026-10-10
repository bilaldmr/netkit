package generate

import (
	"embed"
	"net/http"
	"strconv"
	"time"

	"github.com/enteresanlikk/go-toolkit/internal/platform"
)

//go:embed ui
var ui embed.FS

func New(_ platform.Deps) platform.Module {
	return platform.Module{
		Name:        "generate",
		Title:       "Generators",
		Description: "UUIDs, strong passwords and timestamp conversion",
		UI:          ui,
		Views: []platform.View{
			{Path: "/generate", Dir: "uuid", Title: "UUID"},
			{Path: "/generate/password", Dir: "password", Title: "Password"},
			{Path: "/generate/timestamp", Dir: "timestamp", Title: "Timestamp"},
		},
		Routes: []platform.Route{
			{Pattern: "GET /api/generate/uuid", Example: "/api/generate/uuid?version=7&count=5", Handler: GetUUID},
			{Pattern: "GET /api/generate/password", Example: "/api/generate/password?length=24&symbols=false", Handler: GetPassword},
			{Pattern: "GET /api/generate/timestamp", Example: "/api/generate/timestamp?value=1700000000&tz=Europe/Istanbul", Handler: GetTimestamp},
		},
	}
}

func intParam(r *http.Request, key string, def int) int {
	if n, err := strconv.Atoi(r.URL.Query().Get(key)); err == nil {
		return n
	}
	return def
}

func boolParam(r *http.Request, key string) bool {
	return r.URL.Query().Get(key) != "false"
}

func GetUUID(w http.ResponseWriter, r *http.Request) {
	ids, err := UUID(intParam(r, "version", 4), intParam(r, "count", 1))
	platform.Reply(w, map[string]any{"uuids": ids}, err)
}

func GetPassword(w http.ResponseWriter, r *http.Request) {
	pws, err := Password(intParam(r, "length", 20), intParam(r, "count", 1),
		boolParam(r, "lower"), boolParam(r, "upper"), boolParam(r, "digits"), boolParam(r, "symbols"))
	platform.Reply(w, map[string]any{"passwords": pws}, err)
}

func GetTimestamp(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t, err := Timestamp(q.Get("value"), q.Get("tz"), time.Now())
	platform.Reply(w, t, err)
}
