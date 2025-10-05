# Zitadel Multi-SaaS Organization Management System

## 📋 Proje Özeti

Bu proje, HashiCorp Vault ve Zitadel IAM kullanarak çoklu SaaS organizasyonlarını otomatik yönetmek için geliştirilmiş tam entegre bir sistemdir. Hem local development hem de Hetzner server production deployment'ı destekler.

## 🏗️ Sistem Mimarisi

### Core Components
- **Zitadel**: Identity and Access Management (IAM)
- **HashiCorp Vault**: Secret Management
- **PostgreSQL**: Database
- **Terraform**: Infrastructure as Code
- **Docker**: Containerization
- **Traefik**: Reverse Proxy (Production)

### SaaS Organization Structure
```
Organization 1 (SP1/SaaS1)
├── OAuth Application
├── Admin Users
├── Security Policies
├── Custom Roles
└── Vault Credentials

Organization 2 (SP2/SaaS2)
├── OAuth Application
├── Admin Users  
├── Security Policies
├── Custom Roles
└── Vault Credentials
```

## 📁 Proje Yapısı

```
go-default-bff/
├── 🚀 BAŞLATMA VE YÖNETİM
│   ├── start-vault-zitadel.sh          # Local development başlatma
│   ├── manage-saas-orgs.sh             # Local SaaS yönetimi
│   └── test-services.sh                # Health check testleri
│
├── 🔧 TERRAFORM (Infrastructure as Code)
│   ├── main.tf                         # Development config
│   ├── organizations.tf                # Organization resources
│   ├── policies.tf                     # Security policies
│   ├── outputs.tf                      # Outputs
│   ├── setup-admin-token.sh           # Admin token setup
│   ├── get-zitadel-token.sh           # Token retrieval
│   └── environments/
│       └── production/                 # Production Terraform
│           ├── main.tf                 # Production provider config
│           ├── organizations.tf        # Production organizations
│           ├── policies.tf            # Production security
│           ├── outputs.tf             # Production outputs
│           └── terraform.tfvars.example
│
├── 🐳 DEPLOYMENT
│   └── hetzner/                        # Hetzner production deployment
│       ├── docker-compose.production.yml
│       ├── setup-production.sh        # Server setup script
│       ├── manage-production-saas.sh   # Production SaaS manager
│       ├── .env.production.example
│       └── README.md
│
├── 💻 EXAMPLES (Integration Examples)
│   ├── go-saas-integration/            # Go example
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── README.md
│   └── express-saas-integration/       # Express.js example
│       ├── server.js
│       ├── package.json
│       ├── middleware/auth.js
│       ├── routes/api.js
│       └── README.md
│
├── 🔐 VAULT & ZITADEL CONFIG
│   ├── vault/dev/                      # Development Vault
│   ├── vault/prod/                     # Production Vault
│   └── zitadel/                        # Zitadel configuration
│
└── 📚 DOCUMENTATION
    ├── PROJECT_STRUCTURE.md
    ├── AI_CONTEXT.md
    └── ZITADEL_MULTI_SAAS_DOCUMENTATION.md (bu dosya)
```

## 🎯 Kullanım Senaryoları

### 1. Local Development
```bash
# Servisleri başlat
./start-vault-zitadel.sh

# SaaS organizasyonları oluştur (SP1, SP2)
./manage-saas-orgs.sh

# Test et
./test-services.sh

# Go example test
cd examples/go-saas-integration && go run main.go

# Express example test  
cd examples/express-saas-integration && npm start
```

### 2. Hetzner Production Deployment
```bash
# 1. Server setup (Hetzner server'da)
curl -sSL https://raw.githubusercontent.com/your-repo/deploy/hetzner/setup-production.sh | bash

# 2. Configuration
cd /opt/zitadel-saas
nano .env.production  # Domain'leri güncelle

# 3. DNS Records
auth.yourdomain.com -> HETZNER_IP
vault.yourdomain.com -> HETZNER_IP
saas1.yourdomain.com -> HETZNER_IP

# 4. Start services
docker-compose -f docker-compose.production.yml up -d

# 5. Create SaaS organizations
./manage-production-saas.sh create
```

## 🔧 Terraform Konfigürasyonu

### Development Organizations (SP1, SP2)
```hcl
# terraform/main.tf
saas_organizations = {
  sp1 = {
    name = "SaaS Project 1"
    domain = "sp1.localhost"
    admin_email = "admin@sp1.localhost"
    admin_password = "SP1AdminPass123!"
    features = {
      login_policy_allow_register = true
      password_complexity_policy = { min_length = 8 }
    }
  }
  sp2 = {
    name = "SaaS Project 2" 
    domain = "sp2.localhost"
    admin_email = "admin@sp2.localhost"
    admin_password = "SP2AdminPass123!"
    features = {
      login_policy_allow_register = false
      password_complexity_policy = { min_length = 10 }
    }
  }
}
```

