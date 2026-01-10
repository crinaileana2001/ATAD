package dtos

type ShortenRequest struct {
	URL        string  `json:"url"`
	CustomCode string  `json:"custom_code"`
	ExpiresAt  *string `json:"expires_at"` 
	WantQR     bool    `json:"want_qr"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
	QRBase64 string `json:"qr_base64,omitempty"`
}
