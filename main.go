package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/riversteve/goplay/api"
)

var (
	db      *bolt.DB
	admindb *bolt.DB
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

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
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
	api.RegisterRoutes(r)
	r.HandleFunc("/", serveIndex).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))

	for i := 0; i < 10; i++ {
		apiKey, err := api.GenerateAPIKey()
		if err != nil {
			fmt.Println("Error generating API key:", err)
			return
		}

		fmt.Println("Generated API key:", apiKey)
	}
}
