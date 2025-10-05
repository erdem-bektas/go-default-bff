#!/bin/bash

# Production SaaS Organization Management Script for Hetzner
# Bu script production Zitadel'de SaaS organizasyonlarını yönetir

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform/environments/production"

# Renkli output için
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Production SaaS Manager${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if we're on the Hetzner server or local machine
check_environment() {
    if [ -f "/opt/zitadel-saas/.env.production" ]; then
        ENVIRONMENT="hetzner"
        ENV_FILE="/opt/zitadel-saas/.env.production"
        print_info "Running on Hetzner server"
    else
        ENVIRONMENT="local"
        ENV_FILE="$SCRIPT_DIR/.env.production"
        print_info "Running on local machine"
    fi
}

# Load environment variables
load_environment() {
    if [ ! -f "$ENV_FILE" ]; then
        print_error "Environment file not found: $ENV_FILE"
        print_info "Please create it from .env.production.example"
        exit 1
    fi
    
    source "$ENV_FILE"
    print_success "Environment loaded from $ENV_FILE"
}

check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Terraform kontrolü
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform not installed. Please install: https://terraform.io/downloads"
        exit 1
    fi
    
    # Vault CLI kontrolü
    if ! command -v vault &> /dev/null; then
        print_error "Vault CLI not installed. Please install."
        exit 1
    fi
    
    # Environment variables kontrolü
    if [ -z "$ZITADEL_EXTERNAL_DOMAIN" ]; then
        print_error "ZITADEL_EXTERNAL_DOMAIN not set in environment"
        exit 1
    fi
    
    if [ -z "$VAULT_ROOT_TOKEN" ]; then
        print_error "VAULT_ROOT_TOKEN not set in environment"
        exit 1
    fi
    
    print_success "All prerequisites met"
}

# Configure Vault connection
configure_vault() {
    print_info "Configuring Vault connection..."
    
    if [ "$ENVIRONMENT" = "hetzner" ]; then
        export VAULT_ADDR="https://vault.${ZITADEL_EXTERNAL_DOMAIN}"
    else
        export VAULT_ADDR="https://vault.yourdomain.com"  # Update with your domain
    fi
    
    export VAULT_TOKEN="$VAULT_ROOT_TOKEN"
    
    # Test Vault connection
    if ! vault status &> /dev/null; then
        print_error "Cannot connect to Vault at $VAULT_ADDR"
        print_info "Make sure Vault is running and accessible"
        exit 1
    fi
    
    print_success "Vault connection configured"
}

# Setup admin token in Vault
setup_admin_token() {
    print_info "Setting up admin token..."
    
    # Check if admin token already exists
    if vault kv get secret/zitadel/admin &> /dev/null; then
        print_info "Admin token already exists in Vault"
        return 0
    fi
    
    print_warning "Admin token not found in Vault"
    print_info "You need to manually create an admin PAT in Zitadel console"
    print_info "1. Go to https://${ZITADEL_EXTERNAL_DOMAIN}/ui/console"
    print_info "2. Login with admin credentials"
    print_info "3. Go to Organization Settings > Personal Access Tokens"
    print_info "4. Create a new token with all permissions"
    
    read -p "Enter the admin PAT: " ADMIN_PAT
    
    if [ -z "$ADMIN_PAT" ]; then
        print_error "Admin PAT cannot be empty"
        exit 1
    fi
    
    # Store in Vault
    vault kv put secret/zitadel/admin admin_pat="$ADMIN_PAT"
    print_success "Admin token stored in Vault"
}

# Prepare Terraform
prepare_terraform() {
    print_info "Preparing Terraform..."
    
    cd "$TERRAFORM_DIR"
    
    # Create terraform.tfvars if it doesn't exist
    if [ ! -f "terraform.tfvars" ]; then
        print_info "Creating terraform.tfvars from example"
        cp terraform.tfvars.example terraform.tfvars
        
        # Update with environment variables
        sed -i "s/YOUR_HETZNER_SERVER_IP/$(curl -s ifconfig.me)/g" terraform.tfvars
        sed -i "s/auth.yourdomain.com/$ZITADEL_EXTERNAL_DOMAIN/g" terraform.tfvars
        sed -i "s/vault.yourdomain.com/vault.${ZITADEL_EXTERNAL_DOMAIN}/g" terraform.tfvars
        
        print_warning "Please review and update terraform.tfvars with your specific values"
    fi
    
    # Initialize Terraform
    if ! terraform init; then
        print_error "Terraform init failed"
        exit 1
    fi
    
    print_success "Terraform prepared"
}

