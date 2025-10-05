#!/bin/bash

echo "🚀 Starting Vault and Zitadel with Vault integration..."

# 1. Start Vault dev
echo "📦 Starting Vault dev environment..."
cd vault/dev
docker compose --profile dev up -d
echo "⏳ Waiting for Vault to be ready..."
sleep 5

# 2. Setup Zitadel secrets in Vault (if not already exists)
echo "🔐 Setting up Zitadel secrets in Vault..."
docker exec -i vault-dev sh -c '
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="dev-root"

# Check if secrets already exist
if ! vault kv get secret/zitadel/database >/dev/null 2>&1; then
    echo "Creating Zitadel secrets..."
    vault kv put secret/zitadel/database \
        postgres_host=db \
        postgres_port=5432 \
        postgres_database=zitadel \
        postgres_admin_username=postgres \
        postgres_admin_password=postgres \
        postgres_user_username=zitadel \
        postgres_user_password=zitadel

    vault kv put secret/zitadel/config \
        masterkey="MasterkeyNeedsToHave32Characters" \
        external_domain=localhost \
        external_secure=false \
        tls_enabled=false

    vault kv put secret/zitadel/firstinstance \
        org_name="My Organization" \
        org_human_username=root \
        org_human_password="RootPassword1!" \
        login_client_username=login-client \
        login_client_name="Automatically Initialized IAM_LOGIN_CLIENT" \
        pat_expiration_date="2029-01-01T00:00:00Z"
else
    echo "Zitadel secrets already exist in Vault"
fi
'

# 3. Start Zitadel with Vault credentials
echo "🔄 Fetching credentials from Vault and starting Zitadel..."
cd ../../zitadel
./get-vault-secrets.sh
docker compose up -d

echo ""
echo "✅ Setup complete!"
echo ""
echo "🌐 Services available at:"
echo "   - Vault UI: http://localhost:8200 (token: dev-root)"
echo "   - Zitadel Console: http://localhost:8080/ui/console (👈 Buraya gidin!)"
echo "   - Zitadel API: http://localhost:8080"
echo ""
echo "🔑 Zitadel Admin Login (Console'da kullanın):"
echo "   - Username: root@my-organization.localhost"
echo "   - Password: RootPassword1!"
echo ""
echo "ℹ️  Not: Login v2 (port 3000) sadece auth request'ler için kullanılır"
echo ""
echo "📊 Check status:"
echo "   docker ps"
echo ""
echo "💾 Named volumes created for persistent data:"
echo "   - vault-dev-data (Vault development data)"
echo "   - zitadel-postgres-data (Zitadel database)"
echo ""
echo "🗑️  To clean up volumes (WARNING: This will delete all data):"
echo "   docker volume rm vault-dev-data zitadel-postgres-data"