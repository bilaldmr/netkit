package platform

import (
	"bytes"
	_ "embed"
	"html/template"
	"io/fs"
	"net/http"
	"sort"
	"time"
)

type Module struct {
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	UI          fs.FS   `json:"-"`
	Views       []View  `json:"views,omitempty"`
	Routes      []Route `json:"routes"`
}

type View struct {
	Path  string `json:"path"`
	Dir   string `json:"-"`
	Title string `json:"title"`
}

type Route struct {
	Pattern string           `json:"pattern"`
	Example string           `json:"example,omitempty"`
	Handler http.HandlerFunc `json:"-"`
}

type Factory func(Deps) Module

//go:embed layout.html
var layoutHTML string

//go:embed kit.js
var kitJS []byte

//go:embed kit.css
var kitCSS string

var layout = template.Must(template.New("layout").Parse(layoutHTML))

type page struct {
	Title       string
	Mod         *Module
	Path        string
	Description string
	Body        template.HTML
	Mods        []Module
}

func NewServer(d Deps, factories ...Factory) http.Handler {
	var mods []Module
	for _, f := range factories {
		mods = append(mods, f(d))
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].Name < mods[j].Name })

	mux := http.NewServeMux()
	for i := range mods {
		m := &mods[i]
		for _, rt := range m.Routes {
			mux.Handle(rt.Pattern, http.TimeoutHandler(rt.Handler, 15*time.Second, `{"error":"timeout"}`))
		}
		if len(m.Views) == 0 {
			continue
		}
		ui, err := fs.Sub(m.UI, "ui")
		if err != nil {
			panic(err)
		}
		prefix := "/ui/" + m.Name
		mux.Handle("GET "+prefix+"/", http.StripPrefix(prefix, http.FileServerFS(ui)))
		shared, _ := fs.ReadFile(ui, "shared.css")
		for _, v := range m.Views {
			title := m.Title
			if len(m.Views) > 1 {
				title = v.Title + " · " + m.Title
			}
			mux.Handle("GET "+v.Path, render(page{title, m, v.Path, m.Description, component(m.Name, v.Dir, ui, shared), mods}))
		}
	}
	mux.Handle("GET /ui/kit.js", script(kitJS))
	mux.Handle("GET /{$}", render(page{"Toolset", nil, "/", "Network and developer tools. Every tool has a JSON API too, see /api.", "", mods}))
	mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) { Reply(w, mods, nil) })
	return noIndex(mux)
}

func component(mod, dir string, ui fs.FS, shared []byte) template.HTML {
	if _, err := fs.Stat(ui, dir+"/view.js"); err != nil {
		panic("module " + mod + ": ui/" + dir + "/view.js is required: " + err.Error())
	}
	html, _ := fs.ReadFile(ui, dir+"/view.html")
	css, _ := fs.ReadFile(ui, dir+"/view.css")
	tag := "tool-" + mod + "-" + dir
	return template.HTML("<" + tag + `><template shadowrootmode="open">` +
		"<style>" + kitCSS + string(shared) + string(css) + "</style>" + string(html) +
		`<div class="err" role="alert"></div><div class="out" aria-live="polite"></div>` +
		"</template></" + tag + `><script type="module" src="/ui/` + mod + "/" + dir + `/view.js"></script>`)
}

func noIndex(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		h.ServeHTTP(w, r)
	})
}

func script(js []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Write(js)
	}
}

func render(p page) http.HandlerFunc {
	var buf bytes.Buffer
	if err := layout.Execute(&buf, p); err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(buf.Bytes())
	}
}
