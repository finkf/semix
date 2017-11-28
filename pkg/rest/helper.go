package rest

import (
	"fmt"
	"log"
	"net/http"
)

// WithLogging wraps a HandlerFunc and logs the handling of the request.
func WithLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handling request for [%s] %s", r.Method, r.RequestURI)
		f(w, r)
		log.Printf("handled request for [%s] %s", r.Method, r.RequestURI)
	}
}

// WithPost wraps a HandlerFunc and checks if the request method is POST.
// If not, an error is logged and the handler function is not called.
func WithPost(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("invalid request method: %v", r.Method)
			http.Error(w, fmt.Sprintf("invalid request method: %s", r.Method), http.StatusBadRequest)
			return
		}
		f(w, r)
	}
}

// WithGet wraps a HandlerFunc and checks if the request method is GET.
// If not, an error is logged and the handler function is not called.
func WithGet(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("invalid request method: %v", r.Method)
			http.Error(w, fmt.Sprintf("invalid request method: %s", r.Method), http.StatusBadRequest)
			return
		}
		f(w, r)
	}
}

// WithGetOrPost wraps a HandlerFunc and checks if the request method is GET or POST.
// If not, an error is logged and the handler function is not called.
func WithGetOrPost(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			log.Printf("invalid request method: %v", r.Method)
			http.Error(w, fmt.Sprintf("invalid request method: %s", r.Method), http.StatusBadRequest)
			return
		}
		f(w, r)
	}
}
