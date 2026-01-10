package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashIP(ip, salt string) string {
	h := sha256.Sum256([]byte(ip + "|" + salt))
	return hex.EncodeToString(h[:])
}
