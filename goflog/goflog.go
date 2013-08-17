package goflog

import (
	"appengine"
	"encoding/json"
	_ "errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	templates = template.Must(template.ParseFiles(
		"templates/themes/twentyten/index.html",
		"templates/themes/twentyten/header.html",
		"templates/themes/twentyten/footer.html",
		"templates/themes/twentyten/loop.html",
		"templates/themes/twentyten/sidebar.html",
	))

	tmplPost = template.Must(template.ParseFiles(
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
	blog = make(map[string]string)
)

func init() {
	blog["charset"] = "UTF-8"
	blog["name"] = "Vika's Blog"
	blog["description"] = "a longer way"
	blog["siteurl"] = ""
	//blog["blog_url"] = "http://localhost:8012"
	loadConfig()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/post", handleViewPost)
}

func loadConfig() {
	var configFile string = "config.json"
	if strings.Index(appengine.ServerSoftware(), "Dev") >= 0 {
		configFile = "config_dev.json"
	}
	log.Println("Start to load config from file ", configFile)

	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("Error when read config file %v, %v", configFile, err)
		return
	}
	log.Printf("config file length is %v", len(buf))

	config := make(map[string]interface{}, 0)
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Printf("Error when Unmarshal config json: %v", err)
		return
	}
	for k, v := range config {
		blog[k] = v.(string)
	}
}

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
	//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//w.WriteHeader(http.StatusInternalServerError)
	//io.WriteString(w, err.Error())
	c.Errorf("serveError: %v", err)
	var errText string = err.Error()
	if !appengine.IsDevAppServer() {
		errText = "Internal Server Error"
	}
	http.Error(w, errText, http.StatusInternalServerError)
}

func serveNotFound(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["Blog"] = blog
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl404.Execute(w, data)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	//handle for path all
	if r.URL.Path != "/" {
		if ref := r.Header.Get("Referer"); ref != "" {
			c.Debugf("Referer: %v", ref)
			if refURL, err := url.Parse(ref); err == nil {
				if strings.Index(refURL.Path, "/webproxy/") >= 0 {
					c.Debugf("Handle request reference from webproxy")
					if originProxyUrlString, err := DecodeProxyUrl(refURL.Query().Get("url")); err == nil {
						if prxyForURL, err := url.Parse(originProxyUrlString); err == nil {
							prxyForURL.Path = r.URL.Path
							prxyForURL.RawQuery = r.URL.RawQuery
							prxyForURL.Fragment = r.URL.Fragment
							fetchUrlToResponse(c, w, prxyForURL.String())
							return
						}
					}
				}
			}
		}

		serveNotFound(w, r)
		return
	}

	executeTemplate(w, "home", http.StatusOK, nil)
}

func handleViewPost(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	//var postKey *datastore.Key

	idStr := r.FormValue("id")
	cateIDStr := r.FormValue("category")
	singlePost := true
	var posts []Post

	if idStr != "" {
		if i, err := strconv.Atoi(idStr); err == nil {
			post := GetPostByID(c, int64(i))
			if post != nil && post.Published {
				//posts[0] = *post
				posts = append(posts, *post)
			}
		} else {
			c.Debugf("Failed to convert str to int64 for post_id: ,%v", idStr)
		}

	} else if cateIDStr != "" {
		singlePost = false

		if cateID, err := strconv.Atoi(cateIDStr); err == nil {
			posts = GetPostByCategory(c, int64(cateID), true)
		}
	} else {
		singlePost = false
		var err error
		posts, err = GetLatestPosts(c, 10, true)
		if err != nil {
			c.Errorf("Error when GetLatestPosts: %v", err)
		}
	}

	if len(posts) == 0 {
		c.Debugf("Post not found for URL %v", r.URL.Path)
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
		serveError(c, w, err)
	}
}
