package utils

import (
	"testing"
)

func TestGenerateActivationKey(t *testing.T) {
	key1, err := GenerateActivationKey()
	if err != nil {
		t.Fatalf("GenerateActivationKey() error = %v", err)
	}

	if len(key1) != 8 {
		t.Errorf("GenerateActivationKey() length = %d, want 8", len(key1))
	}

	// Should be alphanumeric
	for _, c := range key1 {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Errorf("GenerateActivationKey() contains invalid character: %c", c)
		}
	}

	// Should be different each time
	key2, err := GenerateActivationKey()
	if err != nil {
		t.Fatalf("GenerateActivationKey() error = %v", err)
	}

	if key1 == key2 {
		t.Error("GenerateActivationKey() produced duplicate keys")
	}
}

func TestGenerateSecurityCode(t *testing.T) {
	code, err := GenerateSecurityCode(6)
	if err != nil {
		t.Fatalf("GenerateSecurityCode() error = %v", err)
	}

	if len(code) != 6 {
		t.Errorf("GenerateSecurityCode(6) length = %d, want 6", len(code))
	}

	// Should be numeric
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("GenerateSecurityCode() contains non-digit: %c", c)
		}
	}
}

func TestValidateTIN(t *testing.T) {
	tests := []struct {
		tin  string
		want bool
	}{
		{"1234567890", true},
		{"ABCD123456", true},
		{"123", false},
		{"12345678901", false},
		{"123-456-789", true}, // Should clean and validate
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tin, func(t *testing.T) {
			if got := ValidateTIN(tt.tin); got != tt.want {
				t.Errorf("ValidateTIN(%q) = %v, want %v", tt.tin, got, tt.want)
			}
		})
	}
}

func TestFormatTIN(t *testing.T) {
	tests := []struct {
		tin  string
		want string
	}{
		{"1234567890", "1234567890"},
		{"abcd123456", "ABCD123456"},
		{"123-456-7890", "1234567890"},
		{"ABC DEF 1234", "ABCDEF1234"},
	}

	for _, tt := range tests {
		t.Run(tt.tin, func(t *testing.T) {
			if got := FormatTIN(tt.tin); got != tt.want {
				t.Errorf("FormatTIN(%q) = %q, want %q", tt.tin, got, tt.want)
			}
		})
	}
}

func TestValidateVATNumber(t *testing.T) {
	tests := []struct {
		vat  string
		want bool
	}{
		{"123456789", true},
		{"12345678", false},
		{"1234567890", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.vat, func(t *testing.T) {
			if got := ValidateVATNumber(tt.vat); got != tt.want {
				t.Errorf("ValidateVATNumber(%q) = %v, want %v", tt.vat, got, tt.want)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		currency string
		want     bool
	}{
		{"USD", true},
		{"ZWL", true},
		{"EUR", true},
		{"GBP", true},
		{"ZAR", true},
		{"BWP", true},
		{"XXX", false},
		{"usd", true}, // Should handle case insensitivity
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			if got := ValidateCurrency(tt.currency); got != tt.want {
				t.Errorf("ValidateCurrency(%q) = %v, want %v", tt.currency, got, tt.want)
			}
		})
	}
}

func TestRoundFloat(t *testing.T) {
	tests := []struct {
		value     float64
		precision int
		want      float64
	}{
		{1.234, 2, 1.23},
		{1.235, 2, 1.24},
		{1.999, 2, 2.00},
		{1.234567, 4, 1.2346},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := RoundFloat(tt.value, tt.precision); got != tt.want {
				t.Errorf("RoundFloat(%v, %d) = %v, want %v", tt.value, tt.precision, got, tt.want)
			}
		})
	}
}

func TestCalculateTaxAmount(t *testing.T) {
	tests := []struct {
		total      float64
		taxPercent float64
		want       float64
	}{
		{115.00, 15.0, 15.00},
		{100.00, 0.0, 0.00},
		{230.00, 15.0, 30.00},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := CalculateTaxAmount(tt.total, tt.taxPercent)
			if got != tt.want {
				t.Errorf("CalculateTaxAmount(%v, %v) = %v, want %v", tt.total, tt.taxPercent, got, tt.want)
			}
		})
	}
}

func TestCalculateTaxFromNet(t *testing.T) {
	tests := []struct {
		netAmount  float64
		taxPercent float64
		want       float64
	}{
		{100.00, 15.0, 15.00},
		{100.00, 0.0, 0.00},
		{200.00, 15.0, 30.00},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := CalculateTaxFromNet(tt.netAmount, tt.taxPercent)
			if got != tt.want {
				t.Errorf("CalculateTaxFromNet(%v, %v) = %v, want %v", tt.netAmount, tt.taxPercent, got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		{"test@example.com", true},
		{"user@domain.co.zw", true},
		{"invalid", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.want {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		phone string
		want  bool
	}{
		{"+263771234567", true},
		{"0771234567", true},
		{"077-123-4567", true},
		{"123", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			if got := ValidatePhoneNumber(tt.phone); got != tt.want {
				t.Errorf("ValidatePhoneNumber(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "orange"}

	if !Contains(slice, "banana") {
		t.Error("Contains() should find 'banana'")
	}

	if Contains(slice, "grape") {
		t.Error("Contains() should not find 'grape'")
	}
}

func TestMinMax(t *testing.T) {
	if got := Min(5, 10); got != 5 {
		t.Errorf("Min(5, 10) = %d, want 5", got)
	}

	if got := Max(5, 10); got != 10 {
		t.Errorf("Max(5, 10) = %d, want 10", got)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value int
		min   int
		max   int
		want  int
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := Clamp(tt.value, tt.min, tt.max); got != tt.want {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
