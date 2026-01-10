package repositories

import (
	"gorm.io/gorm"

	"shorty/internal/entities"
)

type ClickRepo struct {
	db *gorm.DB
}

func NewClickRepo(db *gorm.DB) *ClickRepo {
	return &ClickRepo{db: db}
}

func (r *ClickRepo) Create(evt *entities.ClickEvent) error {
	return r.db.Create(evt).Error
}

func (r *ClickRepo) CountClicks(urlID uint) (int64, error) {
	var clicks int64
	err := r.db.Model(&entities.ClickEvent{}).
		Where("url_id = ?", urlID).
		Count(&clicks).Error
	return clicks, err
}

func (r *ClickRepo) CountUnique(urlID uint) (int64, error) {
	var unique int64
	err := r.db.Model(&entities.ClickEvent{}).
		Where("url_id = ?", urlID).
		Distinct("ip_hash").
		Count(&unique).Error
	return unique, err
}

type CountryRow struct {
	GeoCountry string
	Count      int64
}

func (r *ClickRepo) CountByCountry(urlID uint) ([]CountryRow, error) {
	var rows []CountryRow
	err := r.db.Model(&entities.ClickEvent{}).
		Select("geo_country, COUNT(*) as count").
		Where("url_id = ? AND geo_country IS NOT NULL AND geo_country != ''", urlID).
		Group("geo_country").
		Scan(&rows).Error
	return rows, err
}
