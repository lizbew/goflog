package goflog

import (
    "appengine"
    "appengine/datastore"
    "io"
    "net/http"
    "time"
)

type Post1 struct {
    ID           int64
    Title        string
    Content      string
    Published    bool
    Author       *datastore.Key
    Created      time.Time
    Modified     time.Time
    AuthorObj    User `datastore:"-"`
    CategoryID   int64
    CategoryTerm *Term `datastore:"-"`
    Tags         []string
    Category     string
    //Categories []Term `datastore:"-"`
    //Tags       []Term `datastore:"-"`
}

func handleMaintain(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    q := datastore.NewQuery("Post")
    // var ps []Post
    for it := q.Run(c); ; {
        var post Post1
        key, err := it.Next(&post)
        if err == datastore.Done {
            break
        }
        if err != nil {
            c.Errorf("Failed when query Post by category, ")
            break
        }
        post.ID = key.IntID()
        p2 := copyPostToNew(post)
        datastore.Put(c, key, p2)
    }
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    io.WriteString(w, "Done")
}

func copyPostToNew(post Post1) *Post {
    var p2 Post
    p2.ID = post.ID
    p2.Title = post.Title
    p2.Content = []byte(post.Content)
    p2.Published = post.Published
    p2.Author = post.Author
    p2.Created = post.Created
    p2.Modified = post.Modified
    p2.CategoryID = post.CategoryID
    p2.Category = post.Category
    p2.Tags = post.Tags

    return &p2
}
