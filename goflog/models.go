package goflog

import (
    "time"
    "appengine/datastore"
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
    Title    string
    Content  string
    Author   *datastore.Key
    Created  time.Time
    Modified time.Time
}

type Option struct {
    Name  string
    Value string
}
