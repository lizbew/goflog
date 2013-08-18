package goflog

//copied function from https://github.com/gorilla/site/blob/master/gorillaweb/template.go

import (
	"errors"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

var (
	ErrTemplateNotFound = errors.New("Template Not Found")
)

func urlFmt(path string) string {
	u := url.URL{Path: path}
	return u.String()
}

func gistCopyAtFmt(t time.Time) string {
	return t.Format(time.RFC3339)
}

func executeTemplate(w http.ResponseWriter, name string, status int, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if tpl, ok := tpls[name]; ok {
		return tpl.ExecuteTemplate(w, "base", data)
	}
	return ErrTemplateNotFound
}

var tpls = map[string]*template.Template{
	"admin":         newTemplate("templates/admin/base.html", "templates/admin/index.html"),
	"adminFile":     newTemplate("templates/admin/base.html", "templates/admin/file.html"),
	"adminPage":     newTemplate("templates/admin/base.html", "templates/admin/page.html"),
	"adminPageEdit": newTemplate("templates/admin/base.html", "templates/admin/page_edit.html"),
	"adminPost":     newTemplate("templates/admin/base.html", "templates/admin/post.html"),
	"adminPostEdit": newTemplate("templates/admin/base.html", "templates/admin/post_edit.html"),
	"adminTerm":     newTemplate("templates/admin/base.html", "templates/admin/term.html"),
	"adminGist":     newTemplate("templates/admin/base.html", "templates/admin/gist.html"),
	"serverInfo":    newTemplate("templates/admin/base.html", "templates/admin/server_info.html"),
	"webproxy":      newTemplate("templates/admin/webproxy.html"),
	"home":          newTemplate("templates/index.html"),
}

var funcs = template.FuncMap{
	"url": urlFmt,
	"gistCopyAtFmt": gistCopyAtFmt,
}

func newTemplate(files ...string) *template.Template {
	return template.Must(template.New("*").Funcs(funcs).ParseFiles(files...))
}
