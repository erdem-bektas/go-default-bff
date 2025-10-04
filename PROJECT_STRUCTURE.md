# Go Default BFF - Vault & Zitadel Integration Project

## 📋 Proje Özeti

Bu proje, HashiCorp Vault ile Zitadel IAM sisteminin entegrasyonunu sağlayan bir Go Backend-for-Frontend (BFF) uygulamasıdır. Vault secret management için, Zitadel ise authentication ve authorization için kullanılmaktadır.

## 🏗️ Proje Yapısı

```
go-default-bff/
├── 🚀 BAŞLATMA VE DURDURMA
│   ├── start-vault-zitadel.sh      # Ana başlatma scripti
│   └── stop-vault-zitadel.sh       # Servisleri durdurma scripti
│
├── 🧪 TEST SİSTEMİ
│   ├── test-services.sh            # Ana test wrapper (kolay kullanım)
│   └── tests/                      # Test klasörü
│       ├── README.md               # Test dokümantasyonu
│       ├── test-vault-health.sh    # Vault sağlık testi
│       ├── test-zitadel-health.sh  # Zitadel temel test
│       ├── test-zitadel-functionality.sh # Zitadel gelişmiş test
│       └── test-all-services.sh    # Kapsamlı test runner
│
├── 🔐 VAULT KONFIGÜRASYONU
│   ├── vault/dev/                  # Development ortamı
│   │   ├── docker-compose.yml      # Vault dev container setup
│   │   ├── docker-run-default.sh   # Varsayılan Vault başlatma
│   │   ├── docker-run-from-env.sh  # Environment'tan Vault başlatma
│   │   └── setup-zitadel-secrets.sh # Zitadel secret'larını Vault'a yükleme
│   └── vault/prod/                 # Production ortamı
│       ├── docker-compose.yml      # Production Vault setup
│       ├── .env                    # Production environment variables
│       └── config/
│           ├── vault.hcl           # Vault production konfigürasyonu
│           └── for-prod.sh         # Production setup scripti
│
├── 🔑 ZITADEL KONFIGÜRASYONU
│   ├── zitadel/
│   │   ├── docker-compose.yaml     # Zitadel + PostgreSQL setup
│   │   ├── get-vault-secrets.sh    # Vault'tan secret'ları çekme
│   │   └── .env                    # Zitadel environment variables
│
├── 🗄️ DATABASE
│   └── db-init-zitadel.sql         # Zitadel database initialization
│
└── 📚 API & UYGULAMA
    └── api/                        # Go BFF uygulaması (geliştirilecek)
```

## 🔄 Servis Akışı

### 1. Başlatma Sırası (`start-vault-zitadel.sh`)
```bash
1. Vault Dev Environment başlatılır (docker-compose)
2. Vault'a Zitadel secret'ları yüklenir:
   - Database credentials
   - Zitadel config (masterkey, domain, etc.)
   - First instance setup (admin user, org)
3. Zitadel Vault'tan credential'ları çeker (`get-vault-secrets.sh`)
4. Zitadel + PostgreSQL başlatılır
```

### 2. Test Sistemi (`test-services.sh`)
```bash
# Ana wrapper - kolay kullanım
./test-services.sh [vault|zitadel|functionality|all]

# Vault testleri (8 test)
- Container durumu, HTTP endpoints, authentication
- Secret'ların varlığı, external erişim, UI

# Zitadel testleri (11 test)  
- Container'lar, health endpoints, OIDC discovery
- Management API, login, OAuth2, database

# Functionality testleri (12 test)
- Gelişmiş Zitadel özellikleri ve initialization
```

## 🔐 Vault Konfigürasyonu

### Development Mode
- **Container**: `vault-dev`
- **Port**: `8200`
- **Token**: `dev-root`
- **Storage**: In-memory (geçici)
- **UI**: http://localhost:8200

### Vault Secret Yapısı
```
secret/zitadel/
├── database/
│   ├── postgres_host=db
│   ├── postgres_port=5432
│   ├── postgres_database=zitadel
│   ├── postgres_admin_username=postgres
│   ├── postgres_admin_password=postgres
│   ├── postgres_user_username=zitadel
│   └── postgres_user_password=zitadel
├── config/
│   ├── masterkey="MasterkeyNeedsToHave32Characters"
│   ├── external_domain=localhost
│   ├── external_secure=false
│   └── tls_enabled=false
└── firstinstance/
    ├── org_name="My Organization"
    ├── org_human_username=root
    ├── org_human_password="RootPassword1!"
    ├── login_client_username=login-client
    ├── login_client_name="Automatically Initialized IAM_LOGIN_CLIENT"
    └── pat_expiration_date="2029-01-01T00:00:00Z"
```

