package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	db      *bolt.DB
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
	shortUrl := generateShortUrl()
	err := saveUrl(shortUrl, longUrl)
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
	var shortUrl string
	for {
		b := make([]rune, 7)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		shortUrl = string(b)
		_, err := loadUrl(shortUrl)
		if err != nil {
			break
		}
	}
	return shortUrl
}

func saveUrl(shortUrl, longUrl string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
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

func main() {
	var err error
	db, err = bolt.Open("urls.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("urls"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/shorten", shortenUrl)
	http.HandleFunc("/", redirectToLongUrl)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// curl -X POST http://localhost:8080/shorten -d "url=https://www.example.com"
