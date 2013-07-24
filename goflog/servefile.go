package goflog

import (
    "appengine"
    "appengine/blobstore"
    _ "io"
    "net/http"
    _ "path"
    "strings"
    "strconv"
)


func handleAdminFileList(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    uploadURL, err := blobstore.UploadURL(c, "/admin/file/upload", nil)
    if err != nil {
        serveError(c, w, err)
        return
    }
    w.Header().Set("Content-Type", "text/html")
    renderContext := make(map[string]interface{})
    renderContext["file_list"] = GetFileList(c)
    renderContext["upload_url"] = uploadURL
    err = executeTemplate(w, "adminFile", http.StatusOK, renderContext)
    if err != nil {
        c.Errorf("%v", err)
    }
}

func handleServe(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    //blobstore.Send(w, appengine.BlobKey(r.FormValue("blobKey")))
    es := strings.Split(r.URL.Path, "/")
    //c.Debugf("Request URL: %v", r.URL.Path)
    //c.Debugf("Splited Path: %v", es)
    if r.URL.Path[0] != '/' || len(es) != 4 {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    fileID, err := strconv.Atoi(es[2])
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    
    if file := GetFileByID(c, int64(fileID)); file != nil {
        blobstore.Send(w, file.FileBlob)
        return
    }
    http.Error(w, "", http.StatusNotFound)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    blobs, _, err := blobstore.ParseUpload(r)
    if err != nil {
        serveError(c, w, err)
        return
    }
    file := blobs["file"]
    if len(file) == 0 {
        c.Errorf("no file uploaded")
        http.Redirect(w, r, "/", http.StatusFound) 
        return
    }
    SaveFileBlobKey(c, file[0])
    //http.Redirect(w, r, "/file/?blobKey="+string(file[0].BlobKey), http.StatusFound)
    http.Redirect(w, r, "/admin/file/", http.StatusFound)
}

func init() {
    http.HandleFunc("/file/", handleServe)
    http.HandleFunc("/admin/file/upload", handleUpload)
    http.HandleFunc("/admin/file/", handleAdminFileList)
}

