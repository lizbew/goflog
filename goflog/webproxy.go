package goflog

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrProxyNoURLParam = errors.New("No param url")
)

func writeProxyResponse(c appengine.Context, w http.ResponseWriter, resp *http.Response) {
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
		c.Errorf("Error when write proxy response %v", err)
		panic(err)
	}
	w.Write(result)
}

func DecodeProxyUrl(encodedURL string) (weburl string, reterr error) {
	if encodedURL != "" {
		if strings.HasPrefix(encodedURL, "http:") || strings.HasPrefix(encodedURL, "https:") {
			weburl = encodedURL
			return
		}
		if buf, err := base64.URLEncoding.DecodeString(encodedURL); err == nil {
			weburl = string(buf)
			if !strings.HasPrefix(weburl, "http:") && !strings.HasPrefix(weburl, "https:") {
				weburl = "http://" + weburl
			}
			return
		} else {
			reterr = err
		}
	} else {
		reterr = ErrProxyNoURLParam
	}
	return
}

func handleWebProxy(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	weburl, err := DecodeProxyUrl(r.FormValue("url"))
	if err != nil {
		if err != ErrProxyNoURLParam {
			c.Debugf("Error when get param url: %v", err)
		}
		executeTemplate(w, "webproxy", http.StatusOK, nil)
		return
	}
	c.Debugf("Proxy for: %v", weburl)
	fetchUrlToResponse(c, w, weburl)

}

func fetchUrlToResponse(c appengine.Context, w http.ResponseWriter, weburl string) {
	client := urlfetch.Client(c)
	resp, err := client.Get(weburl)
	if err != nil {
		c.Errorf("Error in handleWebProxy when client.Get(weburl), %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Debugf("http GET returned status %v", resp.Status)
	writeProxyResponse(c, w, resp)
}

// proxy of viiflog
func handleBlogProxy(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	target_url := blog["blog_url"]
	targetURL, err := url.Parse(target_url)

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
	outreq.Header.Set("X-Viifly", fmt.Sprintf("%v://%v", r.URL.Scheme, r.URL.Host))

	//resp, err := http.DefaultClient.Do(outreq)
	resp, err := client.Do(outreq)
	if err != nil {
		//panic(err)
		c.Errorf("Error in when client.Do(outreq), %v", err)
		http.Error(w, "Sorry, Internal Error", http.StatusInternalServerError)
		return
	}

	writeProxyResponse(c, w, resp)
}

func init() {
	http.HandleFunc("/blog/", handleBlogProxy)
	http.HandleFunc("/admin/webproxy/", handleWebProxy)
}
