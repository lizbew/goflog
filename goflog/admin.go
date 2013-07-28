package goflog

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

func init() {
	http.HandleFunc("/admin/", admin)
	http.HandleFunc("/admin/post", handlePostList)
	http.HandleFunc("/admin/post/edit", postEdit)
	http.HandleFunc("/admin/term", handleTerm)
	http.HandleFunc("/admin/term/edit", handleTermEdit)
	http.HandleFunc("/admin/export", handleExport)
	http.HandleFunc("/admin/maintain", handleMaintain)
	http.HandleFunc("/admin/info/", handleServerInfo)
}

func GetCurrentUserKey(c appengine.Context) *datastore.Key {
	u := user.Current(c)
	if u != nil {
		userKey := datastore.NewKey(c, "User", u.Email, 0, nil)

		var blogUser User
		err := datastore.Get(c, userKey, &blogUser)
		if err != nil {
			blogUser = User{Email: u.Email, Nicename: "admin", Active: true, Registered: time.Now()}
			datastore.Put(c, userKey, &blogUser)
		}
		return userKey
	}
	return nil
}

func savePost(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	currentUserKey := GetCurrentUserKey(c)
	var postID int64
	var post *Post = nil
	var postKey *datastore.Key

	idStr := r.FormValue("id")
	if i, err := strconv.Atoi(idStr); err != nil {
		c.Infof("Failed to convert str to int64: ", i)
	} else {
		postID = int64(i)
		postKey = CreatePostKey(c, postID)
		post = GetPostByID(c, postID)
	}

	cateIDStr := r.FormValue("postCategory")
	var categoryID int64 = 0
	//tags := r.FormValue("postTag")
	publishStr := r.FormValue("postPublished")
	published := false
	if publishStr == "published" {
		published = true
	}
	if cateID, err := strconv.Atoi(cateIDStr); err == nil {
		cateTerm := GetTermIDMap(c)[int64(cateID)]
		if cateTerm != nil {
			categoryID = int64(cateID)
		}
	}
	if post != nil {
		post.Title = r.FormValue("postTitle")
		post.Content = []byte(r.FormValue("postContent"))
		post.Modified = time.Now()
		post.CategoryID = categoryID
		post.Published = published
	} else {
		postKey = NewPostKey(c)
		postID = postKey.IntID()
		post = &Post{
			ID:         postID,
			Title:      r.FormValue("postTitle"),
			Content:    []byte(r.FormValue("postContent")),
			Created:    time.Now(),
			Modified:   time.Now(),
			Author:     currentUserKey,
			CategoryID: categoryID,
			Published:  published,
		}
	}

	if _, err := datastore.Put(c, postKey, post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/post", http.StatusFound)
}

func admin(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, "admin", http.StatusOK, nil)
}

func postEdit(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	modifyPost := false

	if r.Method == "GET" {
		idStr := r.FormValue("id")
		c.Debugf("idStr=%v", idStr)
		var post *Post
		var postID int64
		if idStr != "" {

			if i, err := strconv.Atoi(idStr); err != nil {
				c.Infof("Failed to convert str to int64: ", i)
			} else {
				postID = int64(i)
				post = GetPostByID(c, postID)
				if post != nil {
					modifyPost = true
				}
			}

		}

		executeTemplate(w, "adminPostEdit", http.StatusOK, map[string]interface{}{
			"ModifyPost":      modifyPost,
			"Post":            post,
			"PostID":          postID,
			"TermCategoryMap": GetTermCategoryMap(c),
		})
		return
	}
	savePost(w, r)
}

func handlePostList(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	posts, err := GetLatestPosts(c, 0, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logout_url, _ := user.LogoutURL(c, "/")
	executeTemplate(w, "adminPost", http.StatusOK, map[string]interface{}{
		"Posts":           posts,
		"TermCategoryMap": GetTermCategoryMap(c),
		"LogoutUrl":       logout_url,
	})
}

func handleTerm(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var termID int64 = 0
	if i := r.FormValue("id"); i != "" {
		if id, err := strconv.Atoi(i); err == nil {
			termID = int64(id)
		}
	}

	terms := GetAllTerms(c)
	var currentTerm Term
	if termID > 0 {

		for _, t := range terms {
			if t.ID == termID {
				currentTerm = t
			}
		}
	}

	executeTemplate(w, "adminTerm", http.StatusOK, map[string]interface{}{
		"Terms":       terms,
		"CurrentTerm": &currentTerm,
	})
}

func handleTermEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.Redirect(w, r, "/admin/term", http.StatusFound)
		return
	}
	c := appengine.NewContext(r)
	var termID int64 = 0
	if i := r.FormValue("termID"); i != "" {
		if id, err := strconv.Atoi(i); err == nil {
			termID = int64(id)
		}
	}

	term := Term{
		ID:          termID,
		Name:        r.FormValue("termName"),
		Taxonomy:    r.FormValue("termTaxonomy"),
		Description: r.FormValue("termDescription"),
		Slug:        r.FormValue("termSlug"),
	}
	SaveTerm(c, &term)
	http.Redirect(w, r, "/admin/term", http.StatusFound)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	posts, err := GetLatestPosts(c, 0, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model := struct {
		Posts []Post
	}{
		Posts: posts,
	}
	exportTmpl := template.Must(template.ParseFiles("templates/admin/post_export.xml"))
	exportTmpl.Execute(w, model)
}

func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	info := map[string]string{
		"AppID":                  appengine.AppID(c),
		"Datacenter":             appengine.Datacenter(),
		"DefaultVersionHostname": appengine.DefaultVersionHostname(c),
		"InstanceID":             appengine.InstanceID(),
		"IsDevAppServer":         strconv.FormatBool(appengine.IsDevAppServer()),
		"RequestID":              appengine.RequestID(c),
		"ServerSoftware":         appengine.ServerSoftware(),
		"VersionID":              appengine.VersionID(c),
	}
	executeTemplate(w, "serverInfo", http.StatusOK, map[string]interface{}{
		"Info": info,
	})
}
