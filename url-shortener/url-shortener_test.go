package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

var cleanupOnce sync.Once

func TestMain(m *testing.M) {
	// Set up the database for testing
	testDB, err := bolt.Open("test_urls.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}

	err = testDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("urls"))
		return err
	})
	if err != nil {
		panic(err)
	}

	// Temporarily set the global 'db' variable to the test database
	origDB := db
	db = testDB

	// Run tests
	exitCode := m.Run()

	// Restore the original 'db' variable and clean up the test database
	db = origDB
	cleanupOnce.Do(func() {
		testDB.Close()
		os.Remove("test_urls.db")
	})

	// Exit with the test result exit code
	os.Exit(exitCode)
}

func TestGenerateShortUrl(t *testing.T) {
	shortUrl := generateShortUrl()
	if len(shortUrl) != 7 {
		t.Errorf("Short URL should have 7 characters but has %d", len(shortUrl))
	}

	for _, char := range shortUrl {
		if !strings.ContainsRune(string(letters), char) {
			t.Errorf("Unexpected character %q in short URL", char)
		}
	}
}

func TestSaveAndLoadUrl(t *testing.T) {
	testShortUrl := "test123"
	testLongUrl := "https://example.com"

	err := saveUrl(testShortUrl, testLongUrl)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	longUrl, err := loadUrl(testShortUrl)
	if err != nil {
		t.Fatalf("Failed to load URL: %v", err)
	}

	if longUrl != testLongUrl {
		t.Errorf("Expected long URL %q but got %q", testLongUrl, longUrl)
	}

	// Clean up the test entry
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		return b.Delete([]byte(testShortUrl))
	})
	if err != nil {
		t.Errorf("Failed to clean up test entry: %v", err)
	}
}

func TestSaveLoadAndDeleteUrl(t *testing.T) {
	testShortUrl := "test123"
	testLongUrl := "https://example.com"

	err := saveUrl(testShortUrl, testLongUrl)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	longUrl, err := loadUrl(testShortUrl)
	if err != nil {
		t.Fatalf("Failed to load URL: %v", err)
	}

	if longUrl != testLongUrl {
		t.Errorf("Expected long URL %q but got %q", testLongUrl, longUrl)
	}

	err = removeUrl(testShortUrl)
	if err != nil {
		t.Fatalf("Failed to delete URL: %v", err)
	}

	_, err = loadUrl(testShortUrl)
	if err == nil {
		t.Fatalf("URL should not be found after deletion")
	}
}

func TestDeleteUrlHandler(t *testing.T) {
	testShortUrl := "test456"
	testLongUrl := "https://example.org"

	err := saveUrl(testShortUrl, testLongUrl)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	req, err := http.NewRequest("DELETE", "/delete", strings.NewReader("shortUrl="+testShortUrl))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(deleteUrl)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	_, err = loadUrl(testShortUrl)
	if err == nil {
		t.Fatalf("URL should not be found after deletion")
	}
}
