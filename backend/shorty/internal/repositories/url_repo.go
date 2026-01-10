package repositories

import (
	"time"

	"gorm.io/gorm"

	"shorty/internal/entities"
)

type URLRepo struct {
	db *gorm.DB
}

func NewURLRepo(db *gorm.DB) *URLRepo {
	return &URLRepo{db: db}
}

func (r *URLRepo) Create(u *entities.URL) error {
	return r.db.Create(u).Error
}

func (r *URLRepo) GetByCode(code string) (*entities.URL, error) {
	var u entities.URL
	if err := r.db.Where("code = ?", code).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *URLRepo) ExistsCode(code string) (bool, error) {
	var u entities.URL
	err := r.db.Select("id").Where("code = ?", code).First(&u).Error
	if err == nil {
		return true, nil
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return false, err
}

type ListRow struct {
	Code           string
	Original       string
	CreatedAt      time.Time
	ExpiresAt      *time.Time
	Clicks         int64
	UniqueVisitors int64
}

func (r *URLRepo) ListWithStats() ([]ListRow, error) {
	var rows []ListRow
	err := r.db.Table("urls").
		Select(`
			urls.code,
			urls.original,
			urls.created_at,
			urls.expires_at,
			COUNT(click_events.id) AS clicks,
			COUNT(DISTINCT click_events.ip_hash) AS unique_visitors
		`).
		Joins("LEFT JOIN click_events ON click_events.url_id = urls.id").
		Group("urls.id").
		Order("urls.created_at DESC").
		Scan(&rows).Error

	return rows, err
}
