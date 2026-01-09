package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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

	"context"
	"sync"

	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	hashSalt    = "change-me-in-env-later"
	codeMinLen  = 6
	codeMaxLen  = 8
	defaultCode = 7
)

var baseURL = os.Getenv("BASE_URL")

func init() {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
}

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
	GeoCountry string `gorm:"size:2;index"` 
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
	QRBase64 string `json:"qr_base64,omitempty"` // PNG base64
}

type StatsResponse struct {
	Original       string     `json:"original"`
	Clicks         int64      `json:"clicks"`
	UniqueVisitors int64      `json:"unique_visitors"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Countries      map[string]int64   `json:"countries,omitempty"` 
}

// -------------------- MAIN --------------------

func main() {
	var err error
	cwd, _ := os.Getwd()
log.Println("CWD:", cwd)
log.Println("DB file:", cwd+`\\shorty.db`)

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
	
	rl := NewRateLimiter(10, 30*time.Minute)
    stop := make(chan struct{})
    go rl.cleanupLoop(stop)


	// to not see "404" anymore
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Go URL shortener is running ✅"))
	})

	r.Route("/api", func(api chi.Router) {
	api.Use(RateLimitMiddleware(rl)) 

	api.Post("/shorten", handleShorten)
	api.Get("/urls/{code}/stats", handleStats)

	
})

// redirect-ul rămâne nelimitat
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

	resp := ShortenResponse{
    ShortURL: baseURL + "/" + u.Code,
    Code:     u.Code,
}

if req.WantQR {
    qr, err := makeQRBase64(resp.ShortURL, 256)
    if err != nil {
        http.Error(w, "could not generate qr", http.StatusInternalServerError)
        return
    }
    resp.QRBase64 = qr
}

writeJSON(w, resp, http.StatusCreated)

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
	country := lookupCountryISO2(ip)


	ref := r.Referer()
	ua := r.UserAgent()

	evt := ClickEvent{
		URLID:      u.ID,
		CreatedAt:  time.Now().UTC(),
		IPHash:     ipHash,
		Referrer:   truncate(ref, 512),
		UserAgent:  truncate(ua, 512),
		GeoCountry: country,

	}

	if err := db.Create(&evt).Error; err != nil {
	log.Printf("CLICK INSERT FAILED: %v", err)
} else {
	log.Printf("CLICK SAVED: url_id=%d ip=%s country=%s", evt.URLID, ip, country)
}

	http.Redirect(w, r, u.Original, http.StatusFound)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// 1. Get URL by code
	var u URL
	if err := db.Where("code = ?", code).First(&u).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	// 2. Total clicks
	var clicks int64
	_ = db.Model(&ClickEvent{}).
		Where("url_id = ?", u.ID).
		Count(&clicks).Error

	// 3. Unique visitors (distinct IP hash)
	var unique int64
	_ = db.Model(&ClickEvent{}).
		Where("url_id = ?", u.ID).
		Distinct("ip_hash").
		Count(&unique).Error

	// 4. Clicks by country (geo analytics)
	type countryRow struct {
		GeoCountry string
		Count      int64
	}

	var rows []countryRow
	_ = db.Model(&ClickEvent{}).
		Select("geo_country, COUNT(*) as count").
		Where("url_id = ? AND geo_country IS NOT NULL AND geo_country != ''", u.ID).
		Group("geo_country").
		Scan(&rows).Error

	countries := make(map[string]int64)
	for _, r := range rows {
		countries[r.GeoCountry] = r.Count
	}

	// 5. Build response
	resp := StatsResponse{
		Original:       u.Original,
		Clicks:         clicks,
		UniqueVisitors: unique,
		ExpiresAt:      u.ExpiresAt,
		Countries:      countries, // NEW
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
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}
	if xr := strings.TrimSpace(r.Header.Get("X-Real-IP")); xr != "" {
		return xr
	}

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

func makeQRBase64(text string, size int) (string, error) {
    png, err := qrcode.Encode(text, qrcode.Medium, size)
    if err != nil {
        return "", err
    }
    // Return as data URL so frontend can display directly in <img src="...">
    b64 := base64.StdEncoding.EncodeToString(png)
    return "data:image/png;base64," + b64, nil
}

// -------------------- GEO (COUNTRY) --------------------

type geoCacheItem struct {
	country string
	expires time.Time
}

var (
	geoMu    sync.Mutex
	geoCache = map[string]geoCacheItem{}
)

func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	// loopback
	if parsed.IsLoopback() {
		return true
	}
	// IPv4 private ranges
	if v4 := parsed.To4(); v4 != nil {
		switch {
		case v4[0] == 10:
			return true
		case v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31:
			return true
		case v4[0] == 192 && v4[1] == 168:
			return true
		case v4[0] == 169 && v4[1] == 254: // link-local
			return true
		}
	}
	return false
}


// calls ipapi.co: https://ipapi.co/<IP>/json/  :contentReference[oaicite:2]{index=2}

func lookupCountryISO2(ip string) string {
	if ip == "" || isPrivateIP(ip) {
		return ""
	}

	// cache (TTL 24h)
	now := time.Now()
	geoMu.Lock()
	if item, ok := geoCache[ip]; ok && now.Before(item.expires) {
		geoMu.Unlock()
		return item.country
	}
	geoMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ipwho.is/"+ip, nil)
	if err != nil {
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var out struct {
		Success     bool   `json:"success"`
		CountryCode string `json:"country_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ""
	}
	if !out.Success {
		return ""
	}

	country := strings.ToUpper(strings.TrimSpace(out.CountryCode))
	if len(country) != 2 {
		country = ""
	}

	geoMu.Lock()
	geoCache[ip] = geoCacheItem{country: country, expires: now.Add(24 * time.Hour)}
	geoMu.Unlock()

	return country
}



// -------------------- RATE LIMITING --------------------

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*tokenBucket
	capacity   float64       // max tokens
	refillRate float64       // tokens per second
	ttl        time.Duration // cleanup idle buckets
}

func NewRateLimiter(maxPerMinute int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*tokenBucket),
		capacity:   float64(maxPerMinute),
		refillRate: float64(maxPerMinute) / 60.0, // per second
		ttl:        ttl,
	}
}

func (rl *RateLimiter) allow(ip string) (allowed bool, retryAfterSec int) {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &tokenBucket{
			tokens:     rl.capacity - 1, // consume 1 token immediately
			lastRefill: now,
		}
		return true, 0
	}

	// refill tokens
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens = minFloat(rl.capacity, b.tokens+elapsed*rl.refillRate)
		b.lastRefill = now
	}

	if b.tokens >= 1 {
		b.tokens -= 1
		return true, 0
	}

	// compute retry-after (seconds until next token)
	need := 1 - b.tokens
	sec := int((need / rl.refillRate) + 0.999) // ceil
	if sec < 1 {
		sec = 1
	}
	return false, sec
}

func (rl *RateLimiter) cleanupLoop(stop <-chan struct{}) {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			rl.cleanup()
		case <-stop:
			return
		}
	}
}

func (rl *RateLimiter) cleanup() {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, b := range rl.buckets {
		if now.Sub(b.lastRefill) > rl.ttl {
			delete(rl.buckets, ip)
		}
	}
}

func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			allowed, retryAfter := rl.allow(ip)
			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				http.Error(w, "rate limit exceeded (max 10 requests/minute)", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
