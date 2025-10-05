# Hetzner Production Deployment

Bu klasör, Zitadel + Vault + SaaS organizasyonlarını Hetzner server üzerinde production ortamında çalıştırmak için gerekli tüm dosyaları içerir.

## 🎯 Özellikler

- **Production-Ready**: SSL/TLS, monitoring, backup
- **Docker Compose**: Tüm servisler containerized
- **Traefik**: Reverse proxy + Let's Encrypt SSL
- **Monitoring**: Prometheus + Grafana
- **Security**: Firewall, fail2ban, secure configs
- **Terraform**: Otomatik SaaS organization yönetimi

## 🚀 Hızlı Deployment

### 1. Hetzner Server Hazırlığı
```bash
# Hetzner server'a SSH ile bağlan
ssh root@YOUR_SERVER_IP

# Setup script'i çalıştır
curl -sSL https://raw.githubusercontent.com/your-repo/deploy/hetzner/setup-production.sh | bash
```

### 2. Konfigürasyon
```bash
cd /opt/zitadel-saas

# Environment dosyasını düzenle
nano .env.production

# Domain'leri güncelle:
# - ZITADEL_EXTERNAL_DOMAIN=auth.yourdomain.com
# - Diğer domain'ler
```

### 3. DNS Konfigürasyonu
```bash
# A Records oluştur:
auth.yourdomain.com     -> YOUR_SERVER_IP
vault.yourdomain.com    -> YOUR_SERVER_IP
saas1.yourdomain.com    -> YOUR_SERVER_IP
saas2.yourdomain.com    -> YOUR_SERVER_IP
grafana.yourdomain.com  -> YOUR_SERVER_IP
```

### 4. Servisleri Başlat
```bash
cd /opt/zitadel-saas

# Docker Compose dosyasını kopyala
cp /path/to/docker-compose.production.yml .

# Servisleri başlat
docker-compose -f docker-compose.production.yml up -d

# Logları kontrol et
docker-compose -f docker-compose.production.yml logs -f
```

### 5. SaaS Organizasyonları Oluştur
```bash
# Production SaaS manager'ı çalıştır
./manage-production-saas.sh

# Veya local'den remote olarak
./deploy/hetzner/manage-production-saas.sh
```

## 📋 Servis Yapısı

