package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"
	
	"github.com/google/uuid"
)

// GenerateOperationID generates a unique operation ID
func GenerateOperationID() string {
	return uuid.New().String()
}

// GenerateActivationKey generates an 8-character activation key
func GenerateActivationKey() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

// GenerateSecurityCode generates a numeric security code
func GenerateSecurityCode(length int) (string, error) {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

// FormatTIN formats a TIN number
func FormatTIN(tin string) string {
	// Remove any non-alphanumeric characters
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') {
			return r
		}
		return -1
	}, strings.ToUpper(tin))
	
	return cleaned
}

// ValidateTIN validates a TIN format
func ValidateTIN(tin string) bool {
	cleaned := FormatTIN(tin)
	return len(cleaned) == 10
}

// FormatVATNumber formats a VAT number
func FormatVATNumber(vat string) string {
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, vat)
	
	return cleaned
}

// ValidateVATNumber validates a VAT number format
func ValidateVATNumber(vat string) bool {
	cleaned := FormatVATNumber(vat)
	return len(cleaned) == 9
}

// FormatCurrency formats a currency code
func FormatCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

// ValidateCurrency validates currency code
func ValidateCurrency(currency string) bool {
	valid := []string{"USD", "ZWL", "EUR", "GBP", "ZAR", "BWP"}
	formatted := FormatCurrency(currency)
	for _, v := range valid {
		if formatted == v {
			return true
		}
	}
	return false
}

// ParseDateYYYYMMDD parses date in YYYY-MM-DD format
func ParseDateYYYYMMDD(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// ParseDateTimeFull parses datetime in YYYY-MM-DDTHH:mm:ss format
func ParseDateTimeFull(dateTimeStr string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", dateTimeStr)
}

// FormatDateYYYYMMDD formats date to YYYY-MM-DD
func FormatDateYYYYMMDD(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDateTimeFull formats datetime to YYYY-MM-DDTHH:mm:ss
func FormatDateTimeFull(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

// TruncateString truncates string to max length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// NormalizeWhitespace normalizes whitespace in string
func NormalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// SanitizeInput sanitizes user input
func SanitizeInput(input string) string {
	// Remove control characters
	cleaned := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)
	
	return NormalizeWhitespace(cleaned)
}

// EncodeBase64 encodes bytes to base64 string
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decodes base64 string to bytes
func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// RoundFloat rounds float to specified decimal places
func RoundFloat(val float64, precision int) float64 {
	ratio := float64(1)
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}

// FormatMoney formats money amount
func FormatMoney(amount float64, currency string) string {
	return fmt.Sprintf("%s %.2f", currency, amount)
}

// CalculateTaxAmount calculates tax amount from total (tax inclusive)
func CalculateTaxAmount(total float64, taxPercent float64) float64 {
	if taxPercent == 0 {
		return 0
	}
	return RoundFloat(total*taxPercent/(100+taxPercent), 2)
}

// CalculateTaxFromNet calculates tax amount from net (tax exclusive)
func CalculateTaxFromNet(netAmount float64, taxPercent float64) float64 {
	return RoundFloat(netAmount*taxPercent/100, 2)
}

// ValidateEmail validates email format (basic)
func ValidateEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// ValidatePhoneNumber validates phone number format (basic)
func ValidatePhoneNumber(phone string) bool {
	// Remove common separators
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phone)
	
	// Should be at least 10 digits
	digitCount := 0
	for _, r := range cleaned {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}
	
	return digitCount >= 10 && digitCount <= 15
}

// Contains checks if slice contains element
func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// Min returns minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps value between min and max
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// IsWorkingHours checks if current time is within working hours
func IsWorkingHours(t time.Time, startHour, endHour int) bool {
	hour := t.Hour()
	return hour >= startHour && hour < endHour
}

// IsWeekend checks if date is weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// AddBusinessDays adds business days to date
func AddBusinessDays(t time.Time, days int) time.Time {
	result := t
	added := 0
	
	for added < days {
		result = result.AddDate(0, 0, 1)
		if !IsWeekend(result) {
			added++
		}
	}
	
	return result
}
