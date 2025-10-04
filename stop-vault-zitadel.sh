#!/bin/bash

echo "🛑 Stopping Vault and Zitadel services..."

# Stop Zitadel
echo "📦 Stopping Zitadel..."
cd zitadel
docker compose down

# Stop Vault
echo "📦 Stopping Vault..."
cd ../vault/dev
docker compose --profile dev down

echo "✅ All services stopped!"