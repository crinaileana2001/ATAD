package app

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"shorty/internal/config"
	mid "shorty/internal/middleware"
	"shorty/internal/repositories"
	"shorty/internal/services"
)

type App struct {
	cfg config.Config
	db  *gorm.DB
}

func New(cfg config.Config, db *gorm.DB) *App {
	return &App{cfg: cfg, db: db}
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	rl := mid.NewRateLimiter(10, 30*time.Minute)
	stop := make(chan struct{})
	go rl.CleanupLoop(stop)

	urlRepo := repositories.NewURLRepo(a.db)
	clickRepo := repositories.NewClickRepo(a.db)

	codeSvc := services.NewCodeService(a.cfg, a.db)
	qrSvc := services.QRService{}
	geoSvc := services.NewGeoService(24 * time.Hour)

	h := NewHandlers(a.cfg, urlRepo, clickRepo, codeSvc, qrSvc, geoSvc)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Go URL shortener is running ✅"))
	})

	r.Route("/api", func(api chi.Router) {
		api.With(mid.RateLimitMiddleware(rl)).Post("/shorten", h.Shorten)

		api.Get("/urls", h.ListURLs)
		api.Get("/urls/{code}/stats", h.Stats)
	})

	r.Get("/{code}", h.Redirect)
	return r
}

func (a *App) Run(addr string) error {
	log.Println("Go backend running on", addr)
	return http.ListenAndServe(addr, a.Router())
}
