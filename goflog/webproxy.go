package goflog

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var prxyPage = `
<html>
<head>
<title>Simple Proxy</title>
</head>
<body>
<div>
<input id="weburl" type="text" name="weburl">
<input type="button" value="Go" onclick="openFor();return false;">
</div>
<script type="text/javascript">
function openFor() {
var l = document.getElementById("weburl").value;
window.location.href="/admin/webproxy/?url="+ encodeURIComponent(l);
}
</script>
</body>
</html>
`

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

func handleWebProxy(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	weburl := r.FormValue("url")
	if weburl == "" {
		w.Header().Set("Content-Type", "text/html;charset=utf-8")
		//io.WriteString(w, prxyPage)
		fmt.Fprint(w, prxyPage)
		return
	}
	c.Debugf("Param url is: %v", weburl)

	if !strings.HasPrefix(weburl, "http:") && !strings.HasPrefix(weburl, "https:") {
		weburl = "http://" + weburl
		var newURL = r.URL
		q := newURL.Query()
		q.Set("url", weburl)
		newURL.RawQuery = q.Encode()

		http.Redirect(w, r, newURL.String(), http.StatusFound)
		return
	}
	c.Debugf("Proxy for: %v", weburl)

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
	outreq.Header.Set("X-Viifly", "http://blog.viifly.com")

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
