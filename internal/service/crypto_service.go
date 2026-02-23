package service

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"fiscalization-api/internal/config"
)

type CryptoService struct {
	caCert     *x509.Certificate
	caKey      crypto.PrivateKey
	serverCert *x509.Certificate
	serverKey  crypto.PrivateKey
	config     config.CryptoConfig
}

func NewCryptoService(cfg config.CryptoConfig) (*CryptoService, error) {
	service := &CryptoService{
		config: cfg,
	}

	// Load CA certificate
	caCertPEM, err := readFile(cfg.CACertificatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	service.caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load server certificate
	serverCertPEM, err := readFile(cfg.CertificatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read server certificate: %w", err)
	}

	block, _ = pem.Decode(serverCertPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode server certificate PEM")
	}

	service.serverCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server certificate: %w", err)
	}

	// Load server private key
	serverKeyPEM, err := readFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read server private key: %w", err)
	}

	service.serverKey, err = parsePrivateKey(serverKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server private key: %w", err)
	}

	return service, nil
}

// IssueCertificate issues a new certificate based on CSR
func (s *CryptoService) IssueCertificate(csrPEM []byte, deviceID int, deviceSerialNo string) (string, []byte, time.Time, error) {
	// Decode PEM
	block, _ := pem.Decode(csrPEM)
	if block == nil {
		return "", nil, time.Time{}, fmt.Errorf("failed to decode CSR PEM")
	}

	// Parse CSR
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return "", nil, time.Time{}, fmt.Errorf("failed to parse CSR: %w", err)
	}

	// Verify CSR signature
	if err := csr.CheckSignature(); err != nil {
		return "", nil, time.Time{}, fmt.Errorf("invalid CSR signature: %w", err)
	}

	// Validate subject fields
	expectedCN := fmt.Sprintf("ZIMRA-%s-%010d", deviceSerialNo, deviceID)
	if csr.Subject.CommonName != expectedCN {
		return "", nil, time.Time{}, fmt.Errorf("invalid CN in CSR, expected: %s, got: %s", expectedCN, csr.Subject.CommonName)
	}

	// Validate optional fields if present
	if len(csr.Subject.Country) > 0 && csr.Subject.Country[0] != "ZW" {
		return "", nil, time.Time{}, fmt.Errorf("invalid country, must be ZW")
	}
	if len(csr.Subject.Organization) > 0 && csr.Subject.Organization[0] != "Zimbabwe Revenue Authority" {
		return "", nil, time.Time{}, fmt.Errorf("invalid organization")
	}

	// Create certificate
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", nil, time.Time{}, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(0, 0, s.config.CertificateValidityDays)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      csr.Subject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Sign certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, s.caCert, csr.PublicKey, s.caKey)
	if err != nil {
		return "", nil, time.Time{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Calculate thumbprint (SHA-1)
	thumbprint := sha1.Sum(certDER)

	return string(certPEM), thumbprint[:], notAfter, nil
}

// VerifyCertificate verifies a client certificate
func (s *CryptoService) VerifyCertificate(cert *x509.Certificate) error {
	// Create certificate pool with CA
	roots := x509.NewCertPool()
	roots.AddCert(s.caCert)

	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// SignData signs data with server private key
func (s *CryptoService) SignData(data []byte) ([]byte, error) {
	// Hash the data
	hash := sha256.Sum256(data)

	// Sign based on key type
	switch key := s.serverKey.(type) {
	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	case *ecdsa.PrivateKey:
		return ecdsa.SignASN1(rand.Reader, key, hash[:])
	default:
		return nil, fmt.Errorf("unsupported key type")
	}
}

// VerifySignature verifies a signature against data and public key
func (s *CryptoService) VerifySignature(data, signature []byte, publicKey crypto.PublicKey) error {
	// Hash the data
	hash := sha256.Sum256(data)

	// Verify based on key type
	switch key := publicKey.(type) {
	case *rsa.PublicKey:
		return rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature)
	case *ecdsa.PublicKey:
		if !ecdsa.VerifyASN1(key, hash[:], signature) {
			return fmt.Errorf("signature verification failed")
		}
		return nil
	default:
		return fmt.Errorf("unsupported key type")
	}
}

// GenerateThumbprint generates SHA-1 thumbprint of a certificate
func (s *CryptoService) GenerateThumbprint(certDER []byte) []byte {
	thumbprint := sha1.Sum(certDER)
	return thumbprint[:]
}

// GetServerCertificate returns the server certificate chain
func (s *CryptoService) GetServerCertificate(thumbprint []byte) ([]string, time.Time, error) {
	// If thumbprint is provided, verify it matches
	if thumbprint != nil {
		serverThumbprint := sha1.Sum(s.serverCert.Raw)
		if !bytesEqual(thumbprint, serverThumbprint[:]) {
			return nil, time.Time{}, fmt.Errorf("thumbprint mismatch")
		}
	}

	// Build certificate chain
	chain := []string{
		string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: s.serverCert.Raw})),
		string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: s.caCert.Raw})),
	}

	return chain, s.serverCert.NotAfter, nil
}

// Helper functions

func readFile(path string) ([]byte, error) {
	// This would use os.ReadFile in production
	// For now, placeholder
	return nil, fmt.Errorf("file reading not implemented")
}

func parsePrivateKey(keyPEM []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	// Try parsing as different key types
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("unsupported private key format")
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// GenerateKeyPair generates a new key pair for testing
func GenerateKeyPair(keyType string) (crypto.PrivateKey, error) {
	switch keyType {
	case "rsa":
		return rsa.GenerateKey(rand.Reader, 2048)
	case "ecdsa":
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}
