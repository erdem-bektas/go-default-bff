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

## Kurulum

### Gereksinimler
- Go 1.21+
- Air (hot reload için)
- Docker & Docker Compose (PostgreSQL için)

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

Port: `3000`