#!/bin/bash

# SaaS organizasyonlarını manuel olarak oluştur
set -e

ZITADEL_URL="http://localhost:8080"
ADMIN_TOKEN="vEkmHTLRcebGEUttOsaeHD-WJX8hCQs7evoMPNimtBzXIt3ShytTOPS2m9JfpeGvMMBCV8c"

echo "🏢 SaaS organizasyonları oluşturuluyor..."

# SaaS Project 1 organizasyonu oluştur
echo "📋 SaaS Project 1 organizasyonu oluşturuluyor..."
SP1_ORG_RESPONSE=$(curl -s -X POST "$ZITADEL_URL/management/v1/orgs" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "SaaS Project 1"
    }')

SP1_ORG_ID=$(echo "$SP1_ORG_RESPONSE" | jq -r '.id // empty')

if [ -z "$SP1_ORG_ID" ]; then
    echo "❌ SaaS Project 1 organizasyonu oluşturulamadı"
    echo "Response: $SP1_ORG_RESPONSE"
    exit 1
fi

echo "✅ SaaS Project 1 organizasyonu oluşturuldu: $SP1_ORG_ID"

# SaaS Project 2 organizasyonu oluştur
echo "📋 SaaS Project 2 organizasyonu oluşturuluyor..."
SP2_ORG_RESPONSE=$(curl -s -X POST "$ZITADEL_URL/management/v1/orgs" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "SaaS Project 2"
    }')

SP2_ORG_ID=$(echo "$SP2_ORG_RESPONSE" | jq -r '.id // empty')

if [ -z "$SP2_ORG_ID" ]; then
    echo "❌ SaaS Project 2 organizasyonu oluşturulamadı"
    echo "Response: $SP2_ORG_RESPONSE"
    exit 1
fi

echo "✅ SaaS Project 2 organizasyonu oluşturuldu: $SP2_ORG_ID"

echo ""
echo "🎉 Tüm SaaS organizasyonları başarıyla oluşturuldu!"
echo ""
echo "📊 Organizasyon Bilgileri:"
echo "   - SaaS Project 1 ID: $SP1_ORG_ID"
echo "   - SaaS Project 2 ID: $SP2_ORG_ID"
echo ""
echo "🌐 Zitadel Console'da kontrol edin: http://localhost:8080/ui/console"