### Production Organizations (SaaS1, SaaS2)
```hcl
# terraform/environments/production/main.tf
production_saas_organizations = {
  saas1 = {
    name = "Production SaaS 1"
    domain = "saas1.yourdomain.com"
    admin_email = "admin@saas1.yourdomain.com"
    admin_password = "ProductionSecurePassword123!"
    features = {
      login_policy_allow_register = false  # Production'da kapalı
      password_complexity_policy = { min_length = 12 }  # Daha güçlü
    }
  }
}
```

### Terraform Resources Created
```hcl
# Her SaaS için oluşturulan kaynaklar:
resource "zitadel_org" "saas_orgs"                    # Organization
resource "zitadel_human_user" "org_admins"            # Admin users
resource "zitadel_org_member" "org_owners"            # Role assignments
resource "zitadel_project" "saas_projects"            # Projects
resource "zitadel_application_oidc" "saas_apps"       # OAuth apps
resource "zitadel_project_role" "user_roles"          # Custom roles
resource "zitadel_login_policy" "org_login_policies"  # Login policies
resource "zitadel_password_complexity_policy"         # Password policies
resource "vault_generic_secret" "saas_oauth_credentials" # Vault storage
```

## 🔐 Vault Secret Structure

### Development Secrets
```bash
secret/zitadel/database/     # DB credentials
secret/zitadel/config/       # Zitadel config
secret/zitadel/admin/        # Admin PAT
secret/saas/sp1/oauth        # SP1 OAuth credentials
secret/saas/sp2/oauth        # SP2 OAuth credentials
```

### Production Secrets
```bash
secret/production/saas/saas1/oauth  # Production SaaS1 OAuth
secret/production/saas/saas2/oauth  # Production SaaS2 OAuth
secret/zitadel/admin                # Production admin PAT
```

### OAuth Credential Format
```json
{
  "org_id": "123456789",
  "project_id": "987654321", 
  "client_id": "123456789@project_name",
  "client_secret": "secret_value",
  "issuer_url": "https://auth.yourdomain.com",
  "auth_url": "https://auth.yourdomain.com/oauth/v2/authorize",
  "token_url": "https://auth.yourdomain.com/oauth/v2/token",
  "userinfo_url": "https://auth.yourdomain.com/oidc/v1/userinfo"
}
```

## 🌐 Production Deployment Architecture

### Hetzner Server Services
```yaml
# docker-compose.production.yml
services:
  traefik:      # Reverse proxy + Let's Encrypt SSL
    ports: [80, 443]
    labels: # Automatic SSL certificates
    
  postgres:     # Production database
    volumes: [postgres-data]
    healthcheck: pg_isready
    
  vault:        # HashiCorp Vault
    labels: # https://vault.yourdomain.com
    volumes: [vault-data, vault-logs]
    
  zitadel:      # Zitadel IAM
    labels: # https://auth.yourdomain.com
    environment: # Production settings
    
  redis:        # Session store
    command: # Password protected
    
  prometheus:   # Monitoring
    labels: # https://prometheus.yourdomain.com
    
  grafana:      # Dashboards  
    labels: # https://grafana.yourdomain.com
```

### Security Configuration
```bash
# Firewall (UFW)
22/tcp   # SSH
80/tcp   # HTTP (redirect to HTTPS)
443/tcp  # HTTPS

# SSL/TLS
- Let's Encrypt automatic certificates
- HTTPS-only traffic
- Security headers (Helmet.js)
- HSTS enabled

# Authentication Security
- MFA required in production
- Strong password policy (12+ chars)
- Account lockout (3 attempts)
- Password expiry (90 days)
```

## 💻 Integration Examples

### Go Integration
```go
// examples/go-saas-integration/main.go
type SaaSManager struct {
    configs       map[string]*SaaSConfig
    vault         *api.Client
    verifiers     map[string]*oidc.IDTokenVerifier
    oauth2Configs map[string]*oauth2.Config
}

// Vault'tan SaaS configs yükle
func (sm *SaaSManager) loadConfigurations() error {
    // secret/saas/sp1/oauth, secret/saas/sp2/oauth
}

// SaaS-specific authentication
func (sm *SaaSManager) AuthHandler(saasID string) {
    // OAuth flow başlat
}
```

### Express.js Integration
```javascript
// examples/express-saas-integration/server.js
const saasConfigs = new Map();
const passportStrategies = new Map();

// Her SaaS için ayrı Passport strategy
passport.use(`oidc-${org}`, new OpenIDConnectStrategy({
    issuer: saasConfig.issuerUrl,
    clientID: saasConfig.clientId,
    // ...
}));

// SaaS-specific routes
app.get('/auth/:saas', extractSaasId, (req, res, next) => {
    const strategyName = passportStrategies.get(req.saasId);
    passport.authenticate(strategyName)(req, res, next);
});
```

## 🔄 Workflow & Commands

