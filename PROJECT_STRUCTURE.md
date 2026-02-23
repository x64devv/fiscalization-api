# ZIMRA Fiscalization API - Go Implementation

## Project Structure

```
fiscalization-api/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── models/
│   │   ├── device.go              # Device models
│   │   ├── taxpayer.go            # Taxpayer models
│   │   ├── receipt.go             # Receipt models
│   │   ├── fiscal_day.go          # Fiscal day models
│   │   ├── enums.go               # Enumerations
│   │   └── errors.go              # Custom error types
│   ├── repository/
│   │   ├── device_repository.go   # Device data access
│   │   ├── taxpayer_repository.go # Taxpayer data access
│   │   ├── receipt_repository.go  # Receipt data access
│   │   └── fiscal_day_repository.go # Fiscal day data access
│   ├── service/
│   │   ├── device_service.go      # Device business logic
│   │   ├── taxpayer_service.go    # Taxpayer business logic
│   │   ├── receipt_service.go     # Receipt business logic
│   │   ├── fiscal_day_service.go  # Fiscal day business logic
│   │   ├── crypto_service.go      # Certificate & signature handling
│   │   └── validation_service.go  # Validation logic
│   ├── handlers/
│   │   ├── device_handler.go      # Device endpoints
│   │   ├── receipt_handler.go     # Receipt endpoints
│   │   ├── fiscal_day_handler.go  # Fiscal day endpoints
│   │   ├── user_handler.go        # User management endpoints
│   │   └── middleware.go          # HTTP middleware
│   ├── database/
│   │   ├── migrations/            # Database migrations
│   │   └── db.go                  # Database connection
│   └── utils/
│       ├── signature.go           # Signature generation/verification
│       ├── qr_code.go             # QR code generation
│       └── helpers.go             # Helper functions
├── pkg/
│   └── api/
│       └── response.go            # API response helpers
├── migrations/
│   └── *.sql                      # SQL migration files
├── docs/
│   └── api/
│       └── swagger.yaml           # API documentation
├── configs/
│   ├── config.yaml                # Configuration file
│   └── config.example.yaml        # Example configuration
├── scripts/
│   ├── setup.sh                   # Setup script
│   └── migrate.sh                 # Migration script
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Key Features

1. **Multi-tenancy Support**: Separate data for different companies/taxpayers
2. **Multiple Devices**: Support for multiple fiscal devices per company
3. **Online/Offline Modes**: Handle both communication modes
4. **Certificate Management**: PKI infrastructure for device authentication
5. **Signature Verification**: Receipt and fiscal day signature validation
6. **File Processing**: Batch upload support for offline devices
7. **User Management**: Complete user registration and authentication
8. **Fiscal Day Management**: Open, close, and reconcile fiscal days
9. **Receipt Validation**: Multi-level validation with color coding
10. **Audit Trail**: Complete transaction history

## Technology Stack

- **Framework**: Gin or Echo (high-performance HTTP framework)
- **Database**: PostgreSQL (multi-tenancy support)
- **Cache**: Redis (session management, rate limiting)
- **Crypto**: Go crypto libraries for certificate handling
- **Migration**: golang-migrate or goose
- **Logging**: zap or logrus
- **Testing**: testify, mockery

## Database Design

### Core Tables
- `taxpayers` - Company/organization information
- `devices` - Fiscal devices with certificates
- `fiscal_days` - Fiscal day tracking
- `receipts` - All fiscal receipts (invoices, credit notes, debit notes)
- `receipt_lines` - Receipt line items
- `receipt_taxes` - Tax breakdown per receipt
- `receipt_payments` - Payment methods
- `fiscal_counters` - Fiscal day counters
- `users` - User management
- `certificates` - Certificate storage and history
- `file_uploads` - Offline file processing tracking

## API Endpoints Structure

### Device Management
- POST   /api/v1/device/verify-taxpayer
- POST   /api/v1/device/register
- POST   /api/v1/device/issue-certificate
- GET    /api/v1/device/config
- GET    /api/v1/device/status
- POST   /api/v1/device/ping

### Fiscal Day Management
- POST   /api/v1/fiscal-day/open
- POST   /api/v1/fiscal-day/close
- GET    /api/v1/fiscal-day/status

### Receipt Management
- POST   /api/v1/receipt/submit
- POST   /api/v1/receipt/file
- GET    /api/v1/receipt/file-status

### User Management
- GET    /api/v1/users/list
- POST   /api/v1/users/login
- POST   /api/v1/users/create-begin
- POST   /api/v1/users/create-confirm
- POST   /api/v1/users/send-security-code
- PUT    /api/v1/users/update
- PUT    /api/v1/users/change-password
- POST   /api/v1/users/reset-password-begin
- POST   /api/v1/users/reset-password-confirm

### Stock Management
- GET    /api/v1/stock/list

### System
- GET    /api/v1/server/certificate
- GET    /health
- GET    /metrics
