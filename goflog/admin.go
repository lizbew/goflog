package goflog

import (
    "appengine"
    "appengine/datastore"
    "appengine/user"
    _ "fmt"
    _ "html/template"
    _ "io"
    "net/http"
    "time"
)

func savePost(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    p := Post{
        Title:    r.FormValue("postTitle"),
        Content:  r.FormValue("postContent"),
        Created:  time.Now(),
        Modified: time.Now(),
    }

    u := user.Current(c)
    if u != nil {
        userKey := datastore.NewKey(c, "User", u.Email, 0, nil)

        var blogUser User
        err := datastore.Get(c, userKey, &blogUser)
        if err != nil {
            blogUser = User{Email: u.Email, Nicename: "admin", Active: true, Registered: time.Now()}
            datastore.Put(c, userKey, &blogUser)
        }
        p.Author = userKey
    }

    _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Post", nil), &p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/", http.StatusFound)

}

func admin(w http.ResponseWriter, r *http.Request) {
    if err := templates.ExecuteTemplate(w, "admin.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func postEdit(w http.ResponseWriter, r *http.Request) {
    if (r.Method == "GET") {
      if err := templates.ExecuteTemplate(w, "post_edit.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
      }
    return;
    }
    //POST
    //fmt.Fprint(w, r.FormValue("content"));
savePost(w, r);
}