### Core Services
- **Traefik**: Reverse proxy (ports 80, 443)
- **PostgreSQL**: Database (internal)
- **Zitadel**: IAM system (https://auth.yourdomain.com)
- **Vault**: Secret management (https://vault.yourdomain.com)
- **Redis**: Session store (internal)

### Monitoring
- **Prometheus**: Metrics collection
- **Grafana**: Dashboards (https://grafana.yourdomain.com)

### Security
- **Let's Encrypt**: Automatic SSL certificates
- **UFW**: Firewall configuration
- **Fail2ban**: Intrusion prevention

## 🔧 Konfigürasyon Dosyaları

### Environment Variables
```bash
# .env.production
ZITADEL_EXTERNAL_DOMAIN=auth.yourdomain.com
VAULT_DOMAIN=vault.yourdomain.com
POSTGRES_PASSWORD=secure-password
ZITADEL_MASTERKEY=32-character-key
# ... diğer değişkenler
```

### Docker Compose
```yaml
# docker-compose.production.yml
services:
  traefik:     # Reverse proxy + SSL
  postgres:    # Database
  vault:       # Secret management
  zitadel:     # IAM
  redis:       # Session store
  prometheus:  # Monitoring
  grafana:     # Dashboards
```

### Terraform
```hcl
# terraform/environments/production/
# - Organizations
# - OAuth applications
# - Security policies
# - Vault integration
```

## 🔐 Güvenlik Konfigürasyonu

### Firewall (UFW)
```bash
# Açık portlar
22/tcp   # SSH
80/tcp   # HTTP (redirect to HTTPS)
443/tcp  # HTTPS

# Kapalı portlar (internal)
5432     # PostgreSQL
8200     # Vault
6379     # Redis
```

### SSL/TLS
- **Let's Encrypt**: Otomatik certificate yenileme
- **HTTPS Only**: Tüm traffic HTTPS'e redirect
- **HSTS**: HTTP Strict Transport Security
- **Security Headers**: Helmet.js security headers

### Authentication
- **MFA Required**: Production'da zorunlu
- **Strong Passwords**: Min 12 karakter
- **Password Expiry**: 90 gün
- **Account Lockout**: 3 failed attempt

## 📊 Monitoring

### Prometheus Metrics
- **Zitadel**: `/debug/metrics`
- **Vault**: `/v1/sys/metrics`
- **PostgreSQL**: Database metrics
- **Redis**: Cache metrics
- **Traefik**: Proxy metrics

### Grafana Dashboards
- **System Overview**: CPU, Memory, Disk
- **Application Metrics**: Zitadel, Vault
- **Security Events**: Login attempts, failures
- **Performance**: Response times, throughput

### Alerts
```bash
# Grafana alerts için
# - High CPU usage
# - Memory usage
# - Disk space
# - Service downtime
# - Failed login attempts
```

## 💾 Backup Strategy

### Automated Backups
```bash
# Daily backup at 2 AM
0 2 * * * /opt/zitadel-saas/backup.sh

# Backup includes:
# - PostgreSQL dump
# - Vault snapshot
# - Zitadel data
# - Configuration files
```

### Backup Storage
```bash
# Local backups (7 days retention)
/opt/zitadel-saas/backups/

# Optional: S3 backup
# Configure in .env.production:
BACKUP_S3_BUCKET=your-backup-bucket
BACKUP_S3_ACCESS_KEY=your-access-key
BACKUP_S3_SECRET_KEY=your-secret-key
```

## 🔄 Terraform Workflow

### Local Development
```bash
# Local'de test et
cd terraform/
./manage-saas-orgs.sh

# Production'a deploy et
cd terraform/environments/production/
terraform plan
terraform apply
```

### Production Management
```bash
# Hetzner server'da
./manage-production-saas.sh create

# Local'den remote
./deploy/hetzner/manage-production-saas.sh create
```

### Organization Management
```bash
# Yeni SaaS ekle
# terraform.tfvars'a yeni entry ekle
saas3 = {
  name = "New SaaS"
  domain = "saas3.yourdomain.com"
  # ...
}

# Apply changes
terraform apply
```

## 🐛 Sorun Giderme

### Service Health Check
```bash
# Tüm servisleri kontrol et
docker-compose -f docker-compose.production.yml ps

# Specific service logs
docker-compose -f docker-compose.production.yml logs zitadel
docker-compose -f docker-compose.production.yml logs vault

# Health endpoints
curl https://auth.yourdomain.com/debug/ready
curl https://vault.yourdomain.com/v1/sys/health
```

### SSL Certificate Issues
```bash
# Traefik logs
docker logs traefik

# Manual certificate check
openssl s_client -connect auth.yourdomain.com:443

# Force certificate renewal
docker exec traefik traefik version
```

### Database Issues
```bash
# PostgreSQL logs
docker logs postgres-production

# Database connection test
docker exec postgres-production psql -U postgres -d zitadel -c "SELECT version();"

# Database backup test
docker exec postgres-production pg_dump -U postgres zitadel > test_backup.sql
```

### Vault Issues
```bash
# Vault status
docker exec vault-production vault status

# Vault logs
docker logs vault-production

# Vault seal status
docker exec vault-production vault operator raft list-peers
```

## 📚 Maintenance

### Regular Tasks
```bash
# Weekly
- Check disk space
- Review logs
- Update Docker images
- Test backups

# Monthly
- Security updates
- Certificate renewal check
- Performance review
- Backup cleanup
```

### Updates
```bash
# Update Docker images
docker-compose -f docker-compose.production.yml pull
docker-compose -f docker-compose.production.yml up -d

# Update system packages
apt update && apt upgrade -y

# Update Terraform
terraform init -upgrade
```

## 🔗 Useful Commands

### Docker Management
```bash
# Restart all services
docker-compose -f docker-compose.production.yml restart

# Update specific service
docker-compose -f docker-compose.production.yml up -d zitadel

# View resource usage
docker stats

# Cleanup unused resources
docker system prune -a
```

### Log Management
```bash
# Follow all logs
docker-compose -f docker-compose.production.yml logs -f

# Specific service logs
docker-compose -f docker-compose.production.yml logs -f zitadel

# Log rotation (if needed)
docker run --log-driver json-file --log-opt max-size=10m --log-opt max-file=3
```

### Performance Monitoring
```bash
# System resources
htop
df -h
free -h

# Network connections
netstat -tulpn

# Docker resource usage
docker stats --no-stream
```

## 📞 Support

### Health Check URLs
- **Zitadel**: https://auth.yourdomain.com/debug/ready
- **Vault**: https://vault.yourdomain.com/v1/sys/health
- **Grafana**: https://grafana.yourdomain.com/api/health
- **Traefik**: https://traefik.yourdomain.com/ping

### Log Locations
- **Application Logs**: `docker-compose logs`
- **System Logs**: `/var/log/syslog`
- **Nginx Logs**: `/var/log/nginx/`
- **Backup Logs**: `/opt/zitadel-saas/logs/`

Bu production setup ile Hetzner server'ınızda enterprise-grade Zitadel + SaaS sistemi çalıştırabilirsiniz!