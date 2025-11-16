package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

type StatsResponse struct {
	OriginalURL string `json:"original"`
	Clicks      int    `json:"clicks"`
}

func main() {
	fmt.Println("Go backend starting on :8080...")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// POST /api/shorten
	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// TEMP LOGIC — just to test connection with React
		code := "crina123"                         // in real version: random or custom
		shortURL := fmt.Sprintf("http://localhost:8080/%s", code)

		resp := ShortenResponse{
			ShortURL: shortURL,
			Code:     code,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/urls/{code}/stats
	r.Get("/api/urls/{code}/stats", func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		// TEMP STATIC RESPONSE
		resp := StatsResponse{
			OriginalURL: "https://example.com",
			Clicks:      42,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /{code} (redirect)
	r.Get("/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")

		if code == "" || code == "api" {
			http.NotFound(w, r)
			return
		}

		// TEMP redirect
		http.Redirect(w, r, "https://example.com", http.StatusFound)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
