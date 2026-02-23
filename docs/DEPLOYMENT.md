# Production Deployment Guide

## Overview

This guide covers deploying the ZIMRA Fiscalization API to production.

## Pre-Deployment Checklist

### Code Quality
- [ ] All tests passing (`make test`)
- [ ] Test coverage > 70% (`make test-coverage`)
- [ ] Linter passing (`make lint`)
- [ ] Security scan passing (`make security`)
- [ ] No TODO/FIXME in production code
- [ ] Code reviewed

### Configuration
- [ ] Production config file created
- [ ] Environment variables documented
- [ ] Secrets stored securely (not in code)
- [ ] Database credentials rotated
- [ ] API keys configured

### Security
- [ ] TLS certificates obtained
- [ ] Certificate chain validated
- [ ] Strong passwords set
- [ ] Firewall rules configured
- [ ] Rate limiting enabled
- [ ] CORS properly configured

### Infrastructure
- [ ] Database backup configured
- [ ] Monitoring setup
- [ ] Logging configured
- [ ] Alerts configured
- [ ] Load balancer setup (if needed)

## Infrastructure Requirements

### Minimum Requirements

**Application Server:**
- CPU: 2 cores
- RAM: 4GB
- Disk: 20GB SSD
- OS: Ubuntu 22.04 LTS

**Database Server:**
- CPU: 4 cores
- RAM: 8GB
- Disk: 100GB SSD (with IOPS)
- PostgreSQL 15+

**Redis Server:**
- CPU: 2 cores
- RAM: 4GB
- Disk: 10GB

### Recommended Requirements (High Traffic)

**Application Server:**
- CPU: 4+ cores
- RAM: 8GB+
- Disk: 50GB SSD
- Load balanced (2+ instances)

**Database:**
- CPU: 8+ cores
- RAM: 16GB+
- Disk: 500GB SSD
- Replication enabled

## Deployment Methods

### Method 1: Docker Compose (Simple)

1. **Prepare Server**
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose -y
```

2. **Deploy Application**
```bash
# Clone repository
git clone https://github.com/yourusername/fiscalization-api.git
cd fiscalization-api

# Create production config
cp configs/config.example.yaml configs/config.yaml
nano configs/config.yaml

# Set environment variables
export DB_PASSWORD="$(openssl rand -base64 32)"
export JWT_SECRET="$(openssl rand -base64 64)"

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Run migrations
docker-compose exec api make migrate-up

# Check health
curl http://localhost:8080/health
```

### Method 2: Kubernetes (Scalable)

See `k8s/` directory for manifests.

```bash
# Create namespace
kubectl create namespace fiscalization

# Create secrets
kubectl create secret generic app-secrets \
  --from-literal=db-password=$DB_PASSWORD \
  --from-literal=jwt-secret=$JWT_SECRET \
  -n fiscalization

# Deploy
kubectl apply -f k8s/ -n fiscalization

# Check status
kubectl get pods -n fiscalization
```

### Method 3: Systemd Service (Traditional)

1. **Build Binary**
```bash
make build
```

2. **Install Service**
```bash
sudo cp bin/fiscalization-api /usr/local/bin/
sudo cp scripts/fiscalization-api.service /etc/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable fiscalization-api
sudo systemctl start fiscalization-api
```

## SSL/TLS Configuration

### Using Let's Encrypt

```bash
# Install certbot
sudo apt install certbot

# Obtain certificate
sudo certbot certonly --standalone \
  -d fiscalization.zimra.co.zw \
  --email admin@zimra.co.zw

# Certificates will be in:
# /etc/letsencrypt/live/fiscalization.zimra.co.zw/
```

### Using Reverse Proxy (Nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name fiscalization.zimra.co.zw;

    ssl_certificate /etc/letsencrypt/live/fiscalization.zimra.co.zw/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/fiscalization.zimra.co.zw/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    # Client certificate verification
    ssl_client_certificate /etc/ssl/certs/zimra-ca.crt;
    ssl_verify_client optional;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-SSL-Client-Cert $ssl_client_cert;
    }
}
```

## Database Setup

### PostgreSQL Production Config

```sql
-- Create database
CREATE DATABASE fiscalization_db;

-- Create user
CREATE USER fiscalization WITH ENCRYPTED PASSWORD 'strong_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE fiscalization_db TO fiscalization;

-- Enable required extensions
\c fiscalization_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Configure Connection Pooling

```yaml
# config.yaml
database:
  host: db.internal
  port: 5432
  user: fiscalization
  password: ${DB_PASSWORD}
  dbname: fiscalization_db
  sslmode: require
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
```

### Setup Replication

Master:
```conf
# postgresql.conf
wal_level = replica
max_wal_senders = 3
```

Slave:
```conf
# recovery.conf
standby_mode = on
primary_conninfo = 'host=master-db port=5432 user=replicator password=...'
```

## Monitoring

### Prometheus Metrics

Add to `main.go`:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### Grafana Dashboards

Import dashboard ID: `TBD`

### Health Checks

```bash
# Application health
curl https://fiscalization.zimra.co.zw/health

