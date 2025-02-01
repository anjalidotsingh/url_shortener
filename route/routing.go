package route

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(recoveryMiddleware)
	r.HandleFunc("/test", testRedirect).Methods("GET")
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
