package services

import (
	"encoding/base64"

	qrcode "github.com/skip2/go-qrcode"
)

type QRService struct{}

func (s QRService) MakeBase64(text string, size int) (string, error) {
	png, err := qrcode.Encode(text, qrcode.Medium, size)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + b64, nil
}
