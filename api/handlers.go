package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	mrand "math/rand"

	"github.com/boltdb/bolt"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rnd     = mrand.New(mrand.NewSource(time.Now().UnixNano()))

	errShortUrlAlreadyExists = errors.New("shortUrl already exists")
	errShortUrlNotFound      = errors.New("shortUrl not found")
)

func helloSafe(w http.ResponseWriter, r *http.Request) {
	// Your protected route logic
	fmt.Fprint(w, "Hello! This is a protected endpoint")
}

func helloUnsafe(w http.ResponseWriter, r *http.Request) {
	// Your public route logic
	fmt.Fprint(w, "Hello from the public endpoint!")
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func RedirectToLongUrl(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.URL.Path[1:]
	longUrl, err := loadUrl(shortUrl)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.Redirect(w, r, longUrl, http.StatusSeeOther)
}

func ShortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	longUrl := r.FormValue("url")
	// Fail fast if empty request
	if longUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Add https if it does not exist
	if !strings.HasPrefix(longUrl, "http://") && !strings.HasPrefix(longUrl, "https://") {
		longUrl = "https://" + longUrl
	}
	// Begin URL parsing checks
	parsedUrl, err := url.Parse(longUrl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Check https scheme
	if !(parsedUrl.Scheme == "https") && !(parsedUrl.Scheme == "http") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: Invalid scheme.")
	}
	// Check for and remove localhost variations. Does not cover all variations
	if strings.ToLower(parsedUrl.Hostname()) == "localhost" || parsedUrl.Hostname() == "127.0.0.1" {
		parsedUrl.Host = strings.Replace(parsedUrl.Host, parsedUrl.Host, "", 1)
	}
	// Check if parsedUrl is only a relative path
	if parsedUrl.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: URL must be an absolute URL, not a relative path.")
		return
	}
	// Finally put longUrl back after checks
	longUrl = parsedUrl.String()

	var shortUrl string
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

func DeleteUrl(w http.ResponseWriter, r *http.Request) {
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

func loadUrl(shortUrl string) (string, error) {
	var longUrl string
	err := DB.View(func(tx *bolt.Tx) error {
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

func generateShortUrl() string {
	b := make([]rune, 7)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}

func saveUrl(shortUrl, longUrl string) error {
	return DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		if v := b.Get([]byte(shortUrl)); v != nil {
			return errShortUrlAlreadyExists
		}
		return b.Put([]byte(shortUrl), []byte(longUrl))
	})
}

func removeUrl(shortUrl string) error {
	err := DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		longUrl := b.Get([]byte(shortUrl))
		if longUrl == nil {
			return errShortUrlNotFound
		}
		return b.Delete([]byte(shortUrl))
	})
	return err
}