## 🔑 Zitadel Konfigürasyonu

### Container Setup
- **Zitadel Container**: `zitadel-zitadel-1`
- **Database Container**: `zitadel-db-1` (PostgreSQL 17)
- **Login Container**: `zitadel-login-1`

### Ports
- **Console UI**: http://localhost:8080/ui/console
- **Login UI**: http://localhost:3000
- **API**: http://localhost:8080
- **Database**: localhost:5432

### Admin Credentials
- **Username**: `root@my-organization.localhost`
- **Password**: `RootPassword1!`

### Key Endpoints
- Health: `/debug/healthz`, `/debug/ready`
- OIDC Discovery: `/.well-known/openid_configuration`
- Management API: `/management/v1/`
- OAuth2: `/oauth/v2/token`, `/oauth/v2/authorize`

## 🧪 Test Sistemi Detayları

### Test Kategorileri

| Test Scripti | Amaç | Test Sayısı | Süre |
|-------------|------|-------------|------|
| `test-vault-health.sh` | Vault erişilebilirlik ve sağlık | 8 | ~10s |
| `test-zitadel-health.sh` | Zitadel temel sağlık | 11 | ~15s |
| `test-zitadel-functionality.sh` | Zitadel gelişmiş özellikler | 12 | ~20s |
| `test-all-services.sh` | Kapsamlı test | 19 | ~25s |

### Test Sonuçları
- ✅ **PASS**: Test başarılı
- ❌ **FAIL**: Test başarısız  
- ℹ️ **INFO**: Bilgilendirme mesajı

## 🚀 Kullanım Senaryoları

### Geliştirme Ortamı Başlatma
```bash
# 1. Servisleri başlat
./start-vault-zitadel.sh

# 2. Sağlık kontrolü
./test-services.sh

# 3. Geliştirmeye başla
# - Vault UI: http://localhost:8200
# - Zitadel Console: http://localhost:8080/ui/console
```

### Sorun Giderme
```bash
# Sadece Vault test et
./test-services.sh vault

# Sadece Zitadel test et  
./test-services.sh zitadel

# Container'ları kontrol et
docker ps

# Logları incele
docker logs vault-dev
docker logs zitadel-zitadel-1
docker logs zitadel-db-1
```

### Servisleri Durdurma
```bash
./stop-vault-zitadel.sh
```

## 🔧 Önemli Dosyalar

### Konfigürasyon Dosyaları
- `vault/dev/docker-compose.yml` - Vault dev setup
- `zitadel/docker-compose.yaml` - Zitadel + DB setup
- `zitadel/.env` - Zitadel environment variables

### Script Dosyaları
- `start-vault-zitadel.sh` - Ana başlatma scripti
- `vault/dev/setup-zitadel-secrets.sh` - Secret setup
- `zitadel/get-vault-secrets.sh` - Secret retrieval

### Test Dosyaları
- `test-services.sh` - Ana test wrapper
- `tests/` - Tüm test scriptleri

## 🎯 Gelecek Geliştirmeler

1. **Go BFF Uygulaması** - `api/` klasöründe
2. **Production Setup** - `vault/prod/` konfigürasyonu
3. **CI/CD Pipeline** - Otomatik test ve deployment
4. **Monitoring** - Health check ve metrics
5. **Security Hardening** - Production security ayarları

## 📝 Notlar

- **Development Mode**: Vault in-memory storage kullanır (veriler kalıcı değil)
- **Initialization Time**: Zitadel tam başlatma 2-5 dakika sürebilir
- **Port Requirements**: 3000, 5432, 8080, 8200 portları kullanılır
- **macOS Compatibility**: Tüm scriptler macOS/zsh uyumludur
- **Docker Dependency**: Tüm servisler Docker container'ları olarak çalışır

## 🔗 Hızlı Referans

```bash
# Başlat
./start-vault-zitadel.sh

# Test et
./test-services.sh

# Durdur  
./stop-vault-zitadel.sh

# Vault UI
open http://localhost:8200

# Zitadel Console
open http://localhost:8080/ui/console
```