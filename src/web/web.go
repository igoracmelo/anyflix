package web

import (
	"embed"
	_ "embed"
	"io/fs"
	"strings"
	"sync"
	"text/template"
)

//go:embed fs
var embedFS embed.FS

var FS fs.FS
var PublicFS fs.FS

func init() {
	var err error
	FS, err = fs.Sub(embedFS, "fs")
	if err != nil {
		panic(err)
	}

	PublicFS, err = fs.Sub(embedFS, "fs/public")
	if err != nil {
		panic(err)
	}
}

var Template = tmpl{
	templates:   map[string]*template.Template{},
	templatesMu: &sync.Mutex{},
}

type tmpl struct {
	templates   map[string]*template.Template
	templatesMu *sync.Mutex
}

func (t tmpl) Load(patterns ...string) (tmpl *template.Template, err error) {
	key := strings.Join(patterns, "|")

	t.templatesMu.Lock()
	defer t.templatesMu.Unlock()

	tmpl, ok := t.templates[key]
	if ok {
		return
	}

	tmpl, err = template.ParseFS(FS, patterns...)
	if err != nil {
		return
	}

	t.templates[key] = tmpl
	return
}

func (t tmpl) MustLoad(patterns ...string) *template.Template {
	return template.Must(t.Load(patterns...))
}
