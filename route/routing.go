package route

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	DB *sqlx.DB
}

func (h Handler) SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(recoveryMiddleware)
	r.HandleFunc("/test", testRedirect).Methods("GET")
	r.HandleFunc("/{url}", h.addShortnedUrl).Methods("PUT")
	r.HandleFunc("/{referenceKey}", h.addShortnedUrl).Methods("GET")
	return r
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recovered from panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func testRedirect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "https://www.google.com")
	w.WriteHeader(http.StatusPermanentRedirect)
}

// Handler function for URL shortening
func (h Handler) addShortnedUrl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawURL, exists := vars["url"]
	if !exists || rawURL == "" {
		http.Error(w, "Missing 'url' in path parameter", http.StatusBadRequest)
		return
	}
	// Validate if the URL is valid
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	// Check if URL is already shortened
	var existingReferenceKey string
	err = h.DB.Get(&existingReferenceKey, "SELECT reference_key FROM url_mapping WHERE actual_url = ?", rawURL)
	if err == nil {
		// URL already exists, return reference key
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("URL shortened: %s", fmt.Sprintf("%s://%s/%s", r.Proto, r.Host, existingReferenceKey))))
		return
	}

	query := "INSERT INTO url_mapping (actual_url, reference_key) VALUES (?, ?)"
	_, err = h.DB.Exec(query, rawURL, uuid.New().String())
	if err != nil {
		http.Error(w, "Failed to insert URL", http.StatusInternalServerError)
		return
	}
	h.updateURLCount(rawURL)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("URL shortened: %s", fmt.Sprintf("%s://%s/%s", r.Proto, r.Host, existingReferenceKey))))
}

func extractDomain(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	domain := strings.TrimPrefix(parsedURL.Hostname(), "www.")
	return domain, nil
}

func (h *Handler) updateURLCount(url string) error {
	domain, err := extractDomain(url)
	if err != nil {
		return err
	}
	var count int
	err = h.DB.Get(&count, "SELECT count FROM url_count WHERE domain_name = ?", domain)
	if err != nil {
		if err == sql.ErrNoRows {
			// Insert new domain if not exists
			_, err = h.DB.Exec("INSERT INTO url_count (domain_name, count) VALUES (?, ?)", domain, 1)
			return err
		}
		return err
	}
	_, err = h.DB.Exec("UPDATE url_count SET count = count + 1 WHERE domain_name = ?", domain)
	return err
}

// Handler function for getting the original URL from the reference key
func (h Handler) getOriginalUrl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	referenceKey, exists := vars["referenceKey"]
	if !exists || referenceKey == "" {
		http.Error(w, "Missing 'referenceKey' in path parameter", http.StatusBadRequest)
		return
	}

	var originalURL string
	err := h.DB.Get(&originalURL, "SELECT actual_url FROM url_mapping WHERE reference_key = ?", referenceKey)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return 404 if the reference key does not exist
			http.Error(w, "URL not found for the given reference key", http.StatusNotFound)
		} else {
			// Return 500 if there's a server-side error
			http.Error(w, fmt.Sprintf("Failed to retrieve URL: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusPermanentRedirect)
}
