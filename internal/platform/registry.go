package platform

import (
	"bytes"
	"html/template"
	"io/fs"
	"net/http"
	"sort"
	"time"

	"github.com/enteresanlikk/go-toolkit/web"
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

var layout = template.Must(template.New("layout").Funcs(template.FuncMap{"asset": web.Asset}).Parse(web.Layout))

var kitCSS = func() string {
	b, err := fs.ReadFile(web.StaticFS, "css/kit.css")
	if err != nil {
		panic(err)
	}
	return string(b)
}()

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
	mux.Handle("GET "+web.StaticPrefix, http.StripPrefix(web.StaticPrefix, http.FileServerFS(web.StaticFS)))
	mux.Handle("GET /{$}", render(page{"Toolkit", nil, "/", "Network and developer tools. Every tool has a JSON API too, see /api.", "", mods}))
	mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) { Reply(w, mods, nil) })
	return defaultHeaders(mux)
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

func defaultHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd.Set("X-Robots-Tag", "noindex, nofollow")
		hd.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		hd.Set("Pragma", "no-cache")
		hd.Set("Expires", "0")
		h.ServeHTTP(w, r)
	})
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
