package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	mrand "math/rand"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

var (
	letters                  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	db                       *bolt.DB
	admindb                  *bolt.DB
	rnd                      = mrand.New(mrand.NewSource(time.Now().UnixNano()))
	errShortUrlAlreadyExists = errors.New("shortUrl already exists")
	errShortUrlNotFound      = errors.New("shortUrl not found")
)

func shortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	longUrl := r.FormValue("url")
	if longUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var shortUrl string
	var err error
	for {
		shortUrl = generateShortUrl()
		err = saveUrl(shortUrl, longUrl)
		if err != errShortUrlAlreadyExists {
			break
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Shortened URL: http://%s/%s\n", r.Host, shortUrl)
}

func redirectToLongUrl(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.URL.Path[1:]
	longUrl, err := loadUrl(shortUrl)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.Redirect(w, r, longUrl, http.StatusSeeOther)
}

func generateShortUrl() string {
	b := make([]rune, 7)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}

func saveUrl(shortUrl, longUrl string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		if v := b.Get([]byte(shortUrl)); v != nil {
			return errShortUrlAlreadyExists
		}
		return b.Put([]byte(shortUrl), []byte(longUrl))
	})
}

func loadUrl(shortUrl string) (string, error) {
	var longUrl string
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		v := b.Get([]byte(shortUrl))
		if v == nil {
			return fmt.Errorf("URL not found")
		}
		longUrl = string(v)
		return nil
	})
	return longUrl, err
}

func deleteUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	shortUrl := r.FormValue("shortUrl")
	if shortUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := removeUrl(shortUrl)
	if err != nil {
		if err == errShortUrlNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "URL entry deleted: %s\n", shortUrl)
}

func removeUrl(shortUrl string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		longUrl := b.Get([]byte(shortUrl))
		if longUrl == nil {
			return errShortUrlNotFound
		}
		return b.Delete([]byte(shortUrl))
	})
	return err
}

// If error exists then log.Fatal(error)
func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// creates a new bucket if it doesn't already exist.
// Returns an error if the bucket name is blank, or if the bucket name is too long.
func createBucket(db *bolt.DB, bucket string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
}

func main() {
	var err error
	db, err = bolt.Open("urls.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	logError(err)
	admindb, err = bolt.Open("admin.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	logError(err)
	defer db.Close()
	defer admindb.Close()

	err = createBucket(db, "urls")
	logError(err)
	err = createBucket(admindb, "admin")
	logError(err)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", shortenUrl).Methods("POST")
	r.HandleFunc("/delete", deleteUrl).Methods("DELETE")
	r.HandleFunc("/{shortUrl:[a-zA-Z0-9]+}", redirectToLongUrl).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}

// https://tinyurl.com/mysecretinspiration
