package goflog

import (
    "appengine"
    "appengine/datastore"
    "appengine/user"
    "fmt"
    "html/template"
    "io"
    "net/http"
    "time"
)

type Greeting struct {
    Author  string
    Content string
    Date    time.Time
}

var (
    /* templates = template.Must(template.ParseFiles(
        "templates/home.html",
        "templates/admin.html",
        "templates/post_edit.html",
    ))*/
    templates = template.Must(template.ParseFiles(
        "templates/themes/twentyten/index.html",
        "templates/themes/twentyten/header.html",
        "templates/themes/twentyten/footer.html",
        "templates/themes/twentyten/loop.html",
        "templates/themes/twentyten/sidebar.html",
    ))

    blog = make(map[string]string)
)

func init() {
    blog["charset"] = "UTF-8"
    blog["name"] = "Vika's Blog"
    blog["description"] = "a longer way"
    blog["siteurl"] = ""

    http.HandleFunc("/", handleHome)
    http.HandleFunc("/guest", guestHandler)
    http.HandleFunc("/sign", sign)
    http.HandleFunc("/admin", admin)
    http.HandleFunc("/admin/post", postEdit)
    http.HandleFunc("/post", home)
}

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
    w.WriteHeader(http.StatusInternalServerError)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    io.WriteString(w, "Internal Server Error")
    c.Errorf("%v", err)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    posts, err := getLatestPosts(c, 10)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    /*if err := guestbookTemplate.Execute(w, greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }*/
    /* if err := templates.ExecuteTemplate(w, "home.html", posts); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }*/

    /* for i := 0; i < len(posts); i++ {
       posts[i].HTMLContent = template.HTML(posts[i].Content)
     }*/

    data := make(map[string]interface{})
    data["posts"] = posts
    data["blog"] = blog
    if err := templates.Execute(w, data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

}

func home(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    q := datastore.NewQuery("Greeting").Order("-Date").Limit(10)
    greetings := make([]Greeting, 0, 10)
    if _, err := q.GetAll(c, &greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    /*if err := guestbookTemplate.Execute(w, greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }*/
    if err := templates.ExecuteTemplate(w, "home.html", greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    q := datastore.NewQuery("Greeting").Order("-Date").Limit(10)
    greetings := make([]Greeting, 0, 10)
    if _, err := q.GetAll(c, &greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if err := guestbookTemplate.Execute(w, greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    //fmt.Fprint(w, gustbookForm)
}

var guestbookTemplate = template.Must(template.New("sign").Parse(guestbookTemplateHTML))

const guestbookTemplateHTML = `
<html>
  <body>
    {{range .}}
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      <pre>{{.Content}}</pre>
    {{end}}
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Sign Guestbook"></div>
    </form>
  </body>
</html>
`

const gustbookForm = `
<html>
    <body>
        <form action="/sign" method="post">
          <div><textarea name="content" rows="3" cols="6"> </textarea></div>
          <div><input type="submit" value="Sign Guestbook"></div>
        </form>
    </body>
</html>
`

func sign(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    g := Greeting{
        Content: r.FormValue("content"),
        Date:    time.Now(),
    }

    if u := user.Current(c); u != nil {
        g.Author = u.String()
    }

    _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Greeting", nil), &g)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    /*err := signTemplate.Execute(w, r.FormValue("content"))
      if err != nil {
          http.Error(w, err.Error(), http.StatusInternalServerError)
      }*/
    http.Redirect(w, r, "/", http.StatusFound)
}

var signTemplate = template.Must(template.New("sign").Parse(signTemplateHTML))

const signTemplateHTML = `
<html>
  <body>
    <p>you wrote:</p>
    <pre>{{.}}</pre>
  </body>
</html>
`

func guestHandler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
    fmt.Fprint(w, "Hello, world!", u)
}
