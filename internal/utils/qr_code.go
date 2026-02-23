package utils

import (
	"github.com/skip2/go-qrcode"
)

// GenerateQRCodeImage generates a QR code image from data string
func GenerateQRCodeImage(data string, size int) ([]byte, error) {
	// Generate QR code with medium error correction
	return qrcode.Encode(data, qrcode.Medium, size)
}

// GenerateQRCodePNG generates a QR code PNG image
func GenerateQRCodePNG(data string, size int) ([]byte, error) {
	var q *qrcode.QRCode
	q, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	
	return q.PNG(size)
}

// GenerateQRCodeSVG generates a QR code SVG
func GenerateQRCodeSVG(data string, size int) (string, error) {
	q, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", err
	}
	
	return q.ToString(false), nil
}
