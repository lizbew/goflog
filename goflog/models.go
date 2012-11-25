package goflog

import (
    "appengine/datastore"
    "html/template"
    "strconv"
    "time"
)

const (
    TaxonomyCategory     string = "category"
    TaxonomyLinkCategory string = "link_category"
    TaxonomyPostTag      string = "post_tag"
)

type Blog struct {
    Title string
}

type Term struct {
    ID          int64 `datastore:"-"`
    Name        string
    Taxonomy    string
    Description string
    Count       int
}

func NewTerm(id int64, name, taxonomy string) *Term {
    //term := make(Term)
    /* var term Term
    term.Name = name
    term.Taxonomy = taxonomy
    term.Count = 0 */
    term := Term{
        ID:       id,
        Name:     name,
        Taxonomy: taxonomy,
        Count:    0,
    }
    return &term
}

func (term *Term) IsOfCategory() bool {
    return TaxonomyCategory == term.Taxonomy
}

func (term *Term) IsOfLinkCategory() bool {
    return TaxonomyLinkCategory == term.Taxonomy
}

func (term *Term) IsOfPostTag() bool {
    return TaxonomyPostTag == term.Taxonomy
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
    ID        int64
    Title     string
    Content   string
    Published bool
    Author    *datastore.Key
    Created   time.Time
    Modified  time.Time
    AuthorObj User `datastore:"-"`
    Category  string
    Tags      []string
    //Categories []Term `datastore:"-"`
    //Tags       []Term `datastore:"-"`
}

func (p *Post) HTMLContent() template.HTML {
    return template.HTML(p.Content)
}

func (p *Post) DispCreatedTime() string {
    return p.Created.Format("Jan 02, 2006Z08:00")[0:12]
}

func (p *Post) GetPermalink() string {
    return blog["siteurl"] + "/post?id=" + strconv.Itoa(int(p.ID))
}

func (p *Post) HaveComments() bool {
    return false
}

func (p *Post) CommentsCount() int {
    return 0
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
