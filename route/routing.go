package route

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortenUrl string `json:"shortenUrl"`
}

type domainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}
type Handler struct {
	DB *sqlx.DB
}

func (h Handler) SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(recoveryMiddleware)
	r.HandleFunc("/shorten", h.addShortnedUrl).Methods(http.MethodPost)
	r.HandleFunc("/resolve/{referenceKey}", h.getOriginalUrl).Methods("GET")
	r.HandleFunc("/domain-counts", h.getTopDomains).Methods("GET")
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

// Handler function for URL shortening
func (h Handler) addShortnedUrl(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest

	// Decode the request body into req struct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "Invalid JSON input or missing 'url'", http.StatusBadRequest)
		return
	}

	rawURL := req.URL
	// Validate if the URL is valid
	_, err = url.ParseRequestURI(rawURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	// Check if URL is already shortened
	var existingReferenceKey string
	err = h.DB.Get(&existingReferenceKey, "SELECT reference_key FROM url_mapping WHERE actual_url = ?", rawURL)
	if err == nil {
		response := shortenResponse{
			ShortenUrl: fmt.Sprintf("http://%s/resolve/%s", r.Host, existingReferenceKey),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}
	existingReferenceKey = uuid.New().String()
	query := "INSERT INTO url_mapping (actual_url, reference_key) VALUES (?, ?)"
	_, err = h.DB.Exec(query, rawURL, existingReferenceKey)
	if err != nil {
		http.Error(w, "Failed to insert URL", http.StatusInternalServerError)
		return
	}
	h.updateURLCount(rawURL)
	response := shortenResponse{
		ShortenUrl: fmt.Sprintf("http://%s/resolve/%s", r.Host, existingReferenceKey),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
	w.WriteHeader(http.StatusMovedPermanently)
}

func (h Handler) getTopDomains(w http.ResponseWriter, r *http.Request) {
	var topDomains []domainCount
	err := h.DB.Select(&topDomains, `
		SELECT domain_name AS domain, count 
		FROM url_count 
		ORDER BY count DESC 
		LIMIT 4`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve top domains: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(topDomains)
}
