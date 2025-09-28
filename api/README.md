# Fiber App with Hot Reload & Trace ID

Go Fiber framework kullanarak oluşturulmuş web uygulaması. Air ile hot reload desteği ve Zap logger ile trace_id takibi içerir.

## Özellikler

- 🚀 **Go Fiber** - Hızlı web framework
- 🔄 **Air** - Hot reload desteği
- 📝 **Zap Logger** - Yapılandırılmış loglama
- 🔍 **Trace ID** - Her request için unique takip ID'si
- 🛡️ **Middleware** - Recovery, Logger, Trace ID
- 🏥 **Health Check** - Sistem durumu kontrolü
- 📊 **Metrics** - Detaylı sistem metrikleri
- 🗄️ **PostgreSQL** - GORM ile database işlemleri
- 👥 **User CRUD** - Kullanıcı yönetimi API'si
- 🔐 **Role Management** - Rol tabanlı yetkilendirme
- 📚 **Swagger UI** - API dokümantasyonu
- ⚡ **Redis Cache** - Yüksek performans cache sistemi
- 🔐 **Zitadel Auth** - OAuth2/OIDC kimlik doğrulama

## Kurulum

### Gereksinimler
- Go 1.21+
- Air (hot reload için)
- Docker & Docker Compose (PostgreSQL için)
- Redis (cache için)
- Zitadel (authentication için)

### Bağımlılıkları yükle
```bash
make install
```

### Air kurulumu (eğer yoksa)
```bash
make install-air
```

### Database kurulumu
```bash
# PostgreSQL'i Docker ile başlat
make db-up

# Database loglarını izle
make db-logs

# Zitadel'i başlat
make zitadel-up

# Zitadel loglarını izle
make zitadel-logs
```

## Çalıştırma

### Geliştirme ortamı (Hot Reload)
```bash
make dev
```

### Normal çalıştırma
```bash
make run
```

### Build
```bash
make build
```

### Swagger Dokümantasyonu
```bash
# Swagger docs oluştur
make docs

# Swagger tool kurulumu (eğer yoksa)
make swagger-install
```

## API Endpoints

### GET /
Ana sayfa
```json
{
  "message": "Merhaba Fiber!",
  "trace_id": "uuid-string"
}
```

### GET /health
Sistem durumu
```json
{
  "status": "ok",
  "trace_id": "uuid-string"
}
```

### POST /test
Test endpoint
```json
{
  "message": "Test başarılı",
  "data": {...},
  "trace_id": "uuid-string"
}
```

## User API Endpoints

### GET /api/v1/users
Kullanıcıları listele (sayfalama ve arama desteği)
```bash
curl "http://localhost:3002/api/v1/users/?page=1&limit=10&search=ahmet"
```

### GET /api/v1/users/:id
Tek kullanıcı getir
```bash
curl "http://localhost:3002/api/v1/users/uuid"
```

### POST /api/v1/users
Yeni kullanıcı oluştur
```bash
curl -X POST http://localhost:3002/api/v1/users/ \
  -H "Content-Type: application/json" \
  -d '{"name":"Ahmet Yılmaz","email":"ahmet@example.com","age":30}'
```

### PUT /api/v1/users/:id
Kullanıcı güncelle
```bash
curl -X PUT http://localhost:3002/api/v1/users/uuid \
  -H "Content-Type: application/json" \
  -d '{"age":31,"active":false}'
```

### DELETE /api/v1/users/:id
Kullanıcı sil
```bash
curl -X DELETE http://localhost:3002/api/v1/users/uuid
```

## Trace ID

Her request için otomatik olarak unique bir trace_id oluşturulur:
- Response header'da `X-Trace-ID` olarak döner
- Response body'de `trace_id` field'ında bulunur
- Tüm loglarda trace_id ile işaretlenir
- Hata durumlarında trace_id ile takip edilebilir

## Loglama

Zap logger kullanılarak yapılandırılmış loglama:
- Request başlangıcı
- Hata durumları
- Endpoint çağrıları
- Tüm loglarda trace_id mevcut

## Geliştirme

Air sayesinde kod değişikliklerinde otomatik restart:
```bash
make dev
```

Port: `3003`

## API Dokümantasyonu

### Swagger UI
Tüm API endpoint'lerini görüntülemek ve test etmek için:
```
http://localhost:3003/docs
```

### Swagger JSON
API spesifikasyonu JSON formatında:
```
http://localhost:3003/swagger.json
```

## Cache Sistemi

### Redis Cache Özellikleri
- **User Cache**: Kullanıcı detayları 15 dakika cache'lenir
- **Role Cache**: Roller 30 dakika cache'lenir  
- **Auto Invalidation**: Veri güncellendiğinde cache otomatik temizlenir
- **Performance**: Cache hit'lerde 10x daha hızlı response

### Cache Endpoint'leri
```bash
# Cache istatistikleri
curl http://localhost:3003/api/v1/cache/stats

# Cache key'lerini listele
curl "http://localhost:3003/api/v1/cache/keys?pattern=user:*"

# Cache'i temizle
curl -X POST http://localhost:3003/api/v1/cache/flush

# Belirli key'i sil
curl -X DELETE http://localhost:3003/api/v1/cache/keys/user:uuid
```

## Authentication (Zitadel)

### OAuth2/OIDC Flow
1. **Login**: `GET /auth/login` - OAuth2 authorization URL alın
2. **Callback**: `GET /auth/callback` - Authorization code ile token alın
3. **Profile**: `GET /auth/profile` - Kullanıcı profil bilgileri
4. **Logout**: `POST /auth/logout` - Oturumu sonlandır

### Zitadel Kurulumu
```bash
# Zitadel'i başlat
make zitadel-up

# Zitadel admin paneli: http://localhost:8080
# İlk kurulumda admin kullanıcısı oluşturun
# Project ve Application oluşturup CLIENT_ID ve CLIENT_SECRET alın
```

### Environment Variables
```bash
ZITADEL_DOMAIN=http://localhost:8080
ZITADEL_CLIENT_ID=your_client_id
ZITADEL_CLIENT_SECRET=your_client_secret
ZITADEL_REDIRECT_URL=http://localhost:3003/auth/callback
```

### Protected Endpoints
Bazı endpoint'ler authentication gerektirir:
- User management (admin rolü)
- Cache management (admin/moderator rolü)
- Profile bilgileri (authenticated user)