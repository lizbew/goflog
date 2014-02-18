package goflog

import (
    "appengine"
    "appengine/datastore"
    "encoding/json"
    "io"
    "io/ioutil"
    "net/http"
    "time"
)

const (
    SSMONENTITY_KIND string = "SsmonEntity"
)

type SsmonEntity struct {
    DateKey    string    `json:"dateKey"`
    Host       string    `json:"host"`
    Category   string    `json:"category"`
    Name       string    `json:"name"`
    Passed     bool      `json:"passed"`
    PathKey    string    `json:"pathKey"`
    Content    []byte    `json:"content"`
    CreateDate time.Time `json:"-"`
}

type SsmonHttpData struct {
    Count    int           `json:"count"`
    Entities []SsmonEntity `json:"entities"`
}

func handleStoreSsmon(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    //insertTestSsmonEntity(c)
    defer r.Body.Close()
    if reqBody, err := ioutil.ReadAll(r.Body); err == nil {
        bodyData := new(SsmonHttpData)
        if err = json.Unmarshal(reqBody, bodyData); err == nil {
            now := time.Now()
            if bodyData.Count > 0 {
                for _, ent := range bodyData.Entities {
                    ent.CreateDate = now
                    datastore.Put(c, datastore.NewIncompleteKey(c, SSMONENTITY_KIND, nil), &ent)
                }
                w.Header().Set("Content-Type", "text/plain; charset=utf-8")
                io.WriteString(w, "done")
                return
            }
        }
    }
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    io.WriteString(w, "fail")
}

func insertTestSsmonEntity(c appengine.Context) {
    entity := SsmonEntity{
        DateKey:    "2014-02-18AM",
        Host:       "sfsf",
        Category:   "p2p",
        Name:       "test",
        Passed:     true,
        PathKey:    "/datat/sd",
        Content:    []byte("fdf"),
        CreateDate: time.Now(),
    }
    datastore.Put(c, datastore.NewIncompleteKey(c, SSMONENTITY_KIND, nil), &entity)
}

func findSsmonEntityByDatekey(c appengine.Context, dateKey string) []SsmonEntity {
    q := datastore.NewQuery(SSMONENTITY_KIND).
        Filter("DateKey=", dateKey)
    var entities []SsmonEntity
    q.GetAll(c, &entities)
    return entities
}

func handleFetchSsmon(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    dateKey := r.FormValue("dateKey")
    if dateKey == "" {
        http.Error(w, "Please provide more parameter", http.StatusBadRequest)
        return
    }

    entities := findSsmonEntityByDatekey(c, dateKey)
    httpData := SsmonHttpData{
        Count:    len(entities),
        Entities: entities,
    }
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    enc := json.NewEncoder(w)
    if err := enc.Encode(httpData); err != nil {
        http.Error(w, "Error", http.StatusInternalServerError)
    }

}

func init() {
    http.HandleFunc("/ssmon_put", handleStoreSsmon)
    http.HandleFunc("/ssmon_get", handleFetchSsmon)
}
