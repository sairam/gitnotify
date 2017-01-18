package gitnotify

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

// CacheWriterIface should be used to set the cache params
// SetCachePath should be set before Write() is ever called
// WriteFromCache will write to the writer directly . returns true when written from cache
// all status writes >= 400 are treated as uncachable
type CacheWriterIface interface {
	SetCachePath(string)
	WriteFromCache() bool
}

// CacheWriter is the struct to store data temporarily
type CacheWriter struct {
	w         http.ResponseWriter
	cache     bool
	usedCache bool
	path      string
	buf       bytes.Buffer
}

func newCacheHandler(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &CacheWriter{
			w: w,
		}
		defer cw.Close()

		h.ServeHTTP(cw, r)
	})
}

// Close saves data to the underlying cache
func (cw *CacheWriter) Close() {
	if cw.usedCache {
		return
		// do nothing
	}
	// save here only if header status was > 200 < 300 or blank
	if cw.cache {
		// log.Println("wrote to cache")
		addToCache(cw.path, string(cw.buf.Bytes()))
	}
}

// WriteFromCache writes to the writer if it could successfully get from cache
func (cw *CacheWriter) WriteFromCache() bool {
	data, success := retrieveFromCache(cw.path)
	if !success || cw.path == "" {
		return success
	}
	cw.usedCache = true
	cw.w.Write([]byte(data))
	return success
}

// SetCachePath sets the cache path and cachable to true
func (cw *CacheWriter) SetCachePath(path string) {
	cw.path = path
	cw.cache = true
}

// Header delegates from the original
func (cw *CacheWriter) Header() http.Header {
	return cw.w.Header()
}

// WriteHeader writes the headers
func (cw *CacheWriter) WriteHeader(code int) {
	if code >= 300 {
		if cw.cache {
			cw.cache = false
		}
	}
	cw.w.WriteHeader(code)
}

func (cw *CacheWriter) Write(b []byte) (int, error) {
	if cw.cache {
		cw.buf.Write(b)
	}
	return cw.w.Write(b)
}

// Caching will be done for 1 day. This needs to be fixed by reading from the cache headers
const cacheBucket = "gitnotify-caches"

// retrieveFromCache removes the item if its after expiry
func retrieveFromCache(cachePath string) (string, bool) {
	db, err := bolt.Open("cacher.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()

	tx, err := db.Begin(true)
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte(cacheBucket))
	timeStr := b.Get([]byte(cachePath + "__"))
	cacheTill, err := strconv.ParseInt(string(timeStr), 10, 64)
	if err != nil {
		return "", false
	}
	if cacheTill > time.Now().Unix() {
		return string(b.Get([]byte(cachePath))), true
	}
	b.Delete([]byte(cachePath + "__"))
	b.Delete([]byte(cachePath))

	if err = tx.Commit(); err != nil {
		return "", false
	}
	return "", false
}

func addToCache(cachePath string, data string) {
	cacheTill := (time.Now().Add(time.Hour)).Unix()
	db, err := bolt.Open("cacher.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()

	tx, err := db.Begin(true)
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte(cacheBucket))
	err = b.Put([]byte(cachePath), []byte(data))
	err = b.Put([]byte(cachePath+"__"), []byte(fmt.Sprintf("%d", cacheTill)))

	if err = tx.Commit(); err != nil {
		return
	}
}

// add another function which will clean up keys periodically - will be invoked via go func()
