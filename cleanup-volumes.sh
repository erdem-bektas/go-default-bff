#!/bin/bash

echo "🗑️  Vault ve Zitadel volume'larını temizleme scripti"
echo ""
echo "⚠️  UYARI: Bu işlem tüm Vault ve Zitadel verilerini kalıcı olarak silecek!"
echo ""

# Mevcut volume'ları listele
echo "📋 Mevcut volume'lar:"
docker volume ls | grep -E "(vault-dev-data|zitadel-postgres-data)" || echo "   Hiç volume bulunamadı"
echo ""

read -p "Devam etmek istediğinizden emin misiniz? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ İşlem iptal edildi"
    exit 1
fi

echo "🛑 Container'ları durduruluyor..."
cd vault/dev && docker compose down
cd ../../zitadel && docker compose down
cd ..

echo "🗑️  Volume'lar siliniyor..."
docker volume rm vault-dev-data zitadel-postgres-data 2>/dev/null || echo "   Bazı volume'lar zaten silinmiş"

echo "✅ Temizlik tamamlandı!"
echo "🚀 Yeniden başlatmak için: ./start-vault-zitadel.sh"