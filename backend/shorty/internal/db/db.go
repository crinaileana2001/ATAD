package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"shorty/internal/entities"
)

func OpenSQLite(path string) (*gorm.DB, error) {
	gdb, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := gdb.AutoMigrate(&entities.URL{}, &entities.ClickEvent{}); err != nil {
		return nil, err
	}

	return gdb, nil
}
