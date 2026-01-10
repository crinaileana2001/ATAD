package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"shorty/internal/config"
	"shorty/internal/dtos"
	"shorty/internal/entities"
	"shorty/internal/repositories"
	"shorty/internal/services"
	"shorty/internal/utils"
)

type Handlers struct {
	cfg config.Config

	urlRepo   *repositories.URLRepo
	clickRepo *repositories.ClickRepo

	codeSvc *services.CodeService
	qrSvc   services.QRService
	geoSvc  *services.GeoService
}

func NewHandlers(
	cfg config.Config,
	urlRepo *repositories.URLRepo,
	clickRepo *repositories.ClickRepo,
	codeSvc *services.CodeService,
	qrSvc services.QRService,
	geoSvc *services.GeoService,
) *Handlers {
	return &Handlers{
		cfg:       cfg,
		urlRepo:   urlRepo,
		clickRepo: clickRepo,
		codeSvc:   codeSvc,
		qrSvc:     qrSvc,
		geoSvc:    geoSvc,
	}
}

func (h *Handlers) Shorten(w http.ResponseWriter, r *http.Request) {
	var req dtos.ShortenRequest
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

	code := strings.TrimSpace(req.CustomCode)
	if code != "" {
		if !h.codeSvc.IsValidCode(code) {
			http.Error(w, "custom_code must be 6-16 chars, alphanumeric", http.StatusBadRequest)
			return
		}
		exists, err := h.urlRepo.ExistsCode(code)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "custom_code already in use", http.StatusConflict)
			return
		}
	} else {
		var err error
		code, err = h.codeSvc.GenerateUniqueCode(h.cfg.DefaultCode)
		if err != nil {
			http.Error(w, "could not generate code", http.StatusInternalServerError)
			return
		}
	}

	u := entities.URL{
		Code:      code,
		Original:  longURL,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
	}

	if err := h.urlRepo.Create(&u); err != nil {
		if utils.IsUniqueConstraint(err) && req.CustomCode == "" {
			code2, err2 := h.codeSvc.GenerateUniqueCode(h.cfg.DefaultCode)
			if err2 == nil {
				u.Code = code2
				if err3 := h.urlRepo.Create(&u); err3 == nil {
					resp := dtos.ShortenResponse{ShortURL: h.cfg.BaseURL + "/" + u.Code, Code: u.Code}
					if req.WantQR {
						qr, err := h.qrSvc.MakeBase64(resp.ShortURL, 256)
						if err != nil {
							http.Error(w, "could not generate qr", http.StatusInternalServerError)
							return
						}
						resp.QRBase64 = qr
					}
					utils.WriteJSON(w, resp, http.StatusCreated)
					return
				}
			}
		}

		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	resp := dtos.ShortenResponse{
		ShortURL: h.cfg.BaseURL + "/" + u.Code,
		Code:     u.Code,
	}
	if req.WantQR {
		qr, err := h.qrSvc.MakeBase64(resp.ShortURL, 256)
		if err != nil {
			http.Error(w, "could not generate qr", http.StatusInternalServerError)
			return
		}
		resp.QRBase64 = qr
	}

	utils.WriteJSON(w, resp, http.StatusCreated)
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" || code == "api" {
		http.NotFound(w, r)
		return
	}

	u, err := h.urlRepo.GetByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if u.ExpiresAt != nil && time.Now().UTC().After(*u.ExpiresAt) {
		http.Error(w, "link expired", http.StatusGone)
		return
	}

	ip := utils.GetClientIP(r)
	ipHash := utils.HashIP(ip, h.cfg.HashSalt)
	country := h.geoSvc.LookupCountryISO2(ip)

	ref := r.Referer()
	ua := r.UserAgent()

	evt := entities.ClickEvent{
		URLID:      u.ID,
		CreatedAt:  time.Now().UTC(),
		IPHash:     ipHash,
		Referrer:   utils.Truncate(ref, 512),
		UserAgent:  utils.Truncate(ua, 512),
		GeoCountry: country,
	}

	go func() {
		if err := h.clickRepo.Create(&evt); err != nil {
			log.Printf("CLICK INSERT FAILED: %v", err)
		} else {
			log.Printf("CLICK SAVED: url_id=%d ip=%s country=%s", evt.URLID, ip, country)
		}
	}()

	http.Redirect(w, r, u.Original, http.StatusFound)
}

func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	u, err := h.urlRepo.GetByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	clicks, _ := h.clickRepo.CountClicks(u.ID)
	unique, _ := h.clickRepo.CountUnique(u.ID)

	rows, _ := h.clickRepo.CountByCountry(u.ID)
	countries := make(map[string]int64)
	for _, row := range rows {
		countries[row.GeoCountry] = row.Count
	}

	resp := dtos.StatsResponse{
		Original:       u.Original,
		Clicks:         clicks,
		UniqueVisitors: unique,
		ExpiresAt:      u.ExpiresAt,
		Countries:      countries,
	}
	utils.WriteJSON(w, resp, http.StatusOK)
}

func (h *Handlers) ListURLs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.urlRepo.ListWithStats()
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	out := make([]dtos.URLListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, dtos.URLListItem{
			Code:           row.Code,
			ShortURL:       h.cfg.BaseURL + "/" + row.Code,
			Original:       row.Original,
			CreatedAt:      row.CreatedAt,
			ExpiresAt:      row.ExpiresAt,
			Clicks:         row.Clicks,
			UniqueVisitors: row.UniqueVisitors,
		})
	}

	utils.WriteJSON(w, out, http.StatusOK)
}
