package goflog

// Read below for gist API docs
// http://developer.github.com/v3/#cross-origin-resource-sharing
// https://github.com/google/go-github

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	GIST_API_HOST     = "https://api.github.com"
	GIST_DEFAULT_FILE = "index.html"
)

var (
	ErrInvalidGistID   = errors.New("Invalid Gist ID")
	ErrNotFoundGist    = errors.New("Not Found Gist")
	ErrNotFoundInStore = errors.New("Not Found in datastore")
)

type GistCopy struct {
	ID           string
	Description  string
	HTMLURL      string
	FilesContent []byte `datastore:",noindex"`
	CopyAt       time.Time
}

// Copyied from https://github.com/google/go-github/blob/master/github/gists.go
// GistFile represents a file on a gist.
type GistFile struct {
	Size     int    `json:"size,omitempty"`
	Filename string `json:"filename,omitempty"`
	RawURL   string `json:"raw_url,omitempty"`
	Content  string `json:"content,omitempty"`
	Type     string `json:"type,omitempty"`
	Language string `json:"language,omitempty"`
}

func (g *GistCopy) UnmarshalFiles() (map[string]GistFile, error) {
	files := make(map[string]GistFile)
	err := json.Unmarshal(g.FilesContent, &files)
	return files, err
}

func init() {
	http.HandleFunc("/g/", handleGist)
	http.HandleFunc("/admin/gist/", handleAdminGist)
	http.HandleFunc("/admin/gist/edit", handleAdminEditGist)
}

func NewGistCopy(ID string) *GistCopy {
	return &GistCopy{
		ID: ID,
	}
}

func CreateGistCopyStoreKey(c appengine.Context, gistID string) *datastore.Key {
	return datastore.NewKey(c, "GistCopy", gistID, 0, nil)
}

func SaveGistCopy(c appengine.Context, gist *GistCopy) error {
	_, err := datastore.Put(c, CreateGistCopyStoreKey(c, gist.ID), gist)
	if err != nil {
		c.Errorf("Failed to Save GistCopy %v: %v", gist.ID, err)
	}
	return err
}

func GetGistCopyList(c appengine.Context) ([]GistCopy, error) {
	var gistCopyList []GistCopy
	_, err := datastore.NewQuery("GistCopy").GetAll(c, &gistCopyList)
	return gistCopyList, err
}

func FindGistCopyByID(c appengine.Context, id string, gist *GistCopy) error {
	k := CreateGistCopyStoreKey(c, id)
	if err := datastore.Get(c, k, gist); err != nil {
		c.Warningf("Error when get GistCopy: %v", err)
		return ErrNotFoundInStore
	}
	return nil
}

func DeleteGistCopyByID(c appengine.Context, gistID string) error {
	err := datastore.Delete(c, CreateGistCopyStoreKey(c, gistID))
	if err != nil {
		c.Errorf("Erorr when delete GistCopy %v: %v", gistID, err)
	}
	return err
}

func fetchGist(c appengine.Context, gistID string, gist *GistCopy) error {
	if gistID == "" {
		return ErrInvalidGistID
	}

	var gistUrl = GIST_API_HOST + "/gists/" + gistID

	c.Debugf("Start to GET %v", gistUrl)
	client := urlfetch.Client(c)
	resp, err := client.Get(gistUrl)
	if err != nil {
		c.Errorf("Erorr when GET %v, %v", gistUrl, err)
		return err
	}
	defer resp.Body.Close()

	gistJson := make(map[string]interface{}, 0)
	if err := json.NewDecoder(resp.Body).Decode(&gistJson); err != nil {
		c.Errorf("Error when decode json: %v", err)
		return err
	}

	if msg, ok := gistJson["message"]; ok && msg == "Not Found" {
		c.Warningf("Not found gist %v", gistID)
		return ErrNotFoundGist
	}

	gist.ID = gistID
	gist.Description = gistJson["description"].(string)
	gist.HTMLURL = gistJson["html_url"].(string)
	gist.FilesContent, _ = json.Marshal(gistJson["files"])
	gist.CopyAt = time.Now()
	return nil
}

func handleGist(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	var gistID string
	var filename string
	cleanPath := path.Clean(r.URL.Path)
	es := strings.Split(cleanPath[1:len(cleanPath)], "/")
	if len(es) >= 2 {
		gistID = es[1]
		if len(es) > 2 {
			filename = es[2]
		}
	} else {
		serveNotFound(w, r)
		return
	}

	if gistfile, err := getGistFile(c, gistID, filename); err == nil {
		w.Header().Set("Content-Type", gistfile.Type)
		io.WriteString(w, gistfile.Content)
		return
	}
	serveNotFound(w, r)
	return
}

func getGistFile(c appengine.Context, gistID string, filename string) (*GistFile, error) {
	var gist GistCopy
	if err := FindGistCopyByID(c, gistID, &gist); err != nil {
		c.Errorf("Error when FindGistCopyByID %v: %v", gistID, err)
		return nil, err
	}

	files, err := gist.UnmarshalFiles()
	if err != nil {
		return nil, err
	}

	if filename == "" {
		filename = GIST_DEFAULT_FILE
	}

	if gistFile, ok := files[filename]; ok {
		return &gistFile, nil
	}
	return nil, ErrNotFoundInStore
}

func handleAdminGist(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	renderContext := make(map[string]interface{})
	renderContext["gist_list"], _ = GetGistCopyList(c)
	if err := executeTemplate(w, "adminGist", http.StatusOK, renderContext); err != nil {
		c.Errorf("%v", err)
	}
}

func handleAdminEditGist(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	action := r.FormValue("action")
	gistID := r.FormValue("gistid")

	resp := make(map[string]interface{}, 0)
	if action == "" || gistID == "" {
		resp["status"] = "failed"
		resp["message"] = "Invalid params"
		responseJson(w, http.StatusBadRequest, resp)
		return
	}

	var code = http.StatusOK
	if action == "add" || action == "refresh" {
		var gist GistCopy

		if err := fetchGist(c, gistID, &gist); err != nil {
			resp["status"] = "failed"
			resp["message"] = "Failed to Fetch gist"
			code = http.StatusInternalServerError
		} else {
			if err := SaveGistCopy(c, &gist); err != nil {
				resp["status"] = "failed"
				resp["message"] = "Failed to Save GistCopy"
				code = http.StatusInternalServerError
			} else {
				resp["status"] = "success"
			}

		}
	}
	if action == "delete" {
		DeleteGistCopyByID(c, gistID)
		resp["status"] = "success"
	}

	responseJson(w, code, resp)
}

func responseJson(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "{}")
	}
}
