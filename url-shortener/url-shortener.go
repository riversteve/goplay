package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

var (
	urls    map[string]string = make(map[string]string)
	letters                   = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
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
	urls[shortUrl] = longUrl
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Shortened URL: http://%s/%s", r.Host, shortUrl)
}

func redirectToLongUrl(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.URL.Path[1:]
	longUrl, ok := urls[shortUrl]
	if !ok {
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
		if _, ok := urls[shortUrl]; !ok {
			break
		}
	}
	return shortUrl
}

func main() {
	http.HandleFunc("/shorten", shortenUrl)
	http.HandleFunc("/", redirectToLongUrl)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
