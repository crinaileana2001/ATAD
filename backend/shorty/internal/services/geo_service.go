package services

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type geoCacheItem struct {
	country string
	expires time.Time
}

type GeoService struct {
	mu    sync.Mutex
	cache map[string]geoCacheItem
	ttl   time.Duration
}

func NewGeoService(ttl time.Duration) *GeoService {
	return &GeoService{
		cache: make(map[string]geoCacheItem),
		ttl:   ttl,
	}
}

func (g *GeoService) LookupCountryISO2(ip string) string {
	if ip == "" || isPrivateIP(ip) {
		return ""
	}

	now := time.Now()
	g.mu.Lock()
	if item, ok := g.cache[ip]; ok && now.Before(item.expires) {
		g.mu.Unlock()
		return item.country
	}
	g.mu.Unlock()

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

	g.mu.Lock()
	g.cache[ip] = geoCacheItem{country: country, expires: now.Add(g.ttl)}
	g.mu.Unlock()

	return country
}

func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	if parsed.IsLoopback() {
		return true
	}
	if v4 := parsed.To4(); v4 != nil {
		switch {
		case v4[0] == 10:
			return true
		case v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31:
			return true
		case v4[0] == 192 && v4[1] == 168:
			return true
		case v4[0] == 169 && v4[1] == 254:
			return true
		}
	}
	return false
}