### Development Workflow
```bash
# 1. Start local environment
./start-vault-zitadel.sh

# 2. Create SaaS organizations
./manage-saas-orgs.sh create

# 3. Test integrations
cd examples/go-saas-integration && go run main.go
cd examples/express-saas-integration && npm start

# 4. Test authentication
open http://localhost:8090/auth/sp1  # Go example
open http://localhost:3001/auth/sp1  # Express example
```

### Production Workflow
```bash
# 1. Server setup (one-time)
ssh root@hetzner-ip
curl -sSL setup-production.sh | bash

# 2. Configuration
cd /opt/zitadel-saas
nano .env.production

# 3. Deploy services
docker-compose -f docker-compose.production.yml up -d

# 4. Create SaaS organizations
./manage-production-saas.sh create

# 5. Monitor
open https://grafana.yourdomain.com
```

### Terraform Commands
```bash
# Development
./manage-saas-orgs.sh plan     # Show plan
./manage-saas-orgs.sh create   # Create orgs
./manage-saas-orgs.sh destroy  # Destroy orgs

# Production
./deploy/hetzner/manage-production-saas.sh plan
./deploy/hetzner/manage-production-saas.sh create
./deploy/hetzner/manage-production-saas.sh status
```

## 🧪 Testing & Validation

### Health Checks
```bash
# Local
./test-services.sh           # All tests
./test-services.sh vault     # Vault only
./test-services.sh zitadel   # Zitadel only

# Production
curl https://auth.yourdomain.com/debug/ready
curl https://vault.yourdomain.com/v1/sys/health
```

### Authentication Tests
```bash
# SP1 Login (Development)
1. http://localhost:8090/auth/sp1
2. admin@sp1.localhost / SP1AdminPass123!

# SaaS1 Login (Production)  
1. https://saas1.yourdomain.com/auth/callback
2. admin@saas1.yourdomain.com / ProductionSecurePassword123!
```

## 📊 Monitoring & Backup

### Monitoring Stack
```bash
# Prometheus metrics
- Zitadel: /debug/metrics
- Vault: /v1/sys/metrics  
- PostgreSQL: Database metrics
- System: CPU, Memory, Disk

# Grafana dashboards
- System overview
- Application metrics
- Security events
- Performance metrics
```

### Backup Strategy
```bash
# Automated daily backups (2 AM)
/opt/zitadel-saas/backup.sh

# Backup contents:
- PostgreSQL dump
- Vault snapshot  
- Zitadel data
- Configuration files
- 7-day retention
```

## 🔧 Customization & Extension

### Adding New SaaS Organization
```hcl
# 1. Update terraform variables
saas_organizations = {
  sp1 = { ... }
  sp2 = { ... }
  sp3 = {  # New SaaS
    name = "SaaS Project 3"
    domain = "sp3.localhost"
    admin_email = "admin@sp3.localhost"
    admin_password = "SP3AdminPass123!"
    features = { ... }
  }
}

# 2. Apply changes
terraform apply
```

### Custom Security Policies
```hcl
# terraform/policies.tf
resource "zitadel_password_complexity_policy" "custom_policy" {
  min_length = 16      # Very strong
  has_uppercase = true
  has_lowercase = true  
  has_number = true
  has_symbol = true
}
```

### Integration with External Systems
```go
// Custom SaaS integration
type CustomSaaSIntegration struct {
    saasManager *SaaSManager
    externalAPI *ExternalAPIClient
}

func (c *CustomSaaSIntegration) SyncUsers() {
    // Zitadel'den users al
    // External system'e sync et
}
```

## 🚨 Troubleshooting

### Common Issues
```bash
# Vault connection error
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root
vault status

# Zitadel not ready
curl http://localhost:8080/debug/ready
docker logs zitadel-zitadel-1

# Terraform state issues
terraform refresh
terraform import zitadel_org.saas_orgs["sp1"] ORG_ID

# SSL certificate issues (production)
docker logs traefik
```

### Log Locations
```bash
# Development
docker logs vault-dev
docker logs zitadel-zitadel-1

# Production  
docker-compose -f docker-compose.production.yml logs
/var/log/nginx/
/opt/zitadel-saas/logs/
```

## 📚 Key Files Reference

### Critical Scripts
- `start-vault-zitadel.sh` - Local environment startup
- `manage-saas-orgs.sh` - Local SaaS management
- `deploy/hetzner/setup-production.sh` - Production server setup
- `deploy/hetzner/manage-production-saas.sh` - Production SaaS management

### Configuration Files
- `terraform/main.tf` - Development Terraform
- `terraform/environments/production/main.tf` - Production Terraform
- `deploy/hetzner/docker-compose.production.yml` - Production services
- `deploy/hetzner/.env.production.example` - Production environment

### Integration Examples
- `examples/go-saas-integration/main.go` - Go integration
- `examples/express-saas-integration/server.js` - Express.js integration

Bu dokümantasyon, projenin tüm bileşenlerini, kullanım senaryolarını ve deployment süreçlerini kapsamlı şekilde açıklar. Yeni bir prompt'ta bu bilgileri referans alarak soruları yanıtlayabilirim.