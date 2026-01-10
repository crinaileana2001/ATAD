package config

import "os"

type Config struct {
	BaseURL  string
	HashSalt string

	CodeMinLen  int
	CodeMaxLen  int
	DefaultCode int
}

func Load() Config {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	hashSalt := os.Getenv("HASH_SALT")
	if hashSalt == "" {
		hashSalt = "change-me-in-env-later"
	}

	return Config{
		BaseURL:     baseURL,
		HashSalt:    hashSalt,
		CodeMinLen:  6,
		CodeMaxLen:  8,
		DefaultCode: 7,
	}
}
