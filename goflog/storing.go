package goflog

import (
    "appengine"
    "appengine/datastore"
    "appengine/memcache"
    //"net/http"
)

// func queryByKey()

func getLatestPosts(c appengine.Context, count int) ([]Post, error) {
    q := datastore.NewQuery("Post").Order("-Created").Limit(count)
    posts := make([]Post, 0, 10)
    if _, err := q.GetAll(c, &posts); err != nil {
        return nil, err
    }

    //var au *User;
    for i := range posts {
        if posts[i].Author != nil {
            key := "fd"
            item0, err := memcache.Get(c, key)
            /*if err != nil && err != memcache.ErrCacheMiss {
                return err
            }*/
            if err == nil {
                posts[i].AuthorObj = item0.Object.(User)
            } else {
                //au := User{}
                au := new(User)
                if err := datastore.Get(c, posts[i].Author, au); err == nil {
                    posts[i].AuthorObj = *au
                    item1 := &memcache.Item{
                        Key: key,
                        //Value: []byte("bar"),
                        Object: *au,
                    }

                    memcache.Set(c, item1)
                }
            }

            //au = &make(User)

        }
    }
    return posts, nil

}
