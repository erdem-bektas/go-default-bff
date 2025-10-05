# Zitadel SaaS Organization Management

Bu Terraform konfigürasyonu, Zitadel'de SaaS projeleriniz için otomatik organization yönetimi sağlar.

## 🎯 Özellikler

- **Otomatik Organization Oluşturma**: Her SaaS için izole organizasyonlar
- **OAuth Uygulamaları**: Her SaaS için OIDC uygulaması
- **Admin Kullanıcıları**: Her organization için admin user
- **Güvenlik Politikaları**: Şifre, giriş ve güvenlik politikaları
- **Vault Entegrasyonu**: Credentials Vault'ta güvenli saklanır

## 🚀 Hızlı Başlangıç

### 1. Ön Koşullar
```bash
# Terraform yükleyin
brew install terraform

# jq yükleyin (JSON parsing için)
brew install jq

# Zitadel ve Vault'u başlatın
./start-vault-zitadel.sh
```

### 2. Organizasyonları Oluşturun
```bash
# Ana script ile (önerilen)
./manage-saas-orgs.sh

# Veya manuel olarak
cd terraform
./setup-admin-token.sh
./get-zitadel-token.sh
terraform init
terraform plan
terraform apply
```

## 📋 Varsayılan Konfigürasyon

### SP1 Organization
- **Name**: SaaS Project 1
- **Domain**: sp1.localhost
- **Admin**: admin@sp1.localhost / SP1AdminPass123!
- **Features**: 
  - Kayıt açık
  - Username/password giriş
  - External IDP kapalı
  - Şifre: min 8 karakter, büyük/küçük harf, rakam, sembol

### SP2 Organization  
- **Name**: SaaS Project 2
- **Domain**: sp2.localhost
- **Admin**: admin@sp2.localhost / SP2AdminPass123!
- **Features**:
  - Kayıt kapalı
  - Username/password giriş
  - External IDP açık
  - Şifre: min 10 karakter, büyük/küçük harf, rakam, sembol

## 🔧 Konfigürasyon Değiştirme

### Yeni Organization Eklemek
`main.tf` dosyasında `saas_organizations` variable'ına yeni entry ekleyin:

```hcl
variable "saas_organizations" {
  default = {
    sp1 = { ... }
    sp2 = { ... }
    sp3 = {
      name        = "SaaS Project 3"
      domain      = "sp3.localhost"
      admin_email = "admin@sp3.localhost"
      admin_password = "SP3AdminPass123!"
      features = {
        login_policy_allow_register = true
        login_policy_allow_username_password = true
        login_policy_allow_external_idp = false
        password_complexity_policy = {
          min_length    = 12
          has_uppercase = true
          has_lowercase = true
          has_number    = true
          has_symbol    = true
        }
      }
    }
  }
}
```

### Politika Değiştirme
`policies.tf` dosyasında ilgili policy resource'larını düzenleyin.

## 📊 Çıktılar

### Organizations
```bash
terraform output organizations
```

### OAuth Credentials
```bash
terraform output oauth_applications
```

### Admin Users
```bash
terraform output admin_users
```

## 🔐 Vault Integration

Credentials otomatik olarak Vault'a kaydedilir:

```bash
# OAuth credentials
vault kv get secret/saas/sp1/oauth
vault kv get secret/saas/sp2/oauth

# Admin token
vault kv get secret/zitadel/admin
```

## 🧪 Test Etme

### 1. Organization Erişimi
```bash
# SP1 Console
open http://localhost:8080/ui/console/orgs/[SP1_ORG_ID]

# SP2 Console  
open http://localhost:8080/ui/console/orgs/[SP2_ORG_ID]
```

### 2. Admin Giriş
- **SP1**: admin@sp1.localhost / SP1AdminPass123!
- **SP2**: admin@sp2.localhost / SP2AdminPass123!

### 3. OAuth Test
```bash
# Authorization URL
http://localhost:8080/oauth/v2/authorize?client_id=[CLIENT_ID]&response_type=code&scope=openid%20profile%20email&redirect_uri=http://sp1.localhost/auth/callback

# Token Exchange
curl -X POST http://localhost:8080/oauth/v2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=[CODE]&redirect_uri=http://sp1.localhost/auth/callback" \
  -u "[CLIENT_ID]:[CLIENT_SECRET]"
```

## 🔄 Lifecycle Management

### Güncelleme
```bash
# Konfigürasyon değiştirdikten sonra
terraform plan
terraform apply
```

### Silme
```bash
# Tüm organizasyonları sil
./manage-saas-orgs.sh destroy

# Veya manuel
terraform destroy
```

### Backup
```bash
# State backup
cp terraform.tfstate terraform.tfstate.backup

# Vault backup
vault kv get -format=json secret/saas/ > saas-backup.json
```

## 🚨 Güvenlik Notları

1. **Production'da**:
   - `terraform.tfvars` dosyasını git'e eklemeyin
   - Remote backend kullanın (S3, Terraform Cloud)
   - Vault production mode kullanın
   - TLS enable edin

2. **Şifreler**:
   - Production'da güçlü şifreler kullanın
   - Şifre rotation policy uygulayın
   - MFA enable edin

3. **Network**:
   - Production'da proper domain kullanın
   - HTTPS zorunlu yapın
   - Network segmentation uygulayın

## 🐛 Sorun Giderme

### Terraform Hataları
```bash
# State lock sorunu
terraform force-unlock [LOCK_ID]

# State refresh
terraform refresh

# Import existing resource
terraform import zitadel_org.saas_orgs["sp1"] [ORG_ID]
```

### Zitadel API Hataları
```bash
# Token kontrolü
vault kv get secret/zitadel/admin

# API test
curl -H "Authorization: Bearer [TOKEN]" http://localhost:8080/management/v1/orgs/_search
```

### Vault Hataları
```bash
# Vault status
vault status

# Secret kontrolü
vault kv list secret/
```

## 📚 Referanslar

- [Zitadel Terraform Provider](https://registry.terraform.io/providers/zitadel/zitadel/latest/docs)
- [Zitadel Management API](https://zitadel.com/docs/apis/resources/mgmt)
- [Zitadel OIDC Guide](https://zitadel.com/docs/guides/integrate/login/oidc)
- [HashiCorp Vault](https://vaultproject.io/docs)