# Plan organizations
plan_organizations() {
    print_info "Planning organization changes..."
    
    cd "$TERRAFORM_DIR"
    
    # Set Terraform environment variables
    export TF_VAR_vault_address="$VAULT_ADDR"
    export VAULT_TOKEN="$VAULT_ROOT_TOKEN"
    
    if ! terraform plan; then
        print_error "Terraform plan failed"
        exit 1
    fi
    
    print_success "Plan completed"
}

# Create organizations
create_organizations() {
    print_info "Creating organizations..."
    
    cd "$TERRAFORM_DIR"
    
    # Set Terraform environment variables
    export TF_VAR_vault_address="$VAULT_ADDR"
    export VAULT_TOKEN="$VAULT_ROOT_TOKEN"
    
    if ! terraform apply -auto-approve; then
        print_error "Terraform apply failed"
        exit 1
    fi
    
    print_success "Organizations created"
}

# Show results
show_results() {
    print_info "Showing results..."
    
    cd "$TERRAFORM_DIR"
    
    echo ""
    echo -e "${BLUE}📋 Created Organizations:${NC}"
    terraform output -json organizations | jq -r 'to_entries[] | "• \(.key): \(.value.name) (\(.value.domain))"'
    
    echo ""
    echo -e "${BLUE}🔗 Production URLs:${NC}"
    terraform output -json production_urls | jq -r 'to_entries[] | "• \(.key): \(.value.app_url)"'
    
    echo ""
    echo -e "${BLUE}🔐 Vault Paths:${NC}"
    terraform output -json vault_paths | jq -r 'to_entries[] | "• \(.key): \(.value.oauth_path)"'
    
    echo ""
    print_info "Detailed output: terraform output"
}

# Destroy organizations
destroy_organizations() {
    print_warning "This will destroy ALL production organizations!"
    read -p "Are you absolutely sure? Type 'yes' to confirm: " -r
    
    if [ "$REPLY" != "yes" ]; then
        print_info "Operation cancelled"
        exit 0
    fi
    
    cd "$TERRAFORM_DIR"
    
    export TF_VAR_vault_address="$VAULT_ADDR"
    export VAULT_TOKEN="$VAULT_ROOT_TOKEN"
    
    if ! terraform destroy -auto-approve; then
        print_error "Terraform destroy failed"
        exit 1
    fi
    
    print_success "Organizations destroyed"
}

# Show status
show_status() {
    cd "$TERRAFORM_DIR"
    
    if [ -f "terraform.tfstate" ]; then
        show_results
    else
        print_info "No organizations created yet"
    fi
    
    echo ""
    echo -e "${BLUE}🏥 Service Health:${NC}"
    
    # Check Zitadel
    if curl -s "https://${ZITADEL_EXTERNAL_DOMAIN}/debug/ready" &> /dev/null; then
        echo -e "• Zitadel: ${GREEN}✅ Healthy${NC}"
    else
        echo -e "• Zitadel: ${RED}❌ Unhealthy${NC}"
    fi
    
    # Check Vault
    if vault status &> /dev/null; then
        echo -e "• Vault: ${GREEN}✅ Healthy${NC}"
    else
        echo -e "• Vault: ${RED}❌ Unhealthy${NC}"
    fi
}

# Show help
show_help() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  create    - Create SaaS organizations (default)"
    echo "  plan      - Show plan only"
    echo "  destroy   - Destroy all organizations"
    echo "  status    - Show current status"
    echo "  help      - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0                # Create organizations"
    echo "  $0 plan          # Show plan"
    echo "  $0 status        # Show status"
    echo ""
    echo "Environment:"
    echo "  Local:    Uses .env.production in script directory"
    echo "  Hetzner:  Uses /opt/zitadel-saas/.env.production"
}

# Main function
main() {
    print_header
    
    case "${1:-create}" in
        "create")
            check_environment
            load_environment
            check_prerequisites
            configure_vault
            setup_admin_token
            prepare_terraform
            plan_organizations
            create_organizations
            show_results
            print_success "Production SaaS organizations created successfully!"
            ;;
        "plan")
            check_environment
            load_environment
            check_prerequisites
            configure_vault
            setup_admin_token
            prepare_terraform
            plan_organizations
            ;;
        "destroy")
            check_environment
            load_environment
            check_prerequisites
            configure_vault
            prepare_terraform
            destroy_organizations
            ;;
        "status")
            check_environment
            load_environment
            configure_vault
            show_status
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"