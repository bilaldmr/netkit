package web

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

//go:embed static
var staticFiles embed.FS

//go:embed template/layout.html
var Layout string

const StaticPrefix = "/static/"

var StaticFS = mustSub(staticFiles, "static")

func Asset(name string) string {
	return path.Join(StaticPrefix, strings.TrimPrefix(name, "/"))
}

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic("web: embed sub " + dir + ": " + err.Error())
	}
	return sub
}
