package main

import (
	"crypto/md5"
	"encoding/json"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

var (
	FaviconBucket []byte = []byte("favicons")
	TitleBucket   []byte = []byte("titles")
)

var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("cache.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(FaviconBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(TitleBucket)
		return err
	})

	if err != nil {
		log.Fatal(err)
	}
}

type cacheFavicon struct {
	Favicon   string    `json:"favicon"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cf cacheFavicon) bytes() []byte {
	bs, _ := json.Marshal(cf)
	return bs
}

type cacheTitle struct {
	Favicon   string    `json:"favicon"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ct cacheTitle) bytes() []byte {
	bs, _ := json.Marshal(ct)
	return bs
}

func readFavicon(host string) (favicon string, exists bool) {
	var value []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(FaviconBucket)
		value = b.Get([]byte(host))
		return nil
	})

	if len(value) == 0 {
		return
	}

	var cfavicon cacheFavicon
	if err := json.Unmarshal(value, &cfavicon); err != nil {
		log.Println(err)
		return
	}

	if cfavicon.UpdatedAt.Add(30 * 24 * time.Hour).Before(time.Now()) {
		return
	}

	favicon = cfavicon.Favicon
	exists = true
	return
}

func readTitle(url string) (title, favicon string, exists bool) {
	var value []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TitleBucket)
		value = b.Get(urlNormalize(url))
		return nil
	})

	if len(value) == 0 {
		return
	}

	var ctitle cacheTitle
	if err := json.Unmarshal(value, &ctitle); err != nil {
		log.Println(err)
		return
	}

	if ctitle.UpdatedAt.Add(30 * 24 * time.Hour).Before(time.Now()) {
		return
	}

	favicon = ctitle.Favicon
	title = ctitle.Title
	exists = true
	return
}

func update(url, favicon, title string) {
	host := getHost(url)
	db.Update(func(tx *bolt.Tx) error {
		if favicon != "" {
			b := tx.Bucket(FaviconBucket)
			err := b.Put([]byte(host), cacheFavicon{
				Favicon:   favicon,
				UpdatedAt: time.Now(),
			}.bytes())

			if err != nil {
				log.Println(err)
			}
		}

		if title != "" || favicon == "" {
			b := tx.Bucket(TitleBucket)
			err := b.Put(urlNormalize(url), cacheTitle{
				Favicon:   favicon,
				Title:     title,
				UpdatedAt: time.Now(),
			}.bytes())

			if err != nil {
				log.Println(err)
			}
		}

		return nil
	})
}

func urlNormalize(key string) []byte {
	hash := md5.Sum([]byte(key))
	return hash[:]
}
