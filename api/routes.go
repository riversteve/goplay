package api

import (
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	// Protected routes under "/api/v1" with authentication middleware
	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.Use(AuthenticationMiddleware)
	apiV1.HandleFunc("/hello", helloSafe).Methods("GET")
	// Including /delete endpoint in both protected and public routes for easy testing
	apiV1.HandleFunc("/delete", DeleteUrl).Methods("DELETE")

	// Unprotected routes
	r.HandleFunc("/hello", helloUnsafe).Methods("GET")
	r.HandleFunc("/shorten", ShortenUrl).Methods("POST")
	r.HandleFunc("/delete", DeleteUrl).Methods("DELETE")
	r.HandleFunc("/{shortUrl:[a-zA-Z0-9]+}", RedirectToLongUrl).Methods("GET")
}
