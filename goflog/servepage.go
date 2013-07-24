package goflog

import (
    "appengine"
    "appengine/datastore"
    "time"
    "net/http"
    "strconv"
    "errors"
    "path"
    "io"
    "strings"
)

type Page struct {
    ID int64 `datastore:"-"`
    Title string
    Content []byte
    CreatedTime time.Time
    Layout string
}

type PageAttachment struct {
    ID int64 `datastore:"-"`
    FileID int64
    Description string
}

func SavePage(c appengine.Context, id int64, page *Page) int64{
    var k *datastore.Key = nil
    var savedPage Page
    if id > 0 {
        k = datastore.NewKey(c, "Page", "", id, nil)
    }
    err := datastore.Get(c, k, &savedPage)
    if err == nil {
        //c.Debugf("SavePage, err is nil")
        savedPage.Title = page.Title
        savedPage.Content = page.Content
    } else {
        //c.Debugf("SavePage, err is not nil: %v", err)
        k = datastore.NewIncompleteKey(c, "Page", nil)
        savedPage = *page
        savedPage.CreatedTime = time.Now()
    }

    if k,err = datastore.Put(c, k, page); err != nil {
        c.Errorf("Error when save Page: %v", err)
        return -1
    }
    return k.IntID()
}

func FindPageByID(c appengine.Context, id int64, page *Page) error{
    k := datastore.NewKey(c, "Page", "", id, nil)
    if err := datastore.Get(c, k, page); err != nil {
        c.Warningf("Error when get Page: %v", err)
        return errors.New("Not Found")
    }
    page.ID = id
    return nil
}

func GetPageList(c appengine.Context) []Page {
    var pageList []Page
    q := datastore.NewQuery("Page")
    for it := q.Run(c);; {
        var p Page
        k, err := it.Next(&p)
        if err == datastore.Done {
            break
        }
        p.ID = k.IntID()
        pageList = append(pageList, p)
    }
    return pageList
}


func handleAdminPage(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    renderContext := make(map[string]interface{})
    renderContext["page_list"] = GetPageList(c)
	err := executeTemplate(w, "adminPage", http.StatusOK, renderContext)
	
    if err != nil {
        c.Errorf("%v", err)
    }
}

func handleAdminEditPage(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    renderContext := make(map[string]interface{})

    var page_id int64 = -1
    if r.FormValue("page-id") != "" {
        if i, err := strconv.Atoi(r.FormValue("page-id")); err == nil {
           page_id = int64(i) 
        }
    }
    var page Page
    if r.Method == "POST" {
        page.Title = r.FormValue("page-title")
        page.Content = []byte(r.FormValue("page-content"))
        page_id = SavePage(c, page_id, &page)
    } else if page_id > 0 {
        if err := FindPageByID(c, page_id, &page); err != nil {
            page_id = -1
            c.Warningf("Not Found Page when edit: %v", page_id)
        }
    }
    if page_id > 0 {
        renderContext["pageID"] = page_id
    }
    renderContext["pageTitle"] = page.Title
    renderContext["pageContent"] = string(page.Content)

	err := executeTemplate(w, "adminPageEdit", http.StatusOK, renderContext)
    if err != nil {
        c.Errorf("%v", err)
    }
}

func handlePage(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    cleanPath := path.Clean(r.URL.Path)
    es := strings.Split(cleanPath[1:len(cleanPath)], "/")

    var page_id int64 = -1
    var page Page
    if len(es) >= 2 {
        if i,err := strconv.Atoi(es[1]); err == nil {
            page_id = int64(i)
            
        }
    }
    if page_id > 0 {
        if err := FindPageByID(c, page_id, &page); err != nil {
            serveNotFound(w, r)        
            return 
        }
    } else {
        serveNotFound(w, r)        
        return 
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    io.WriteString(w, string(page.Content))
}

func init() {
    http.HandleFunc("/page/", handlePage)
    http.HandleFunc("/admin/page/", handleAdminPage)
    http.HandleFunc("/admin/page/edit", handleAdminEditPage)
}
