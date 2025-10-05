#!/bin/bash

# Zitadel admin token'ını Vault'tan al ve Terraform için hazırla
# Bu script Vault'tan admin PAT'ı alır ve terraform.tfvars dosyasına yazar

set -e

VAULT_ADDR="http://localhost:8200"
VAULT_TOKEN="dev-root"

echo "🔐 Zitadel admin token'ı Vault'tan alınıyor..."

# Vault'tan admin PAT'ı al
ADMIN_PAT=$(vault kv get -field=admin_pat secret/zitadel/admin 2>/dev/null || echo "")

if [ -z "$ADMIN_PAT" ]; then
    echo "❌ Admin PAT Vault'ta bulunamadı. Önce Zitadel'i başlatın ve admin PAT'ı oluşturun."
    echo "💡 Alternatif: Manuel token oluşturun:"
    echo "   1. http://localhost:8080/ui/console adresine gidin"
    echo "   2. root@my-organization.localhost ile giriş yapın"
    echo "   3. Organization Settings > Personal Access Tokens"
    echo "   4. Yeni token oluşturun ve aşağıdaki komutu çalıştırın:"
    echo "   vault kv put secret/zitadel/admin admin_pat=YOUR_TOKEN"
    exit 1
fi

# terraform.tfvars dosyasını oluştur
cat > terraform.tfvars << EOF
# Zitadel Configuration
zitadel_domain = "localhost"
zitadel_port = "8080"
zitadel_insecure = true
zitadel_token = "$ADMIN_PAT"
EOF

echo "✅ Terraform variables hazırlandı (terraform.tfvars)"
echo "🚀 Şimdi 'terraform plan' ve 'terraform apply' çalıştırabilirsiniz"