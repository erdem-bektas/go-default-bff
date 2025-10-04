# AI Assistant Context - Vault & Zitadel Project

> Bu dosya AI asistanları için proje yapısını hızlıca anlamak amacıyla oluşturulmuştur.

## 🤖 AI Assistant Quick Context

### Proje Türü
- **Go Backend-for-Frontend (BFF)** projesi
- **HashiCorp Vault** + **Zitadel IAM** entegrasyonu
- **Docker-based** development environment
- **macOS/zsh** uyumlu scriptler

### Ana Bileşenler

#### 1. Vault (Secret Management)
```yaml
Service: HashiCorp Vault
Mode: Development (in-memory)
Port: 8200
Token: dev-root
UI: http://localhost:8200
Container: vault-dev
```

#### 2. Zitadel (IAM/Auth)
```yaml
Service: Zitadel IAM
Ports: 8080 (API/Console), 3000 (Login)
Database: PostgreSQL 17 (port 5432)
Admin: root@my-organization.localhost / RootPassword1!
Containers: zitadel-zitadel-1, zitadel-db-1, zitadel-login-1
```

### Kritik Dosyalar (AI için önemli)

#### Başlatma/Durdurma
- `start-vault-zitadel.sh` - Ana başlatma scripti
- `stop-vault-zitadel.sh` - Durdurma scripti

#### Test Sistemi
- `test-services.sh` - Ana test wrapper
- `tests/test-vault-health.sh` - Vault testleri (8 test)
- `tests/test-zitadel-health.sh` - Zitadel testleri (11 test)
- `tests/test-zitadel-functionality.sh` - Gelişmiş testler (12 test)

#### Konfigürasyon
- `vault/dev/docker-compose.yml` - Vault dev setup
- `zitadel/docker-compose.yaml` - Zitadel setup
- `vault/dev/setup-zitadel-secrets.sh` - Secret initialization
- `zitadel/get-vault-secrets.sh` - Secret retrieval

### Vault Secret Yapısı
```
secret/zitadel/database/    # DB credentials
secret/zitadel/config/      # Zitadel config (masterkey, domain)
secret/zitadel/firstinstance/ # Admin user setup
```

### Servis Başlatma Akışı
1. Vault dev başlatılır
2. Zitadel secret'ları Vault'a yüklenir
3. Zitadel Vault'tan credential'ları çeker
4. Zitadel + PostgreSQL başlatılır

### Test Sistemi Kullanımı
```bash
./test-services.sh           # Tüm testler
./test-services.sh vault     # Sadece Vault
./test-services.sh zitadel   # Sadece Zitadel
./test-services.sh functionality # Gelişmiş Zitadel
```

### Yaygın Sorunlar & Çözümler

#### Vault Sorunları
- Token: `dev-root` kullanılmalı
- Port: 8200 erişilebilir olmalı
- Container: `vault-dev` çalışıyor olmalı

#### Zitadel Sorunları
- Initialization: 2-5 dakika sürebilir
- OIDC endpoint: Tam initialize olduktan sonra çalışır
- Database: PostgreSQL container'ı sağlıklı olmalı

#### Test Sorunları
- macOS: `timeout` komutu yok, `curl` timeout'suz kullanılır
- Container isimleri: Dynamic olabilir, `docker ps` ile kontrol
- Port çakışması: 3000, 5432, 8080, 8200 portları kullanılır

### AI Asistan İçin Öneriler

#### Yeni Prompt Geldiğinde
1. `PROJECT_STRUCTURE.md` dosyasını oku
2. Mevcut container durumunu kontrol et: `docker ps`
3. Test scriptlerini çalıştır: `./test-services.sh`
4. Sorun varsa ilgili logları incele: `docker logs <container>`

#### Vault İle İlgili Promptlar
- Secret yapısını kontrol et: `vault/dev/setup-zitadel-secrets.sh`
- Vault health: `./test-services.sh vault`
- UI erişimi: http://localhost:8200 (token: dev-root)

#### Zitadel İle İlgili Promptlar
- Konfigürasyon: `zitadel/docker-compose.yaml`
- Health check: `./test-services.sh zitadel`
- Console: http://localhost:8080/ui/console
- Admin: root@my-organization.localhost / RootPassword1!

#### Test İle İlgili Promptlar
- Test dokümantasyonu: `tests/README.md`
- Ana wrapper: `test-services.sh`
- Bireysel testler: `tests/` klasöründe

### Dosya Lokasyonları (AI için hızlı erişim)
```
📁 Ana dizin: /Users/erdem/side/go-default-bff/
📁 Vault config: vault/dev/ ve vault/prod/
📁 Zitadel config: zitadel/
📁 Test scriptleri: tests/
📄 Ana başlatma: start-vault-zitadel.sh
📄 Ana test: test-services.sh
📄 Proje dokümantasyonu: PROJECT_STRUCTURE.md
```

### Geliştirme Durumu
- ✅ Vault dev environment - Tamamlandı
- ✅ Zitadel setup - Tamamlandı  
- ✅ Test sistemi - Tamamlandı
- ✅ Vault-Zitadel entegrasyonu - Tamamlandı
- 🔄 Go BFF uygulaması - Geliştirilecek
- 🔄 Production setup - Geliştirilecek

### Kullanıcı Tercihleri
- **Dil**: Türkçe (kullanıcı Türkçe konuşuyor)
- **Platform**: macOS (zsh shell)
- **Stil**: Pratik, hızlı çözümler tercih ediyor
- **Test odaklı**: Çalışıp çalışmadığını test etmeyi önemsiyor