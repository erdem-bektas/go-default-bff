#!/bin/bash

# Zitadel admin token'ını otomatik oluştur ve Vault'a kaydet
# Bu script Zitadel Management API kullanarak admin PAT oluşturur

set -e

VAULT_ADDR="http://localhost:8200"
VAULT_TOKEN="dev-root"
ZITADEL_URL="http://localhost:8080"

echo "🔐 Zitadel admin token'ı otomatik oluşturuluyor..."

# Zitadel'in hazır olup olmadığını kontrol et
echo "⏳ Zitadel'in hazır olması bekleniyor..."
for i in {1..30}; do
    if curl -s "$ZITADEL_URL/debug/ready" > /dev/null 2>&1; then
        echo "✅ Zitadel hazır"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ Zitadel 30 saniye içinde hazır olmadı"
        exit 1
    fi
    sleep 1
done

# Login client PAT'ını kullanarak admin PAT oluştur
LOGIN_CLIENT_PAT_FILE="../zitadel/login-client.pat"

if [ ! -f "$LOGIN_CLIENT_PAT_FILE" ]; then
    echo "❌ Login client PAT dosyası bulunamadı: $LOGIN_CLIENT_PAT_FILE"
    echo "💡 Önce Zitadel'i başlatın: ../start-vault-zitadel.sh"
    exit 1
fi

LOGIN_CLIENT_PAT=$(cat "$LOGIN_CLIENT_PAT_FILE")

# Admin user ID'sini al
echo "🔍 Admin user bilgileri alınıyor..."
ADMIN_USER_RESPONSE=$(curl -s -X POST "$ZITADEL_URL/management/v1/users/_search" \
    -H "Authorization: Bearer $LOGIN_CLIENT_PAT" \
    -H "Content-Type: application/json" \
    -d '{
        "query": {
            "offset": "0",
            "limit": 10,
            "asc": true
        },
        "queries": [
            {
                "userNameQuery": {
                    "userName": "root@my-organization.localhost",
                    "method": "TEXT_QUERY_METHOD_EQUALS"
                }
            }
        ]
    }')

ADMIN_USER_ID=$(echo "$ADMIN_USER_RESPONSE" | jq -r '.result[0].id // empty')

if [ -z "$ADMIN_USER_ID" ]; then
    echo "❌ Admin user bulunamadı"
    echo "Response: $ADMIN_USER_RESPONSE"
    exit 1
fi

echo "✅ Admin user ID: $ADMIN_USER_ID"

# Admin için PAT oluştur
echo "🔑 Admin PAT oluşturuluyor..."
PAT_RESPONSE=$(curl -s -X POST "$ZITADEL_URL/management/v1/users/$ADMIN_USER_ID/personal_access_tokens" \
    -H "Authorization: Bearer $LOGIN_CLIENT_PAT" \
    -H "Content-Type: application/json" \
    -d '{
        "expirationDate": "2029-01-01T00:00:00Z",
        "scopes": [
            "openid",
            "profile",
            "email",
            "urn:zitadel:iam:org:project:id:zitadel:aud"
        ]
    }')

ADMIN_PAT=$(echo "$PAT_RESPONSE" | jq -r '.token // empty')

if [ -z "$ADMIN_PAT" ]; then
    echo "❌ Admin PAT oluşturulamadı"
    echo "Response: $PAT_RESPONSE"
    exit 1
fi

echo "✅ Admin PAT oluşturuldu"

# PAT'ı Vault'a kaydet
echo "💾 Admin PAT Vault'a kaydediliyor..."
vault kv put secret/zitadel/admin admin_pat="$ADMIN_PAT"

echo "✅ Admin PAT Vault'a kaydedildi"
echo "🚀 Şimdi './terraform/get-zitadel-token.sh' çalıştırabilirsiniz"