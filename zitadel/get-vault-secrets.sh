#!/bin/bash

# Vault'tan Zitadel credentials'larını çek ve environment dosyası oluştur
VAULT_ADDR="http://localhost:8200"
VAULT_TOKEN="dev-root"

echo "Fetching Zitadel credentials from Vault..."

# Database credentials
DB_SECRETS=$(curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/secret/data/zitadel/database" | jq -r '.data.data')
CONFIG_SECRETS=$(curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/secret/data/zitadel/config" | jq -r '.data.data')
FIRSTINSTANCE_SECRETS=$(curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/secret/data/zitadel/firstinstance" | jq -r '.data.data')

# .env dosyası oluştur
cat > .env << EOF
# Database Configuration
POSTGRES_HOST=$(echo $DB_SECRETS | jq -r '.postgres_host')
POSTGRES_PORT=$(echo $DB_SECRETS | jq -r '.postgres_port')
POSTGRES_DATABASE=$(echo $DB_SECRETS | jq -r '.postgres_database')
POSTGRES_ADMIN_USERNAME=$(echo $DB_SECRETS | jq -r '.postgres_admin_username')
POSTGRES_ADMIN_PASSWORD=$(echo $DB_SECRETS | jq -r '.postgres_admin_password')
POSTGRES_USER_USERNAME=$(echo $DB_SECRETS | jq -r '.postgres_user_username')
POSTGRES_USER_PASSWORD=$(echo $DB_SECRETS | jq -r '.postgres_user_password')

# Zitadel Configuration
ZITADEL_MASTERKEY=$(echo $CONFIG_SECRETS | jq -r '.masterkey')
ZITADEL_EXTERNAL_DOMAIN=$(echo $CONFIG_SECRETS | jq -r '.external_domain')
ZITADEL_EXTERNAL_SECURE=$(echo $CONFIG_SECRETS | jq -r '.external_secure')
ZITADEL_TLS_ENABLED=$(echo $CONFIG_SECRETS | jq -r '.tls_enabled')

# First Instance Configuration
ZITADEL_ORG_NAME=$(echo $FIRSTINSTANCE_SECRETS | jq -r '.org_name')
ZITADEL_ORG_HUMAN_USERNAME=$(echo $FIRSTINSTANCE_SECRETS | jq -r '.org_human_username')
ZITADEL_ORG_HUMAN_PASSWORD=$(echo $FIRSTINSTANCE_SECRETS | jq -r '.org_human_password')
ZITADEL_LOGIN_CLIENT_USERNAME=$(echo $FIRSTINSTANCE_SECRETS | jq -r '.login_client_username')
ZITADEL_LOGIN_CLIENT_NAME=$(echo $FIRSTINSTANCE_SECRETS | jq -r '.login_client_name')
ZITADEL_PAT_EXPIRATION_DATE=$(echo $FIRSTINSTANCE_SECRETS | jq -r '.pat_expiration_date')
EOF

echo "Environment file created successfully!"
echo "You can now run: docker compose up -d"