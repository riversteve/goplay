package main

import (
	"strings"
	"testing"

	"github.com/boltdb/bolt"
)

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
