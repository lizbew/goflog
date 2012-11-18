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
    "log"
    "strconv"
)

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
    if i,err := strconv.Atoi(idStr); err != nil {
              c.Infof("Failed to convert str to int64: ", i)            
          } else {
            postID = int64(i)
            postKey = CreatePostKey(c, postID)
            post = GetPostByID(c, postID)
         }
   
    if post != nil {
       post.Title = r.FormValue("postTitle")
       post.Content =r.FormValue("postContent")
       post.Modified = time.Now()

    } else {


    postKey = NewPostKey(c)
    postID = postKey.IntID()
    post = &Post{
        ID: postID,
        Title:    r.FormValue("postTitle"),
        Content:  r.FormValue("postContent"),
        Created:  time.Now(),
        Modified: time.Now(),
        Author: currentUserKey,
    }
}
  
    
   if  _, err := datastore.Put(c, postKey, post); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/admin/post", http.StatusFound)
}

func admin(w http.ResponseWriter, r *http.Request) {
    /*if err := templates.ExecuteTemplate(w, "admin.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }*/
    log.Print("Server admin from file")
    http.ServeFile(w, r, "templates/admin.html")
}

func postEdit(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    modifyPost := false

    if r.Method == "GET" {
        /*if err := templates.ExecuteTemplate(w, "post_edit.html", nil); err != nil {
          http.Error(w, err.Error(), http.StatusInternalServerError)
        }*/
        idStr := r.FormValue("id")
        //http.ServeFile(w, r, "templates/post_edit.html")
        //log.Print("r.URL.Path = ", r.URL.Path)
        log.Print("idStr=", idStr)
        var post *Post
        var postID int64
        if idStr != "" {
           
          if i,err := strconv.Atoi(idStr); err != nil {
              c.Infof("Failed to convert str to int64: ", i)            
          } else {
            postID = int64(i)
            post = GetPostByID(c, postID)
            if post != nil {
              modifyPost = true
           }
         }

        }

        model := struct{
          ModifyPost bool
          Post *Post
          PostID int64          
        } {Post: post, PostID: postID, ModifyPost: modifyPost,}
        tmplPostEdit.Execute(w, model)
        return
    }
    //POST
    //fmt.Fprint(w, r.FormValue("content"));
    savePost(w, r)
}

func handlePostList(w http.ResponseWriter, r *http.Request) {
c := appengine.NewContext(r)
posts, err := getLatestPosts(c, 10)
if err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
        return
}

model := struct {
  Posts []Post
} {
Posts : posts,
}
tmplPostList.Execute(w, model)
}
