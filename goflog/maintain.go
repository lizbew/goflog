package goflog

import (
    //"appengine"
    //"appengine/datastore"
    "io"
    "net/http"
    //"time"
)

func handleMaintain(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    io.WriteString(w, "No maintain task currently!")
}

