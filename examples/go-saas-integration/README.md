# Go SaaS Integration Example

Bu örnek, Zitadel ile çoklu SaaS organizasyonu entegrasyonunu gösterir.

## 🎯 Özellikler

- **Multi-tenant Authentication**: Her SaaS için ayrı organization
- **Vault Integration**: Credentials Vault'tan otomatik alınır
- **OIDC Flow**: Tam OAuth2/OIDC implementasyonu
- **Isolated Sessions**: Her SaaS için izole oturum yönetimi

## 🚀 Çalıştırma

### 1. Ön Koşullar
```bash
# Zitadel ve Vault'u başlatın
./start-vault-zitadel.sh

# SaaS organizasyonlarını oluşturun
./manage-saas-orgs.sh
```

### 2. Go Uygulamasını Çalıştırın
```bash
cd examples/go-saas-integration

# Dependencies
go mod tidy

# Run
go run main.go
```

### 3. Test Edin
```bash
# Ana sayfa
open http://localhost:8090

# SP1 login
open http://localhost:8090/auth/sp1

# SP2 login  
open http://localhost:8090/auth/sp2

# Status
open http://localhost:8090/status
```

## 🔄 Authentication Flow

1. **User**: `/auth/sp1` adresine gider
2. **App**: Zitadel'e yönlendirir (SP1 organization)
3. **User**: SP1 credentials ile giriş yapar
4. **Zitadel**: Callback'e yönlendirir
5. **App**: Token exchange yapar ve user bilgilerini döner

## 📋 Endpoint'ler

- `GET /` - Ana sayfa
- `GET /auth/{saas}` - SaaS login başlat
- `GET /auth/{saas}/callback` - OAuth callback
- `GET /status` - Konfigürasyon durumu

## 🔧 Konfigürasyon

Uygulama Vault'tan otomatik olarak şu bilgileri alır:

```bash
# SP1 OAuth config
vault kv get secret/saas/sp1/oauth

# SP2 OAuth config  
vault kv get secret/saas/sp2/oauth
```

## 🧪 Test Senaryoları

### SP1 Test
1. http://localhost:8090/auth/sp1
2. admin@sp1.localhost / SP1AdminPass123! ile giriş
3. User bilgilerini JSON olarak görün

### SP2 Test
1. http://localhost:8090/auth/sp2
2. admin@sp2.localhost / SP2AdminPass123! ile giriş
3. User bilgilerini JSON olarak görün

## 🔐 Güvenlik

- **State Parameter**: CSRF koruması
- **ID Token Verification**: Token doğrulama
- **Isolated Configs**: Her SaaS için ayrı config
- **Vault Integration**: Credentials güvenli saklanır

## 📚 Kod Yapısı

```go
type SaaSManager struct {
    configs       map[string]*SaaSConfig    // SaaS configs
    vault         *api.Client               // Vault client
    verifiers     map[string]*oidc.IDTokenVerifier // OIDC verifiers
    oauth2Configs map[string]*oauth2.Config // OAuth2 configs
}
```

## 🐛 Sorun Giderme

### Vault Connection Error
```bash
# Vault status kontrol
vault status

# Token kontrol
echo $VAULT_TOKEN
```

### OIDC Provider Error
```bash
# Zitadel health check
curl http://localhost:8080/debug/ready

# OIDC discovery
curl http://localhost:8080/.well-known/openid_configuration
```

### Configuration Not Found
```bash
# SaaS configs kontrol
vault kv list secret/saas/

# Organizasyonları yeniden oluştur
./manage-saas-orgs.sh
```