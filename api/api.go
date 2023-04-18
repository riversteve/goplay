package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

var (
	DB      *bolt.DB
	AdminDB *bolt.DB
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		hashedKey, err := hashAPIKey(apiKey)
		//hashedKey, err := getHashedKeyFromDatabase(hashedKey)
		if err != nil || !checkHashedKeyFromDatabase(hashedKey) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hashAPIKey(apiKey string) (string, error) {
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedKey), nil
}

/*
func validateAPIKey(apiKey string, hashedKey string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(apiKey))
	return err == nil
}
*/

// Generate API key of preset length
// Optional: set API key prefix. Default is BEEF
func GenerateAPIKey(prefix ...string) (string, error) {
	// Set default prefix
	defaultPrefix := "BEEF"
	// Set API key length
	length := 40
	// Check if a custom prefix is provided and within the allowed length
	if len(prefix) > 0 && len(prefix[0]) <= 4 {
		defaultPrefix = prefix[0]
	}

	// Calculate the length of the random part of the API key
	randomPartLength := length - len(defaultPrefix)

	if randomPartLength <= 0 {
		return "", fmt.Errorf("the specified length is too short for the prefix")
	}

	// Generate random bytes of the specified length
	randomBytes := make([]byte, randomPartLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("error generating random bytes: %v", err)
	}

	// Encode the random bytes using base64
	apiKey := base64.URLEncoding.EncodeToString(randomBytes)

	// Remove padding characters from the base64-encoded string
	apiKey = strings.TrimRight(apiKey, "=")

	// Truncate the encoded string to the desired length of the random part
	apiKey = apiKey[:randomPartLength]

	// Prepend prefix to the API key
	apiKey = defaultPrefix + apiKey

	return apiKey, nil
}

func checkHashedKeyFromDatabase(apiKey string) bool {
	err := DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("APIKeys"))
		if bucket == nil {
			return fmt.Errorf("APIKeys bucket not found")
		}
		hashedKeyBytes := bucket.Get([]byte(apiKey))
		if hashedKeyBytes == nil {
			return fmt.Errorf("API key not found")
		}
		return nil
	})
	return err == nil
}
