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
  POST_KEY_KIND = "Post"
  ALLOCATE_ID_NUM = 2
)

var (
  PostIDLow int64 = 0
  PostIDHigh int64 = 0
  PostIDNext int64 = 0
  MEMCACHE_CODEC = memcache.Gob
  mu sync.Mutex
)

func NewPostKey(c appengine.Context) *datastore.Key{
  newID := NewPostID(c)
  return datastore.NewKey(c, POST_KEY_KIND, "", newID, nil)
}

func CreatePostKey(c appengine.Context, postID int64) *datastore.Key{
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
  PostIDLow, PostIDHigh, err  = datastore.AllocateIDs(c, POST_KEY_KIND, nil,
  ALLOCATE_ID_NUM)
  if err != nil {
    log.Fatal("Failed to allocate new ID range for Post")
    return
  }
  PostIDNext = PostIDLow
}

func getLatestPosts(c appengine.Context, count int) ([]Post, error) {
    q := datastore.NewQuery("Post").Order("-Created").Limit(count)
    posts := make([]Post, 0, 10)
    ks, err := q.GetAll(c, &posts);
    if err != nil {
        return nil, err
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
                        Object: *au,
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

func GetPostByID(c appengine.Context, postID int64) *Post {
  postKey := CreatePostKey(c, postID)  
  post := new(Post)
  if err := datastore.Get(c, postKey, post); err != nil {
    return nil
  }
  return post
}
