package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	mrand "math/rand"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/riversteve/goplay/api"
)

var (
	letters                  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	db                       *bolt.DB
	admindb                  *bolt.DB
	rnd                      = mrand.New(mrand.NewSource(time.Now().UnixNano()))
	errShortUrlAlreadyExists = errors.New("shortUrl already exists")
	errShortUrlNotFound      = errors.New("shortUrl not found")
)

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

func Init() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080" // default when missing
	}
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
	err = createBucket(admindb, "APIKeys")
	logError(err)

	api.DB = db
	api.AdminDB = admindb
	r := mux.NewRouter()

	// Protected routes under "/api/v1" with authentication middleware
	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.Use(api.AuthenticationMiddleware)
	apiV1.HandleFunc("/hello", helloSafe).Methods("GET")
	// Including /delete endpoint in both protected and public routes for easy testing
	apiV1.HandleFunc("/delete", deleteUrl).Methods("DELETE")

	// Public routes without authentication middleware
	r.HandleFunc("/", serveIndex).Methods("GET")
	r.HandleFunc("/hello", helloUnsafe).Methods("GET")
	r.HandleFunc("/shorten", shortenUrl).Methods("POST")
	r.HandleFunc("/delete", deleteUrl).Methods("DELETE")
	r.HandleFunc("/{shortUrl:[a-zA-Z0-9]+}", redirectToLongUrl).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}

// https://tinyurl.com/mysecretinspiration
