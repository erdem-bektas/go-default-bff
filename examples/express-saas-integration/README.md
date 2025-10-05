# Express.js Multi-SaaS Integration

Bu örnek, Express.js ile Zitadel çoklu SaaS organizasyonu entegrasyonunu gösterir.

## 🎯 Özellikler

- **Multi-tenant Authentication**: Her SaaS için ayrı OIDC stratejisi
- **Session Management**: Express session ile güvenli oturum yönetimi
- **Vault Integration**: Credentials otomatik Vault'tan alınır
- **Role-based Access**: Admin ve user rolleri
- **API Endpoints**: RESTful API ile SaaS yönetimi
- **Middleware System**: Modüler authentication middleware
- **Security**: Helmet, CORS, rate limiting

## 🚀 Kurulum ve Çalıştırma

### 1. Ön Koşullar
```bash
# Node.js yükleyin (v16+)
node --version

# Zitadel ve Vault'u başlatın
./start-vault-zitadel.sh

# SaaS organizasyonlarını oluşturun
./manage-saas-orgs.sh
```

### 2. Dependencies Yükleyin
```bash
cd examples/express-saas-integration

# NPM packages
npm install

# Environment variables
cp .env.example .env
```

### 3. Uygulamayı Başlatın
```bash
# Development mode
npm run dev

# Production mode
npm start

# Test endpoints
npm test
```

## 🌐 Endpoints

### Web Routes
- `GET /` - Ana sayfa (SaaS listesi)
- `GET /auth/:saas` - SaaS login başlat
- `GET /auth/:saas/callback` - OAuth callback
- `GET /dashboard/:saas` - SaaS dashboard (protected)
- `GET /logout` - Çıkış yap

### API Routes
- `GET /api/saas` - Tüm SaaS listesi
- `GET /api/saas/:saas` - SaaS bilgileri
- `GET /api/:saas/user` - Kullanıcı bilgileri (protected)
- `GET /api/:saas/token` - Access token (protected)
- `GET /api/:saas/profile` - Zitadel profil (protected)
- `GET /api/:saas/admin/org` - Organization bilgileri (admin)
- `GET /api/:saas/admin/config` - SaaS konfigürasyonu (admin)

### API Key Routes
- `GET /api/:saas/validate-key` - API key doğrula
- `GET /api/:saas/info` - SaaS bilgileri (API key ile)

### Webhook Routes
- `POST /api/:saas/webhooks/zitadel` - Zitadel webhook

## 🔐 Authentication Flow

### 1. Web Authentication
```bash
# 1. Kullanıcı SaaS login'e gider
GET /auth/sp1

# 2. Zitadel'e yönlendirilir
# 3. Kullanıcı giriş yapar
# 4. Callback'e döner
GET /auth/sp1/callback?code=...

# 5. Dashboard'a yönlendirilir
GET /dashboard/sp1
```

### 2. API Authentication
```bash
# Session-based (web login sonrası)
curl -H "Cookie: connect.sid=..." \
  http://localhost:3001/api/sp1/user

# API Key-based
curl -H "X-API-Key: sp1_12345678" \
  http://localhost:3001/api/sp1/info
```

## 🧪 Test Senaryoları

### 1. SP1 Web Login
```bash
# 1. Ana sayfaya git
open http://localhost:3001

# 2. SP1 Login'e tıkla
# 3. admin@sp1.localhost / SP1AdminPass123! ile giriş
# 4. Dashboard'ı gör
```

### 2. SP2 Web Login
```bash
# 1. SP2 Login'e tıkla
# 2. admin@sp2.localhost / SP2AdminPass123! ile giriş
# 3. Dashboard'ı gör
```

### 3. API Tests
```bash
# SaaS listesi
curl http://localhost:3001/api/saas

# SP1 bilgileri
curl http://localhost:3001/api/saas/sp1

# Health check
curl http://localhost:3001/api/health
```

## 🔧 Konfigürasyon

### Environment Variables
```bash
# .env dosyası
PORT=3001
NODE_ENV=development
SESSION_SECRET=your-secret-key
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=dev-root
```

### SaaS Configurations
Vault'tan otomatik yüklenir:
```bash
# SP1 config
vault kv get secret/saas/sp1/oauth

# SP2 config
vault kv get secret/saas/sp2/oauth
```

## 🛡️ Security Features

### 1. Session Security
- HttpOnly cookies
- Secure flag (production)
- Session timeout
- CSRF protection

### 2. API Security
- Rate limiting
- CORS configuration
- Helmet security headers
- Input validation

### 3. Authentication
- OIDC token verification
- Role-based access control
- API key authentication
- Webhook signature verification

## 📊 Middleware System

### Authentication Middleware
```javascript
const { requireAuth, requireAdmin, validateSaaS } = require('./middleware/auth');

// SaaS validation
app.use('/:saas/*', validateSaaS(saasConfigs));

// Authentication required
app.get('/dashboard/:saas', requireAuth(), handler);

// Admin required
app.get('/admin/:saas', requireAuth(), requireAdmin(), handler);
```

### Custom Middleware
```javascript
// User context
app.use(addUserContext());

// API key auth
app.use('/api/:saas/public', requireApiKey(saasConfigs));
```

## 🔄 Session Management

### Session Store
```javascript
// Memory store (development)
app.use(session({
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false
}));

// Redis store (production)
const RedisStore = require('connect-redis')(session);
app.use(session({
  store: new RedisStore({ client: redisClient }),
  // ...
}));
```

### Session Data
```javascript
req.session = {
  passport: { user: {...} },
  currentSaas: 'sp1',
  returnTo: '/dashboard/sp1'
}
```

## 🐛 Sorun Giderme

### 1. Server Başlamıyor
```bash
# Port kontrolü
lsof -i :3001

# Dependencies
npm install

# Environment
cp .env.example .env
```

### 2. Authentication Hatası
```bash
# Vault connection
curl http://localhost:8200/v1/sys/health

# SaaS configs
curl http://localhost:3001/api/saas

# Zitadel health
curl http://localhost:8080/debug/ready
```

### 3. Session Problemi
```bash
# Clear browser cookies
# Check session secret
# Restart server
```

## 📚 Dependencies

### Core
- `express` - Web framework
- `passport` - Authentication
- `passport-openidconnect` - OIDC strategy
- `express-session` - Session management

### Security
- `helmet` - Security headers
- `cors` - CORS handling
- `express-rate-limit` - Rate limiting

### Integration
- `node-vault` - Vault client
- `axios` - HTTP client

### Development
- `nodemon` - Auto-restart
- `dotenv` - Environment variables

## 🚀 Production Deployment

### 1. Environment
```bash
NODE_ENV=production
SESSION_SECRET=strong-random-secret
VAULT_ADDR=https://vault.company.com
VAULT_TOKEN=production-token
```

### 2. Security
- HTTPS zorunlu
- Secure cookies
- Redis session store
- Rate limiting
- Input validation

### 3. Monitoring
- Health check endpoints
- Error logging
- Performance metrics
- Webhook monitoring

## 📖 API Documentation

### Response Format
```json
{
  "success": true,
  "data": {...},
  "error": "Error message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Error Codes
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

### Rate Limits
- `100 requests/15 minutes` per IP
- `1000 requests/hour` per authenticated user
- `10 requests/minute` for webhook endpoints