# Implementation Guide

## Complete Implementation Roadmap

This guide outlines the step-by-step implementation of the remaining components for a production-ready ZIMRA fiscalization API.

## Phase 1: Core Infrastructure ✓

Already provided in the file structure:
- [x] Project structure
- [x] Database schema
- [x] Models and enums
- [x] Configuration management
- [x] Docker setup

## Phase 2: Repository Layer (Data Access)

### Files to Implement

#### `/internal/repository/device_repository.go`
```go
package repository

import (
    "database/sql"
    "fiscalization-api/internal/models"
    "github.com/jmoiron/sqlx"
)

type DeviceRepository interface {
    Create(device *models.Device) error
    GetByDeviceID(deviceID int) (*models.Device, error)
    GetBySerialNo(serialNo string) (*models.Device, error)
    Update(device *models.Device) error
    VerifyActivationKey(deviceID int, activationKey string) (bool, error)
    UpdateCertificate(deviceID int, cert string, thumbprint []byte, validTill time.Time) error
    IsBlacklisted(modelName, modelVersion string) (bool, error)
}

type deviceRepository struct {
    db *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) DeviceRepository {
    return &deviceRepository{db: db}
}

// Implement all interface methods with proper SQL queries
// Example:
func (r *deviceRepository) GetByDeviceID(deviceID int) (*models.Device, error) {
    var device models.Device
    query := `SELECT * FROM devices WHERE device_id = $1`
    err := r.db.Get(&device, query, deviceID)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &device, err
}
```

#### `/internal/repository/taxpayer_repository.go`
```go
package repository

// TaxpayerRepository interface and implementation
// Methods: Create, GetByID, GetByTIN, Update, GetTaxes
```

#### `/internal/repository/receipt_repository.go`
```go
package repository

// ReceiptRepository interface and implementation
// Methods: Create, GetByID, GetByGlobalNo, GetPreviousReceipt, 
// UpdateValidation, GetMissingReceipts, etc.
```

#### `/internal/repository/fiscal_day_repository.go`
```go
package repository

// FiscalDayRepository interface and implementation
// Methods: Create, GetCurrent, Close, GetCounters, UpdateCounters
```

#### `/internal/repository/user_repository.go`
```go
package repository

// UserRepository interface and implementation
// Methods: Create, GetByUsername, Update, UpdatePassword, etc.
```

## Phase 3: Service Layer (Business Logic)

### Files to Implement

#### `/internal/service/crypto_service.go`
```go
package service

import (
    "crypto"
    "crypto/ecdsa"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/x509"
    "crypto/x509/pkix"
    "encoding/pem"
    "fmt"
    "math/big"
    "time"
)

type CryptoService struct {
    caCert    *x509.Certificate
    caKey     crypto.PrivateKey
    serverCert *x509.Certificate
    serverKey  crypto.PrivateKey
}

func NewCryptoService(cfg CryptoConfig) (*CryptoService, error) {
    // Load CA certificate and key
    // Load server certificate and key
    return &CryptoService{...}, nil
}

// Key methods to implement:
// - IssueCertificate(csr []byte, deviceName string) (string, error)
// - VerifyCertificate(cert []byte) error
// - SignData(data []byte) ([]byte, error)
// - VerifySignature(data, signature, cert []byte) error
// - GenerateThumbprint(cert []byte) ([]byte, error)
```

#### `/internal/service/validation_service.go`
```go
package service

// ValidationService implements all receipt validation rules
// Methods for each validation code (RCPT010, RCPT011, etc.)
```

#### `/internal/service/device_service.go`
```go
package service

// DeviceService handles device registration, configuration
```

#### `/internal/service/receipt_service.go`
```go
package service

// ReceiptService handles receipt submission and validation
```

#### `/internal/service/fiscal_day_service.go`
```go
package service

// FiscalDayService handles fiscal day operations
```

## Phase 4: Handler Layer (HTTP Controllers)

### Files to Implement

#### `/internal/handlers/device_handler.go`
```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "fiscalization-api/internal/models"
    "fiscalization-api/internal/service"
)

type DeviceHandler struct {
    deviceService   *service.DeviceService
    taxpayerService *service.TaxpayerService
    logger          *zap.Logger
}

func NewDeviceHandler(
    deviceService *service.DeviceService,
    taxpayerService *service.TaxpayerService,
    logger *zap.Logger,
) *DeviceHandler {
    return &DeviceHandler{
        deviceService:   deviceService,
        taxpayerService: taxpayerService,
        logger:          logger,
    }
}

func (h *DeviceHandler) VerifyTaxpayerInformation(c *gin.Context) {
    var req models.VerifyTaxpayerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.NewAPIError(400, "Invalid request", ""))
        return
    }

    // Validate device model headers
    modelName := c.GetHeader("DeviceModelName")
    modelVersion := c.GetHeader("DeviceModelVersionNo")
    if modelName == "" || modelVersion == "" {
        c.JSON(422, models.NewAPIError(422, "Missing device model headers", "DEV06"))
        return
    }

    // Call service
    resp, err := h.deviceService.VerifyTaxpayer(req, modelName, modelVersion)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, resp)
}

// Implement other handlers: RegisterDevice, IssueCertificate, GetConfig, etc.
```

