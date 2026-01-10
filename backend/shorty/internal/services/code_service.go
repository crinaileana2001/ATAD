package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"gorm.io/gorm"

	"shorty/internal/config"
	"shorty/internal/entities"
)

type CodeService struct {
	cfg config.Config
	db  *gorm.DB
}

func NewCodeService(cfg config.Config, db *gorm.DB) *CodeService {
	return &CodeService{cfg: cfg, db: db}
}

func (s *CodeService) IsValidCode(code string) bool {
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

func (s *CodeService) GenerateUniqueCode(n int) (string, error) {
	for i := 0; i < 10; i++ {
		code, err := s.generateCode(n)
		if err != nil {
			return "", err
		}
		var exists entities.URL
		err = s.db.Where("code = ?", code).First(&exists).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not find unique code")
}

func (s *CodeService) generateCode(n int) (string, error) {
	if n < s.cfg.CodeMinLen {
		n = s.cfg.CodeMinLen
	}
	if n > s.cfg.CodeMaxLen {
		n = s.cfg.CodeMaxLen
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
