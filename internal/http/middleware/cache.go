package middleware

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// ETag generates an ETag from data
func generateETag(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(jsonData)
	return `"` + base64.StdEncoding.EncodeToString(hash[:])[:16] + `"`
}

// CacheControl adds cache control headers to responses
func CacheControl(maxAge int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set cache control header
			w.Header().Set("Cache-Control", "private, max-age="+string(rune(maxAge)))
			
			next.ServeHTTP(w, r)
		})
	}
}

// ETagMiddleware handles ETag-based conditional requests
func ETagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Check if client sent If-None-Match header
		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch != "" {
			// Store the ETag we'll generate in the response
			// We'll need to intercept the response to generate ETag
			// For now, just pass through - handlers will set ETag manually
		}

		next.ServeHTTP(w, r)
	})
}

// SetETag sets ETag header and checks If-None-Match for 304 response
func SetETag(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	etag := generateETag(data)
	w.Header().Set("ETag", etag)

	// Check if client has matching ETag
	ifNoneMatch := r.Header.Get("If-None-Match")
	if ifNoneMatch == etag {
		w.WriteHeader(http.StatusNotModified)
		return true // Response already sent
	}

	return false // Continue with normal response
}