#### `/internal/handlers/receipt_handler.go`
```go
package handlers

// ReceiptHandler with methods: SubmitReceipt, SubmitFile, GetFileStatus
```

#### `/internal/handlers/fiscal_day_handler.go`
```go
package handlers

// FiscalDayHandler with methods: OpenDay, CloseDay, GetStatus
```

#### `/internal/handlers/user_handler.go`
```go
package handlers

// UserHandler with all user management methods
```

#### `/internal/handlers/middleware.go`
```go
package handlers

import (
    "crypto/x509"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// LoggerMiddleware adds request logging
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        logger.Info("request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
        )
    }
}

// CertificateAuthMiddleware validates client certificate
func CertificateAuthMiddleware(cryptoService *service.CryptoService, logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract client certificate from TLS connection
        if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
            c.AbortWithStatusJSON(401, models.NewAPIError(401, "Certificate required", ""))
            return
        }

        cert := c.Request.TLS.PeerCertificates[0]
        
        // Verify certificate
        if err := cryptoService.VerifyCertificate(cert); err != nil {
            c.AbortWithStatusJSON(401, models.NewAPIError(401, "Invalid certificate", ""))
            return
        }

        // Extract device ID from certificate CN
        deviceID, err := extractDeviceID(cert.Subject.CommonName)
        if err != nil {
            c.AbortWithStatusJSON(401, models.NewAPIError(401, "Invalid certificate", ""))
            return
        }

        // Store device ID in context
        c.Set("deviceID", deviceID)
        c.Set("certificate", cert)

        c.Next()
    }
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, DeviceModelName, DeviceModelVersionNo")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

## Phase 5: Utility Functions

### Files to Implement

#### `/internal/utils/signature.go`
```go
package utils

import (
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "fiscalization-api/internal/models"
    "sort"
    "strconv"
    "strings"
)

// GenerateReceiptHash generates SHA-256 hash for receipt signature
func GenerateReceiptHash(receipt *models.Receipt, previousHash []byte) ([]byte, error) {
    // Build signature string according to spec (section 13.2.1)
    var sb strings.Builder
    
    // 1. deviceID
    sb.WriteString(strconv.Itoa(receipt.DeviceID))
    
    // 2. receiptType (uppercase)
    sb.WriteString(strings.ToUpper(receipt.ReceiptType.String()))
    
    // 3. receiptCurrency (uppercase)
    sb.WriteString(strings.ToUpper(receipt.ReceiptCurrency))
    
    // 4. receiptGlobalNo
    sb.WriteString(strconv.Itoa(receipt.ReceiptGlobalNo))
    
    // 5. receiptDate (ISO 8601 format)
    sb.WriteString(receipt.ReceiptDate.Format("2006-01-02T15:04:05"))
    
    // 6. receiptTotal in cents
    totalCents := int64(receipt.ReceiptTotal * 100)
    sb.WriteString(strconv.FormatInt(totalCents, 10))
    
    // 7. receiptTaxes (sorted by taxID, then taxCode)
    sortedTaxes := make([]models.ReceiptTax, len(receipt.ReceiptTaxes))
    copy(sortedTaxes, receipt.ReceiptTaxes)
    sort.Slice(sortedTaxes, func(i, j int) bool {
        if sortedTaxes[i].TaxID != sortedTaxes[j].TaxID {
            return sortedTaxes[i].TaxID < sortedTaxes[j].TaxID
        }
        code1 := ""
        code2 := ""
        if sortedTaxes[i].TaxCode != nil {
            code1 = *sortedTaxes[i].TaxCode
        }
        if sortedTaxes[j].TaxCode != nil {
            code2 = *sortedTaxes[j].TaxCode
        }
        return code1 < code2
    })
    
    for _, tax := range sortedTaxes {
        // taxCode
        if tax.TaxCode != nil {
            sb.WriteString(*tax.TaxCode)
        }
        
        // taxPercent (format: XX.XX or empty for exempt)
        if tax.TaxPercent != nil {
            sb.WriteString(fmt.Sprintf("%.2f", *tax.TaxPercent))
        }
        
        // taxAmount in cents
        sb.WriteString(strconv.FormatInt(int64(tax.TaxAmount*100), 10))
        
        // salesAmountWithTax in cents
        sb.WriteString(strconv.FormatInt(int64(tax.SalesAmountWithTax*100), 10))
    }
    
    // 8. previousReceiptHash (if not first receipt)
    if previousHash != nil {
        sb.Write(previousHash)
    }
    
    // Generate SHA-256 hash
    data := []byte(sb.String())
    hash := sha256.Sum256(data)
    return hash[:], nil
}

