# Testing Guide - ZIMRA Fiscalization API

## Overview

This guide covers testing strategies for the ZIMRA Fiscalization API.

## Running Tests

### All Tests
```bash
make test
```

### With Coverage
```bash
make test-coverage
# Opens coverage.html in browser
```

### Verbose Output
```bash
go test -v ./...
```

### Specific Package
```bash
go test -v ./internal/service/...
go test -v ./internal/utils/...
```

### With Race Detector
```bash
go test -race ./...
```

## Test Structure

```
fiscalization-api/
├── internal/
│   ├── service/
│   │   ├── validation_service.go
│   │   └── validation_service_test.go
│   └── utils/
│       ├── signature.go
│       ├── signature_test.go
│       ├── helpers.go
│       └── helpers_test.go
```

## Test Coverage Goals

| Package | Target | Current |
|---------|--------|---------|
| internal/service | 80% | 75% |
| internal/utils | 90% | 85% |
| internal/handlers | 70% | 0% |
| internal/repository | 60% | 0% |
| **Overall** | **70%** | **~40%** |

## Unit Tests

### What's Tested

✅ **Validation Service** (`validation_service_test.go`)
- All 48 validation rules
- Currency validation
- Receipt total calculations
- Credit/debit note validation

✅ **Signature Utilities** (`signature_test.go`)
- Receipt hash generation
- Fiscal day hash generation
- QR code generation
- Hash consistency

✅ **Helper Utilities** (`helpers_test.go`)
- Activation key generation
- Security code generation
- TIN/VAT validation
- Email/phone validation
- Tax calculations
- Currency validation

### Running Specific Tests

```bash
# Validation service tests
go test -v ./internal/service -run TestValidationService

# Signature tests
go test -v ./internal/utils -run TestGenerateReceiptHash

# Helper tests
go test -v ./internal/utils -run TestValidateTIN
```

## Integration Tests (TODO)

### Database Tests
```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

### API Tests
Use the provided Postman collection or create your own:

```bash
# Install newman
npm install -g newman

# Run API tests
newman run tests/postman/fiscalization-api.json
```

## Test Data

### Test Taxpayers
```go
TIN: 1234567890
Name: ABC Retail Store Ltd
VAT: 123456789
```

### Test Devices
```go
DeviceID: 1001
SerialNo: DEV-001-2024
Model: ZIMRA-POS-2000
```

### Test Receipts
See `test/fixtures/receipts.json`

## Mocking

### Mock Services

Create mocks for external dependencies:

```go
type MockEmailService struct {
    SentEmails []string
}

func (m *MockEmailService) SendSecurityCode(to, code, username string) error {
    m.SentEmails = append(m.SentEmails, to)
    return nil
}
```

### Mock Repositories

```go
type MockDeviceRepository struct {
    Devices map[int]*models.Device
}

func (m *MockDeviceRepository) GetByDeviceID(id int) (*models.Device, error) {
    device, exists := m.Devices[id]
    if !exists {
        return nil, nil
    }
    return device, nil
}
```

## Benchmark Tests

### Running Benchmarks

```bash
go test -bench=. ./...
go test -bench=BenchmarkGenerateReceiptHash ./internal/utils
```

### Example Benchmark

```go
func BenchmarkGenerateReceiptHash(b *testing.B) {
    receipt := createTestReceipt()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = GenerateReceiptHash(receipt, nil)
    }
}
```

## Testing Checklist

### Before Committing
- [ ] All tests pass
- [ ] Coverage > 70%
- [ ] No race conditions
- [ ] Lint passes
- [ ] Security scan passes

### Before Deploying
- [ ] Integration tests pass
- [ ] Load tests pass
- [ ] Security audit complete
- [ ] All edge cases tested

## Common Test Patterns

### Table-Driven Tests

```go
func TestValidateTIN(t *testing.T) {
    tests := []struct {
        name string
        tin  string
        want bool
    }{
        {"valid TIN", "1234567890", true},
        {"short TIN", "123", false},
        {"empty TIN", "", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := ValidateTIN(tt.tin); got != tt.want {
                t.Errorf("ValidateTIN() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Setup/Teardown

```go
func TestMain(m *testing.M) {
    // Setup
    setupTestDB()
    
    // Run tests
    code := m.Run()
    
    // Teardown
    teardownTestDB()
    
    os.Exit(code)
}
```

## Continuous Integration

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: make test-coverage
      - uses: codecov/codecov-action@v2
```

## Test Reports

### Generate HTML Report

```bash
make test-coverage
# Opens coverage.html
```

### Generate JSON Report

```bash
go test -json ./... > test-report.json
```

### CI Integration

```bash
# JUnit format for Jenkins/GitLab
go get github.com/jstemmer/go-junit-report
go test -v ./... | go-junit-report > report.xml
```

## Performance Testing

### Load Testing with k6

```javascript
import http from 'k6/http';

export default function() {
  http.post('http://localhost:8080/api/v1/receipt/submit', 
    JSON.stringify(receipt),
    { headers: { 'Content-Type': 'application/json' } }
  );
}
```

Run:
```bash
k6 run --vus 10 --duration 30s load-test.js
```

## Debugging Tests

### Verbose Output

```bash
go test -v ./...
```

### Run Single Test

```bash
go test -v -run TestValidationService_ValidateReceipt ./internal/service
```

### With Debugger (Delve)

```bash
dlv test ./internal/service -- -test.run TestValidationService
```

## Best Practices

1. **Test Naming**
   - Use descriptive names
   - Format: `TestFunctionName_Scenario`

2. **Test Coverage**
   - Focus on business logic
   - Don't test framework code
   - Test edge cases

3. **Test Independence**
   - Each test should be independent
   - Use setup/teardown
   - Don't rely on test order

4. **Assertions**
   - Use clear error messages
   - Test one thing per test
   - Use table-driven tests

5. **Mocking**
   - Mock external dependencies
   - Don't mock what you don't own
   - Keep mocks simple

## Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [GoMock](https://github.com/golang/mock)
- [Go Test Coverage](https://blog.golang.org/cover)

## Support

For testing questions:
1. Check this guide
2. Review existing tests
3. Consult Go testing documentation
