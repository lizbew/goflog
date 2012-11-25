package goflog

import (
    "appengine"
    "appengine/datastore"
    "appengine/memcache"
    //"net/http"
    "log"
    "sync"
)

// func queryByKey()

const (
    POST_KEY_KIND   = "Post"
    TERM_KEY_KIND   = "Term"
    ALLOCATE_ID_NUM = 2
)

var (
    PostIDLow       int64 = 0
    PostIDHigh      int64 = 0
    PostIDNext      int64 = 0
    MEMCACHE_CODEC        = memcache.Gob
    mu              sync.Mutex
    termInited      bool = false
    termIDMap       map[int64]*Term
    termCategoryMap map[string][]Term
)

func InitTermMap(c appengine.Context) {
    terms := GetAllTerms(c)
    if termIDMap == nil {
        termIDMap = make(map[int64]*Term)
    }
    if termCategoryMap == nil {
        termCategoryMap = make(map[string][]Term)
    }
    for _, t := range terms {
        termIDMap[t.ID] = &t
        if termCategoryMap[t.Taxonomy] == nil {
            var terms []Term
            terms = append(terms, t)
            termCategoryMap[t.Taxonomy] = terms
        } else {
            terms := termCategoryMap[t.Taxonomy]
            termCategoryMap[t.Taxonomy] = append(terms, t)
        }

    }
    termInited = true
}

func GetTermCategoryMap(c appengine.Context) map[string][]Term {
    if !termInited {
        InitTermMap(c)
    }
    return termCategoryMap
}

func CreateTerm(c appengine.Context, name string, taxonomy string) *Term {
    var term *Term
    if l, _, err := datastore.AllocateIDs(c, TERM_KEY_KIND, nil, 1); err == nil {
        /* term := &make(Term)
           term.ID = l
           term.Name = name
           term.Taxonomy = taxonomy
           term.Count = 0 */
        term := NewTerm(l, name, taxonomy)

        key := datastore.NewKey(c, TERM_KEY_KIND, "", l, nil)
        if _, err = datastore.Put(c, key, term); err != nil {
            c.Errorf("Exception when save new Term", err)
        }
    }

    return term
}

func NewPostKey(c appengine.Context) *datastore.Key {
    newID := NewPostID(c)
    return datastore.NewKey(c, POST_KEY_KIND, "", newID, nil)
}

func CreatePostKey(c appengine.Context, postID int64) *datastore.Key {
    return datastore.NewKey(c, POST_KEY_KIND, "", postID, nil)
}

func NewPostID(c appengine.Context) int64 {
    mu.Lock()
    defer mu.Unlock()
    if PostIDNext >= PostIDHigh {
        RequestPostIDNewRange(c)
    }
    newID := PostIDNext
    PostIDNext++
    return newID
}

func RequestPostIDNewRange(c appengine.Context) {
    var err error = nil
    PostIDLow, PostIDHigh, err = datastore.AllocateIDs(c, POST_KEY_KIND, nil,
        ALLOCATE_ID_NUM)
    if err != nil {
        log.Fatal("Failed to allocate new ID range for Post")
        return
    }
    PostIDNext = PostIDLow
}

func GetLatestPosts(c appengine.Context, count int, published bool) ([]Post, error) {
    q := datastore.NewQuery("Post").Order("-Created")
    if published {
        q = q.Filter("Published = ", true)
    }
    if count > 0 {
        q = q.Limit(count)
    }
    posts := make([]Post, 0, 10)
    ks, err := q.GetAll(c, &posts)
    if err != nil {
        return nil, err
    }

    if !termInited {
        InitTermMap(c)
    }
    //var au *User;
    for i := range posts {
        posts[i].ID = ks[i].IntID()

        if posts[i].Author != nil {
            key := posts[i].Author.Encode()
            //item0, err := memcache.Get(c, key)
            /*if err != nil && err != memcache.ErrCacheMiss {
                return err
            }*/
            au0 := new(User)
            if _, err := MEMCACHE_CODEC.Get(c, key, au0); err == nil {
                posts[i].AuthorObj = *au0
            } else {
                log.Print("Error when get from memcache, ", err)
                //au := User{}
                au := new(User)
                if err := datastore.Get(c, posts[i].Author, au); err == nil {
                    posts[i].AuthorObj = *au
                    item1 := &memcache.Item{
                        Key: key,
                        //Value: []byte("bar"),
                        Object:     *au,
                        Expiration: 0,
                    }

                    //memcache.Set(c, item1)
                    if err := MEMCACHE_CODEC.Add(c, item1); err != nil {
                        log.Print("Error when add item to memcache ", err)
                    }
                    /*if err := memcache.Add(c, item1); err != nil {
                       log.Print("Error when add item to memcache ", err)
                    }*/
                    //log.Print("Query datastore and put in memcache", key)
                }
            }

            //au = &make(User)

        }
    }
    return posts, nil

}

func GetPostByCategory(c appengine.Context, category string, published bool) []Post {
    q := datastore.NewQuery("Post").Filter("Category =", category).Order("-Created")
    if published {
        q = q.Filter("Published = ", true)
    }
    var posts []Post
    for it := q.Run(c); ; {
        var post Post
        key, err := it.Next(&post)
        if err == datastore.Done {
            break
        }
        if err != nil {
            c.Errorf("Failed when query Post by category, " + category)
            break
        }
        post.ID = key.IntID()
        posts = append(posts, post)
    }
    return posts

}
func GetPostByID(c appengine.Context, postID int64) *Post {
    postKey := CreatePostKey(c, postID)
    post := new(Post)
    if err := datastore.Get(c, postKey, post); err != nil {
        return nil
    }
    return post
}

func GetAllTerms(c appengine.Context) []Term {
    q := datastore.NewQuery(TERM_KEY_KIND)
    var terms []Term
    for t := q.Run(c); ; {
        var term Term
        k, err := t.Next(&term)
        if err == datastore.Done {
            break
        }
        if err != nil {
            //serveError(c)
            //c.ErrorF("")
            break
        }
        term.ID = k.IntID()
        terms = append(terms, term)
    }
    return terms
}

func GetCategories(c appengine.Context) []Term {
    q := datastore.NewQuery(TERM_KEY_KIND).Filter("Taxonomy =", TaxonomyCategory)
    var categories []Term
    for it := q.Run(c); ; {
        var term Term
        key, err := it.Next(&term)
        if err == datastore.Done {
            break
        }
        if err != nil {
            //serveError(c)
            //c.ErrorF("")
            break
        }
        term.ID = key.IntID()
        categories = append(categories, term)
    }
    return categories
}

func SaveTerm(c appengine.Context, term *Term) error {
    var k *datastore.Key
    if term.ID == 0 {
        k = datastore.NewIncompleteKey(c, TERM_KEY_KIND, nil)
    } else {
        k = datastore.NewKey(c, TERM_KEY_KIND, "", term.ID, nil)
    }
    termInited = false
    if _, err := datastore.Put(c, k, term); err != nil {
        return nil
    }
    return nil
}