# Database health
psql -h localhost -U fiscalization -d fiscalization_db -c "SELECT 1"

# Redis health
redis-cli ping
```

## Logging

### Centralized Logging (ELK Stack)

1. **Install Filebeat**
```bash
curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.5.0-amd64.deb
sudo dpkg -i filebeat-8.5.0-amd64.deb
```

2. **Configure Filebeat**
```yaml
# /etc/filebeat/filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/fiscalization/*.log
    json.keys_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

### Log Rotation

```bash
# /etc/logrotate.d/fiscalization-api
/var/log/fiscalization/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 fiscalization fiscalization
    sharedscripts
    postrotate
        systemctl reload fiscalization-api
    endscript
}
```

## Backup Strategy

### Database Backups

```bash
#!/bin/bash
# /usr/local/bin/backup-db.sh

BACKUP_DIR="/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)
FILENAME="fiscalization_db_${DATE}.sql.gz"

pg_dump -U fiscalization fiscalization_db | gzip > ${BACKUP_DIR}/${FILENAME}

# Keep only last 30 days
find ${BACKUP_DIR} -name "*.sql.gz" -mtime +30 -delete

# Upload to S3
aws s3 cp ${BACKUP_DIR}/${FILENAME} s3://backups/fiscalization/
```

Schedule:
```cron
0 2 * * * /usr/local/bin/backup-db.sh
```

## Performance Tuning

### PostgreSQL

```conf
# postgresql.conf
shared_buffers = 2GB
effective_cache_size = 6GB
work_mem = 16MB
maintenance_work_mem = 512MB
max_connections = 200
```

### Application

```yaml
# config.yaml
server:
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120
  max_header_bytes: 1048576
```

## Security Hardening

### Firewall Rules

```bash
# Allow only necessary ports
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable
```

### Fail2Ban

```ini
# /etc/fail2ban/jail.local
[fiscalization-api]
enabled = true
port = https
filter = fiscalization-api
logpath = /var/log/fiscalization/api.log
maxretry = 5
bantime = 3600
```

### Regular Updates

```bash
# Weekly security updates
0 3 * * 0 apt update && apt upgrade -y
```

## Scaling

### Horizontal Scaling

```yaml
# docker-compose.scale.yml
services:
  api:
    deploy:
      replicas: 3
```

### Load Balancer

```nginx
upstream fiscalization_backend {
    least_conn;
    server app1:8080;
    server app2:8080;
    server app3:8080;
}

server {
    listen 443 ssl;
    location / {
        proxy_pass http://fiscalization_backend;
    }
}
```

## Rollback Procedure

### Docker

```bash
# Tag current version
docker tag fiscalization-api:latest fiscalization-api:backup

# Deploy new version
docker-compose pull
docker-compose up -d

# If issues, rollback
docker-compose down
docker tag fiscalization-api:backup fiscalization-api:latest
docker-compose up -d
```

### Database Migrations

```bash
# Rollback last migration
make migrate-down

# Rollback to specific version
make migrate-force VERSION=5
```

## Troubleshooting

### Common Issues

**1. High CPU Usage**
```bash
# Check processes
top -c
# Analyze queries
psql -c "SELECT * FROM pg_stat_activity WHERE state = 'active'"
```

**2. Database Connection Errors**
```bash
# Check connections
psql -c "SELECT count(*) FROM pg_stat_activity"
# Increase max_connections if needed
```

**3. Slow Queries**
```sql
-- Enable slow query log
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();
```

## Post-Deployment

### Verification

- [ ] Health check passing
- [ ] Metrics collecting
- [ ] Logs flowing to central system
- [ ] Backups running
- [ ] Alerts working
- [ ] SSL certificate valid
- [ ] Performance acceptable

### Documentation

- [ ] Update runbook
- [ ] Document configuration
- [ ] Update architecture diagrams
- [ ] Document rollback procedure

## Support Contacts

- **Application:** dev-team@company.com
- **Infrastructure:** ops@company.com
- **Database:** dba@company.com
- **Security:** security@company.com

## Disaster Recovery

See `docs/DISASTER_RECOVERY.md` for detailed DR procedures.

## Compliance

- ZIMRA Gateway v7.2 specification
- Data Protection Act compliance
- PCI DSS (if processing payments)
- Regular security audits

---

**Last Updated:** February 7, 2026  
**Version:** 1.0.0
