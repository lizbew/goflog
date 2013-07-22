package goflog

import (
    "appengine"
    "appengine/datastore"
    "appengine/urlfetch"
    "appengine/user"
    "fmt"
    "html/template"
    "io"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "net/url"
    "strconv"
    //"path/filepath"
    "time"
    _ "errors"
    //"os"
    "strings"
)

type Greeting struct {
    Author  string
    Content string
    Date    time.Time
}

const (
    TEMPLATE_DIR string = "templates"
    TEMPLATE_THEME_DIR string = "templates/themes/twentyten"
    TEMPLATE_EXT string = ".html"
)
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
    tmplPostList = template.Must(template.ParseFiles("templates/post_list.html"))
    tmplPostEdit = template.Must(template.ParseFiles("templates/post_edit.html"))
    tmplPost     = template.Must(template.ParseFiles(
        "templates/themes/twentyten/single.html",
        "templates/themes/twentyten/header.html",
        "templates/themes/twentyten/footer.html",
        "templates/themes/twentyten/loop-single.html",
        "templates/themes/twentyten/sidebar.html",
        "templates/themes/twentyten/disqus-comment.html",
        //"templates/themes/twentyten/comments.html",
        //"templates/themes/twentyten/comment-form.html",
    ))
    tmpl404 = template.Must(template.ParseFiles(
        "templates/themes/twentyten/404.html",
        "templates/themes/twentyten/header.html",
        "templates/themes/twentyten/footer.html",
    ))
    tmplTerm = template.Must(template.ParseFiles("templates/term.html"))
    blog     = make(map[string]string)
)

//func initTemplate() {
    /*fi, err := os.Stat(TEMPLATE_DIR)
    if err != nil || !fi.IsDir(){
        log.Printf("Directory %v not exists: %v", TEMPLATE_DIR, err)
        return
    }*/
/*    fis, err := ioutil.ReadDir(TEMPLATE_DIR)
    if err != nil {
        log.Printf("Directory %v not exists: %v", TEMPLATE_DIR, err)
        return 
    }
    var tmplFiles []string
    for _, fi := range fis {
        if fi.IsDir() || !strings.HasSuffix(fi.Name(), TEMPLATE_EXT) {
            continue
        }
        tmplFiles = append(tmplFiles, filepath.Join(TEMPLATE_DIR, fi.Name()))
    }
    templates = template.Must(template.ParseFiles(&tmplFiles))
}*/

func init() {
    blog["charset"] = "UTF-8"
    blog["name"] = "Vika's Blog"
    blog["description"] = "a longer way"
    blog["siteurl"] = ""

    //initTemplate()

    http.HandleFunc("/", handleHome)
    http.HandleFunc("/guest", guestHandler)
    http.HandleFunc("/sign", sign)
    http.HandleFunc("/admin/", admin)
    http.HandleFunc("/admin/post", handlePostList)
    http.HandleFunc("/admin/post/edit", postEdit)
    http.HandleFunc("/post", handleViewPost)

    http.HandleFunc("/admin/term", handleTerm)
    http.HandleFunc("/admin/term/edit", handleTermEdit)
    http.HandleFunc("/admin/export", handleExport)
    http.HandleFunc("/admin/maintain", handleMaintain)
    http.HandleFunc("/welcome", welcome)
    http.HandleFunc("/_ah/login_required", openIdHandler)

    http.HandleFunc("/blog/", handleProxy)

}

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
    w.WriteHeader(http.StatusInternalServerError)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    io.WriteString(w, "Internal Server Error")
    c.Errorf("%v", err)
}

func serveNotFound(w http.ResponseWriter, r *http.Request) {
    data := make(map[string]interface{})
    data["Blog"] = blog
    w.WriteHeader(http.StatusNotFound)
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    tmpl404.Execute(w, data)
}

// proxy of viiflog
func handleProxy(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  client := urlfetch.Client(c)

  target_url := "http://localhost:8080"
  if true {
    target_url = "http://viiflog.appspot.com"
  }
  targetURL,err := url.Parse(target_url)

  outreq := new(http.Request)
  *outreq = *r
  outreq.URL.Scheme = targetURL.Scheme
  outreq.URL.Host = targetURL.Host
  //outreq.URL.Path = r.URL.Path
  //outreq.URL.RawQuery = r.URL.RawQuery
  
  // Request.RequestURI can't be set in client requests.
  outreq.RequestURI = ""

  if clientIp, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
     outreq.Header.Set("X-Forwarded-For", clientIp)
  }
  outreq.Header.Set("X-Viifly", "http://blog.viifly.com")

  //resp, err := http.DefaultClient.Do(outreq)
  resp, err := client.Do(outreq)
  if err != nil {
    //panic(err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  if resp.Header != nil {
    for k, v := range resp.Header {
      for _, vv := range v {
        w.Header().Add(k, vv)
      }
    }
  }

  w.WriteHeader(resp.StatusCode)
  defer resp.Body.Close()
  result, err := ioutil.ReadAll(resp.Body)
  if err != nil && err != io.EOF {
    panic(err)
  }
  w.Write(result)
}


func handleHome(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    //handle for path all
    if r.URL.Path != "/" {
        if ref := r.Header.Get("Referer"); ref != "" {
            c.Debugf("Referer: %v", ref)
            if refURL,err := url.Parse(ref); err == nil {
                if strings.Index(refURL.Path, "/webproxy/") >= 0 {
                    q := refURL.Query()
                    prxyForUrlString := q.Get("url")
                    prxyForURL, err :=  url.Parse(prxyForUrlString)
                    if err == nil {
                        prxyForURL.Path =  r.URL.Path
                        prxyForURL.RawQuery =  r.URL.RawQuery
                        prxyForURL.Fragment = r.URL.Fragment
                        q.Set("url", prxyForURL.String())
                        refURL.RawQuery = q.Encode()
                        r.URL = refURL
                        r.Header.Set("Referer", prxyForUrlString)
                        //http.Redirect(w, r, refURL.String(), http.StatusFound)
                        handleWebProxy(w, r)
                        return
                    }
                }        
            }
            
        }
        
        serveNotFound(w, r)
        return
    }

    posts, err := GetLatestPosts(c, 10, true)
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
    data["Posts"] = posts
    data["Blog"] = blog
    data["Categories"] = GetCategories(c)
    if err := templates.Execute(w, data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func handleViewPost(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    var postID int64
    var post *Post = nil
    //var postKey *datastore.Key

    idStr := r.FormValue("id")
    cateIDStr := r.FormValue("category")
    singlePost := true
    var posts []Post

    if i, err := strconv.Atoi(idStr); err != nil {
        c.Infof("Failed to convert str to int64: ", i)
    } else {
        postID = int64(i)
        //postKey = CreatePostKey(c, postID)
        post = GetPostByID(c, postID)
    }
    if post != nil && post.Published {
        //posts[0] = *post
        posts = append(posts, *post)
    } else if cateIDStr != "" {
        if cateID, err := strconv.Atoi(cateIDStr); err == nil {
            posts = GetPostByCategory(c, int64(cateID), true)
        }

        singlePost = false
    } else {
        log.Print("Post not found for URL", r.URL.Path)
        serveNotFound(w, r)
        return
    }

    model := struct {
        Posts      []Post
        Blog       map[string]string
        Categories []Term
    }{
        posts, blog, GetCategories(c),
    }

    tmpl := tmplPost
    if !singlePost {
        tmpl = templates
    }

    if err := tmpl.Execute(w, model); err != nil {
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
