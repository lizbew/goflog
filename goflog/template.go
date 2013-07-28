package goflog

//copied function from https://github.com/gorilla/site/blob/master/gorillaweb/template.go

import (
	"net/http"
	"net/url"
	"text/template"
)

func urlFmt(path string) string {
	u := url.URL{Path: path}
	return u.String()
}

func executeTemplate(w http.ResponseWriter, name string, status int, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return tpls[name].ExecuteTemplate(w, "base", data)
}

var tpls = map[string]*template.Template{
	"admin":         newTemplate("templates/admin/base.html", "templates/admin/index.html"),
	"adminFile":     newTemplate("templates/admin/base.html", "templates/admin/file.html"),
	"adminPage":     newTemplate("templates/admin/base.html", "templates/admin/page.html"),
	"adminPageEdit": newTemplate("templates/admin/base.html", "templates/admin/page_edit.html"),
	"adminPost":     newTemplate("templates/admin/base.html", "templates/admin/post.html"),
	"adminPostEdit": newTemplate("templates/admin/base.html", "templates/admin/post_edit.html"),
	"adminTerm":     newTemplate("templates/admin/base.html", "templates/admin/term.html"),
	"serverInfo":    newTemplate("templates/admin/base.html", "templates/admin/server_info.html"),
	//"home":          newTemplate("templates/index.html"),
}

var funcs = template.FuncMap{
	"url": urlFmt,
}

func newTemplate(files ...string) *template.Template {
	return template.Must(template.New("*").Funcs(funcs).ParseFiles(files...))
}