// GenerateFiscalDayHash generates SHA-256 hash for fiscal day signature
func GenerateFiscalDayHash(fiscalDay *models.FiscalDay, counters []models.FiscalDayCounter) ([]byte, error) {
    // Implementation according to spec (section 13.3.1)
    // ...
}

// VerifyReceiptSignature verifies receipt device signature
func VerifyReceiptSignature(receipt *models.Receipt, publicKey interface{}) error {
    // Implementation
    // ...
}

// GenerateQRCode generates QR code data for receipt
func GenerateQRCode(receipt *models.Receipt, qrURL string) (string, error) {
    // Implementation according to spec (section 11)
    // Format: <qrUrl>/<deviceID>/<receiptDate>/<receiptGlobalNo>/<receiptQrData>
    // ...
}
```

#### `/internal/utils/qr_code.go`
```go
package utils

import (
    "github.com/skip2/go-qrcode"
)

// GenerateQRCodeImage generates QR code image from data
func GenerateQRCodeImage(data string, size int) ([]byte, error) {
    return qrcode.Encode(data, qrcode.Medium, size)
}
```

## Phase 6: Database Layer

#### `/internal/database/db.go`
```go
package database

import (
    "fmt"
    "fiscalization-api/internal/config"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

func NewConnection(cfg config.DatabaseConfig) (*sqlx.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )

    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)

    return db, nil
}

func RunMigrations(db *sqlx.DB, path string) error {
    // Run migrations using golang-migrate
    // ...
}
```

## Phase 7: Testing

Create test files for each component:

```
internal/
├── handlers/
│   ├── device_handler_test.go
│   ├── receipt_handler_test.go
│   └── ...
├── service/
│   ├── crypto_service_test.go
│   ├── device_service_test.go
│   └── ...
└── repository/
    ├── device_repository_test.go
    └── ...
```

Example test:
```go
package handlers_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestDeviceHandler_VerifyTaxpayer(t *testing.T) {
    // Setup mocks
    // Create test request
    // Assert response
}
```

## Phase 8: Additional Features

### Email Service
```go
// internal/service/email_service.go
// Send security codes, notifications
```

### SMS Service
```go
// internal/service/sms_service.go
// Send security codes via SMS
```

### JWT Token Service
```go
// internal/service/token_service.go
// Generate and verify JWT tokens for users
```

### Audit Logging
```go
// internal/service/audit_service.go
// Log all important operations
```

## Deployment Checklist

- [ ] Generate production certificates
- [ ] Setup PostgreSQL with replication
- [ ] Setup Redis for caching
- [ ] Configure HTTPS/TLS
- [ ] Setup monitoring (Prometheus, Grafana)
- [ ] Configure log aggregation (ELK Stack)
- [ ] Setup backup strategy
- [ ] Configure rate limiting
- [ ] Setup CI/CD pipeline
- [ ] Security audit
- [ ] Load testing
- [ ] Documentation

## Performance Considerations

1. **Database Indexing**: Ensure all foreign keys and frequently queried fields are indexed
2. **Connection Pooling**: Configure appropriate pool sizes
3. **Caching**: Use Redis for frequently accessed data
4. **Batch Processing**: Process offline files asynchronously
5. **Query Optimization**: Use EXPLAIN ANALYZE for slow queries

## Security Best Practices

1. **Certificate Management**: Regular rotation, secure storage
2. **Password Hashing**: Use bcrypt with appropriate cost
3. **Input Validation**: Validate all user inputs
4. **SQL Injection Prevention**: Use parameterized queries
5. **Rate Limiting**: Implement per-endpoint rate limits
6. **HTTPS Only**: Enforce TLS 1.2+
7. **Security Headers**: Add appropriate security headers

## Monitoring & Observability

1. **Metrics**: Track API latency, error rates, throughput
2. **Logging**: Structured logging with correlation IDs
3. **Tracing**: Implement distributed tracing
4. **Alerting**: Setup alerts for critical errors
5. **Health Checks**: Implement comprehensive health checks

This guide provides the foundation for a complete implementation. Each phase builds upon the previous one, creating a robust, production-ready fiscalization API.
