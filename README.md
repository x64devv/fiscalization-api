# ZIMRA Fiscalization API

A comprehensive Go-based API implementation for ZIMRA fiscal device management and receipt processing, supporting multi-tenancy and multiple devices per company.

## Features

- ✅ Multi-tenant architecture (multiple companies/taxpayers)
- ✅ Multiple fiscal devices per taxpayer
- ✅ Online and offline device operation modes
- ✅ PKI certificate management for device authentication
- ✅ Receipt validation with color-coded error system
- ✅ Fiscal day management (open, close, reconciliation)
- ✅ Signature generation and verification (RSA 2048, ECC secp256r1)
- ✅ Batch file processing for offline devices
- ✅ Complete user management system
- ✅ QR code generation for receipts
- ✅ Comprehensive audit logging
- ✅ RESTful API with JSON responses

## Architecture

### Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Migration**: golang-migrate
- **Logging**: zap
- **Container**: Docker & Docker Compose

### Project Structure

```
fiscalization-api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── models/         # Data models and DTOs
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic
│   ├── handlers/       # HTTP handlers
│   ├── database/       # Database connection
│   └── utils/          # Utility functions
├── migrations/         # Database migrations
├── configs/           # Configuration files
├── docs/              # Documentation
└── scripts/           # Utility scripts
```

## Getting Started

### Prerequisites

- Go 1.26 or higher
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/fiscalization-api.git
cd fiscalization-api
```

2. **Install dependencies**
```bash
go mod download
```

3. **Setup configuration**
```bash
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with your settings
```

4. **Setup database**
```bash
# Create database
createdb fiscalization_db

# Run migrations
make migrate-up
```

5. **Run the application**
```bash
make run
```

Or using Docker:
```bash
docker-compose up -d
```

## Configuration

Edit `configs/config.yaml`:

```yaml
server:
  port: 8080
  mode: development

database:
  host: localhost
  port: 5432
  user: fiscalization
  password: your_password
  dbname: fiscalization_db

crypto:
  certificate_path: certs/server.crt
  private_key_path: certs/server.key
  ca_certificate_path: certs/ca.crt
```

## API Endpoints

### Device Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/device/verify-taxpayer` | Verify taxpayer information | No |
| POST | `/api/v1/device/register` | Register new device | No |
| POST | `/api/v1/device/issue-certificate` | Issue/renew certificate | Yes |
| GET | `/api/v1/device/config` | Get device configuration | Yes |
| GET | `/api/v1/device/status` | Get device status | Yes |
| POST | `/api/v1/device/ping` | Device heartbeat | Yes |

### Fiscal Day Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/fiscal-day/open` | Open fiscal day | Yes |
| POST | `/api/v1/fiscal-day/close` | Close fiscal day | Yes |
| GET | `/api/v1/fiscal-day/status` | Get fiscal day status | Yes |

### Receipt Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/receipt/submit` | Submit receipt (online) | Yes |
| POST | `/api/v1/receipt/file` | Submit file (offline) | Yes |
| GET | `/api/v1/receipt/file-status` | Get file processing status | Yes |

### User Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users/list` | List users | Yes |
| POST | `/api/v1/users/login` | User login | Yes |
| POST | `/api/v1/users/create-begin` | Start user creation | Yes |
| POST | `/api/v1/users/create-confirm` | Confirm user creation | Yes |
| PUT | `/api/v1/users/update` | Update user | Yes |
| PUT | `/api/v1/users/change-password` | Change password | Yes |

### Stock Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/stock/list` | Get stock list | Yes |

## Database Schema

### Key Tables

- **taxpayers**: Company/organization information
- **devices**: Fiscal device registration and certificates
- **fiscal_days**: Fiscal day tracking per device
- **receipts**: All fiscal receipts (invoices, credit/debit notes)
- **receipt_lines**: Line items for receipts
- **receipt_taxes**: Tax breakdown
- **receipt_payments**: Payment methods
- **fiscal_counters**: Daily fiscal counters
- **users**: User management
- **file_uploads**: Offline file processing tracking

## Multi-Tenancy Implementation

The API supports multiple companies (taxpayers) with the following approach:

1. **Taxpayer Isolation**: Each taxpayer has a unique TIN
2. **Device Assignment**: Devices belong to specific taxpayers
3. **Data Segregation**: All queries filter by taxpayer_id
4. **Certificate Binding**: Certificates tied to device and taxpayer

## Security Features

### Certificate Authentication

- Client certificate authentication for all protected endpoints
- Certificate issuance and renewal
- Certificate revocation support
- Thumbprint-based validation

### Signature Verification

- Receipt signatures (device and server)
- Fiscal day signatures
- SHA-256 hashing
- Support for RSA 2048 and ECC secp256r1

### Password Security

- bcrypt hashing
- Security code verification
- Token-based authentication
- Password complexity requirements

## Validation System

Receipts are validated with a color-coded system:

- **Grey**: Missing previous receipt (chain break)
- **Yellow**: Minor validation errors
- **Red**: Major validation errors

## Development

### Running Tests

```bash
make test
```

### Running with Coverage

```bash
make test-coverage
```

### Linting

```bash
make lint
```

### Database Migrations

```bash
# Create new migration
make migrate-create NAME=add_new_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Deployment

### Using Docker

```bash
# Build image
make docker-build

# Start services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

### Environment Variables

Set these in production:

```bash
export PORT=8080
export DB_HOST=your-db-host
export DB_PASSWORD=your-db-password
export CONFIG_PATH=/path/to/config.yaml
```

## API Examples

### Register Device

```bash
curl -X POST http://localhost:8080/api/v1/device/register \
  -H "Content-Type: application/json" \
  -H "DeviceModelName: POS-2000" \
  -H "DeviceModelVersionNo: 1.0" \
  -d '{
    "deviceID": 123,
    "activationKey": "ABC12345",
    "certificateRequest": "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----"
  }'
```

### Submit Receipt

```bash
curl -X POST http://localhost:8080/api/v1/receipt/submit \
  -H "Content-Type: application/json" \
  --cert device.crt \
  --key device.key \
  -d '{
    "deviceID": 123,
    "receipt": {
      "receiptType": "FiscalInvoice",
      "receiptCurrency": "USD",
      "receiptCounter": 1,
      "receiptGlobalNo": 1,
      "invoiceNo": "INV-001",
      "receiptDate": "2024-02-06T10:30:00",
      "receiptLinesTaxInclusive": true,
      "receiptLines": [...],
      "receiptTaxes": [...],
      "receiptPayments": [...],
      "receiptTotal": 100.00,
      "receiptDeviceSignature": {...}
    }
  }'
```

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### Metrics

Metrics available at `/metrics` (when enabled)

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

This project is licensed under the MIT License.

## Support

For issues and questions:
- Create an issue on GitHub
- Email: support@example.com

## Acknowledgments

- ZIMRA for the fiscalization specification
- Go community for excellent libraries
- Contributors and maintainers