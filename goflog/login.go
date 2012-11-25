package goflog

import (
    "appengine"
    "appengine/user"
    "fmt"
    "net/http"
)

func welcome(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-type", "text/html; charset=utf-8")
    c := appengine.NewContext(r)
    u := user.Current(c)
    if u == nil {
        url, _ := user.LoginURL(c, "/")
        fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
        return
    }
    url, _ := user.LogoutURL(c, "/")
    fmt.Fprintf(w, `Welcome, %s! (<a href="%s">sing out</a>`, u, url)
}

func openIdHandler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    url, _ := user.LoginURL(c, "/")
    fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
}
