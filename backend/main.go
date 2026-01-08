package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	baseURL     = "http://localhost:8080"
	hashSalt    = "change-me-in-env-later"
	codeMinLen  = 6
	codeMaxLen  = 8
	defaultCode = 7
)

var db *gorm.DB

// -------------------- MODELS --------------------

type URL struct {
	ID        uint       `gorm:"primaryKey"`
	Code      string     `gorm:"uniqueIndex;size:16;not null"`
	Original  string     `gorm:"size:2048;not null"`
	CreatedAt time.Time  `gorm:"not null"`
	ExpiresAt *time.Time `gorm:"index"`
}

type ClickEvent struct {
	ID        uint      `gorm:"primaryKey"`
	URLID     uint      `gorm:"index;not null"`
	CreatedAt time.Time `gorm:"index;not null"`

	IPHash   string `gorm:"size:64;index;not null"`
	Referrer string `gorm:"size:512"`
	UserAgent string `gorm:"size:512"`
	// GeoCountry / GeoCity will be added later
}

// -------------------- DTOs --------------------

type ShortenRequest struct {
	URL        string  `json:"url"`
	CustomCode string  `json:"custom_code"`
	ExpiresAt  *string `json:"expires_at"` // ISO string from React
	WantQR     bool    `json:"want_qr"`    // will be used later
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
	// QRBase64 string `json:"qr_base64,omitempty"` // later
}

type StatsResponse struct {
	Original       string     `json:"original"`
	Clicks         int64      `json:"clicks"`
	UniqueVisitors int64      `json:"unique_visitors"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// -------------------- MAIN --------------------

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("shorty.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	if err := db.AutoMigrate(&URL{}, &ClickEvent{}); err != nil {
		log.Fatal("failed to migrate:", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// to not see "404" anymore
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Go URL shortener is running ✅"))
	})

	r.Post("/api/shorten", handleShorten)
	r.Get("/api/urls/{code}/stats", handleStats)
	r.Get("/{code}", handleRedirect)

	log.Println("Go backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// -------------------- HANDLERS --------------------

func handleShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	longURL := strings.TrimSpace(req.URL)
	if longURL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		http.Error(w, "url must start with http:// or https://", http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			http.Error(w, "expires_at must be ISO/RFC3339 (e.g. 2026-01-08T12:00:00Z)", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	// custom or generated code
	code := strings.TrimSpace(req.CustomCode)
	if code != "" {
		if !isValidCode(code) {
			http.Error(w, "custom_code must be 6-16 chars, alphanumeric", http.StatusBadRequest)
			return
		}
		// check if it already exists
		var exists URL
		if err := db.Where("code = ?", code).First(&exists).Error; err == nil {
			http.Error(w, "custom_code already in use", http.StatusConflict)
			return
		}
	} else {
		// generating code with retry (collision handling)
		var err error
		code, err = generateUniqueCode(defaultCode)
		if err != nil {
			http.Error(w, "could not generate code", http.StatusInternalServerError)
			return
		}
	}

	u := URL{
		Code:      code,
		Original:  longURL,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
	}

	if err := db.Create(&u).Error; err != nil {
		// if custom code collision (rare, but possible if many requests)
		if isUniqueConstraint(err) && req.CustomCode == "" {
			code, err2 := generateUniqueCode(defaultCode)
			if err2 == nil {
				u.Code = code
				if err3 := db.Create(&u).Error; err3 == nil {
					writeJSON(w, ShortenResponse{ShortURL: baseURL + "/" + u.Code, Code: u.Code}, http.StatusCreated)
					return
				}
			}
		}
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ShortenResponse{
		ShortURL: baseURL + "/" + u.Code,
		Code:     u.Code,
	}, http.StatusCreated)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" || code == "api" {
		http.NotFound(w, r)
		return
	}

	var u URL
	if err := db.Where("code = ?", code).First(&u).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	if u.ExpiresAt != nil && time.Now().UTC().After(*u.ExpiresAt) {
		http.Error(w, "link expired", http.StatusGone) // 410
		return
	}

	// log click event async
	ip := getClientIP(r)
	ipHash := hashIP(ip)

	ref := r.Referer()
	ua := r.UserAgent()

	evt := ClickEvent{
		URLID:      u.ID,
		CreatedAt:  time.Now().UTC(),
		IPHash:     ipHash,
		Referrer:   truncate(ref, 512),
		UserAgent:  truncate(ua, 512),
	}

	_ = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&evt).Error

	http.Redirect(w, r, u.Original, http.StatusFound)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	var u URL
	if err := db.Where("code = ?", code).First(&u).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	var clicks int64
	_ = db.Model(&ClickEvent{}).Where("url_id = ?", u.ID).Count(&clicks).Error

	var unique int64
	// distinct ip_hash
	_ = db.Model(&ClickEvent{}).
		Where("url_id = ?", u.ID).
		Distinct("ip_hash").
		Count(&unique).Error

	resp := StatsResponse{
		Original:       u.Original,
		Clicks:         clicks,
		UniqueVisitors: unique,
		ExpiresAt:      u.ExpiresAt,
	}

	writeJSON(w, resp, http.StatusOK)
}

// -------------------- HELPERS --------------------

func writeJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func generateUniqueCode(n int) (string, error) {
	for i := 0; i < 10; i++ { // retry
		code, err := generateCode(n)
		if err != nil {
			return "", err
		}
		var exists URL
		if err := db.Where("code = ?", code).First(&exists).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not find unique code")
}

func generateCode(n int) (string, error) {
	if n < codeMinLen {
		n = codeMinLen
	}
	if n > codeMaxLen {
		n = codeMaxLen
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var b strings.Builder
	for i := 0; i < n; i++ {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[j.Int64()])
	}
	return b.String(), nil
}

func isValidCode(code string) bool {
	if len(code) < 6 || len(code) > 16 {
		return false
	}
	for _, c := range code {
		if !(c >= 'a' && c <= 'z') &&
			!(c >= 'A' && c <= 'Z') &&
			!(c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip + "|" + hashSalt))
	return hex.EncodeToString(h[:])
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func isUniqueConstraint(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
