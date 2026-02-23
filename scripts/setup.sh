#!/bin/bash

# ZIMRA Fiscalization API - Setup Script
# This script sets up the development environment

set -e

echo "==================================="
echo "ZIMRA Fiscalization API Setup"
echo "==================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.21 or higher from https://golang.org/dl/"
    exit 1
fi

echo -e "${GREEN}✓ Go is installed: $(go version)${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}Warning: Docker is not installed${NC}"
    echo "Docker is optional but recommended for development"
else
    echo -e "${GREEN}✓ Docker is installed: $(docker --version)${NC}"
fi

# Create necessary directories
echo "Creating directories..."
mkdir -p bin
mkdir -p logs
mkdir -p certs
mkdir -p tmp

echo -e "${GREEN}✓ Directories created${NC}"

# Copy config if it doesn't exist
if [ ! -f configs/config.yaml ]; then
    echo "Creating config.yaml from example..."
    cp configs/config.example.yaml configs/config.yaml
    echo -e "${YELLOW}⚠ Please edit configs/config.yaml with your settings${NC}"
else
    echo -e "${GREEN}✓ Config file exists${NC}"
fi

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod download
go mod tidy

echo -e "${GREEN}✓ Dependencies downloaded${NC}"

# Install development tools
echo "Installing development tools..."
go install github.com/golang/mock/mockgen@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

echo -e "${GREEN}✓ Development tools installed${NC}"

# Generate self-signed certificates for development
if [ ! -f certs/server.key ]; then
    echo "Generating self-signed certificates for development..."
    
    # Generate CA private key
    openssl genrsa -out certs/ca.key 2048
    
    # Generate CA certificate
    openssl req -new -x509 -days 3650 -key certs/ca.key -out certs/ca.crt \
        -subj "/C=ZW/ST=Zimbabwe/L=Harare/O=ZIMRA/OU=Development/CN=ZIMRA CA"
    
    # Generate server private key
    openssl genrsa -out certs/server.key 2048
    
    # Generate server CSR
    openssl req -new -key certs/server.key -out certs/server.csr \
        -subj "/C=ZW/ST=Zimbabwe/L=Harare/O=ZIMRA/OU=Development/CN=localhost"
    
    # Sign server certificate with CA
    openssl x509 -req -days 365 -in certs/server.csr -CA certs/ca.crt \
        -CAkey certs/ca.key -CAcreateserial -out certs/server.crt
    
    # Clean up CSR
    rm certs/server.csr
    
    echo -e "${GREEN}✓ Self-signed certificates generated${NC}"
    echo -e "${YELLOW}⚠ These are development certificates only!${NC}"
else
    echo -e "${GREEN}✓ Certificates already exist${NC}"
fi

# Check if PostgreSQL is running (Docker or local)
if command -v docker &> /dev/null && docker ps | grep -q postgres; then
    echo -e "${GREEN}✓ PostgreSQL is running in Docker${NC}"
elif command -v psql &> /dev/null; then
    echo -e "${GREEN}✓ PostgreSQL is installed locally${NC}"
else
    echo -e "${YELLOW}⚠ PostgreSQL not detected${NC}"
    echo "You can start PostgreSQL using: docker-compose up -d postgres"
fi

echo ""
echo "==================================="
echo "Setup Complete!"
echo "==================================="
echo ""
echo "Next steps:"
echo "1. Edit configs/config.yaml with your settings"
echo "2. Start PostgreSQL: docker-compose up -d postgres"
echo "3. Run migrations: make migrate-up"
echo "4. Start the API: make run"
echo ""
echo "For more information, see README.md"
