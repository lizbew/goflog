package goflog

import (
    "appengine"
    "appengine/urlfetch"
    "io"
    "io/ioutil"
    "net/http"
    "fmt"
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
    //fmt.Fprintf(w, "http GET returned status %v", resp.Status)
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
    c.Errorf("Error in handleWebProxy, %v", err)
    panic(err)
  }
  w.Write(result)
}


func init(){
    http.HandleFunc("/admin/webproxy/", handleWebProxy)
}
