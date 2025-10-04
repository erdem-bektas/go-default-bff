# Test Scripts - Vault & Zitadel Health Checker

Bu klasör Vault ve Zitadel servislerinin sağlık durumunu ve fonksiyonalitesini test eden scriptleri içerir.

## 📋 Test Scriptleri

### 1. `test-vault-health.sh`
**Vault Erişilebilirlik ve Sağlık Testi**

- ✅ Container durumu kontrolü
- ✅ HTTP endpoint erişilebilirliği
- ✅ Authentication testi (dev-root token)
- ✅ Zitadel secret'larının varlığı
- ✅ External client erişimi
- ✅ UI erişilebilirliği
- ✅ Vault unsealed durumu

### 2. `test-zitadel-health.sh`
**Zitadel Temel Sağlık Testi**

- ✅ Container durumu (Zitadel + Database)
- ✅ Health endpoint'ler
- ✅ Console UI erişimi
- ✅ OIDC Discovery endpoint
- ✅ Management API
- ✅ Login sayfası
- ✅ OAuth2 token endpoint
- ✅ Database bağlantısı
- ✅ Database tabloları

### 3. `test-zitadel-functionality.sh`
**Zitadel Gelişmiş Fonksiyonalite Testi**

- ✅ Tam initialization kontrolü
- ✅ Console erişim testi
- ✅ Management API detaylı test
- ✅ OIDC konfigürasyon
- ✅ Login sayfası fonksiyonalitesi
- ✅ Database schema kontrolü
- ✅ Health endpoint'ler
- ✅ OAuth2 endpoint'ler

### 4. `test-all-services.sh`
**Kapsamlı Test Runner**

Tüm test scriptlerini sırayla çalıştırır ve genel sonuçları özetler.

## 🚀 Kullanım

### Ana Dizinden (Önerilen)
```bash
# Tüm testleri çalıştır
./test-services.sh

# Sadece Vault testi
./test-services.sh vault

# Sadece Zitadel testi
./test-services.sh zitadel

# Gelişmiş Zitadel testleri
./test-services.sh functionality

# Yardım
./test-services.sh help
```

### Doğrudan Test Klasöründen
```bash
cd tests

# Bireysel testler
./test-vault-health.sh
./test-zitadel-health.sh
./test-zitadel-functionality.sh

# Tüm testler
./test-all-services.sh
```

## 📊 Test Sonuçları

### Başarılı Test Çıktısı
```
✅ PASS: Vault container is running
✅ PASS: Vault HTTP endpoint is accessible
✅ PASS: Vault authentication with dev-root token works
...
🎉 All tests passed! Services are healthy and functional.
```

### Başarısız Test Çıktısı
```
❌ FAIL: Vault external access failed
❌ FAIL: Zitadel database tables not found
...
⚠️  Some tests failed. Check the output above.
```

## 🔧 Troubleshooting

### Vault Sorunları
- Container çalışıyor mu? `docker ps | grep vault`
- Port erişilebilir mi? `curl http://localhost:8200/v1/sys/health`
- Token doğru mu? `dev-root` token'ı kullanılıyor

### Zitadel Sorunları
- Container'lar çalışıyor mu? `docker ps | grep zitadel`
- Initialization tamamlandı mı? Birkaç dakika bekleyin
- Database bağlantısı var mı? `docker logs zitadel-db-1`

### Genel Sorunlar
1. Servisleri yeniden başlatın: `./start-vault-zitadel.sh`
2. Container loglarını kontrol edin: `docker logs <container-name>`
3. Port çakışması var mı? `lsof -i :8200,8080,5432`

## 🎯 Test Kapsamı

| Kategori | Vault | Zitadel Basic | Zitadel Advanced |
|----------|-------|---------------|------------------|
| Container Health | ✅ | ✅ | ✅ |
| HTTP Endpoints | ✅ | ✅ | ✅ |
| Authentication | ✅ | ✅ | ✅ |
| Database | ✅ | ✅ | ✅ |
| UI Access | ✅ | ✅ | ✅ |
| API Functionality | ✅ | ✅ | ✅ |
| OIDC/OAuth2 | - | ✅ | ✅ |
| External Access | ✅ | - | - |

## 📝 Notlar

- Testler macOS/Linux uyumludur
- Zitadel initialization 2-5 dakika sürebilir
- Vault dev mode'da çalışır (production için uygun değil)
- Tüm testler non-destructive'dir (veri değiştirmez)