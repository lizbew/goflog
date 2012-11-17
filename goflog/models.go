package goflog

import (
    "appengine/datastore"
    "html/template"
    "time"
)

type Blog struct {
    Title string
}

type User struct {
    Email      string
    Nicename   string
    Active     bool
    Registered time.Time
}

type Comment struct {
    Content     string
    Author      string
    AuthorEmail string
    AuthorUrl   string
    AuthorIp    string
    Created     time.Time
    LoginUser   User
}

type Post struct {
    Title     string
    Content   string
    Author    *datastore.Key
    Created   time.Time
    Modified  time.Time
    AuthorObj User `datastore:"-"`
}

func (p *Post) HTMLContent() template.HTML {
    return template.HTML(p.Content)
}

func (p *Post) DispCreatedTime() string {
    return p.Created.Format("Jan 02, 2006Z08:00")[0:12]
}

/*func (p *Post) getAuthorDisplay() string {
if p.AuthorObj != nil {
return p.AuthorObj.Nicename
}
//return p.AuthorObj.Nicename
return ""
}*/

/*func (p *Post) Load(c <-chan Property) error {
if err := datastore.LoadStruct(p, c); err != nil {
        return err
    }
return nil;
}*/

type Option struct {
    Name  string
    Value string
}
