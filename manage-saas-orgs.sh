#!/bin/bash

# SaaS Organization Management Script
# Bu script Zitadel'de SaaS organizasyonlarını otomatik yönetir

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/terraform"

# Renkli output için
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  SaaS Organization Manager${NC}"
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

check_prerequisites() {
    print_info "Ön koşullar kontrol ediliyor..."
    
    # Terraform kontrolü
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform yüklü değil. Lütfen yükleyin: https://terraform.io/downloads"
        exit 1
    fi
    
    # Vault kontrolü
    if ! command -v vault &> /dev/null; then
        print_error "Vault CLI yüklü değil. Lütfen yükleyin."
        exit 1
    fi
    
    # jq kontrolü
    if ! command -v jq &> /dev/null; then
        print_error "jq yüklü değil. Lütfen yükleyin: brew install jq"
        exit 1
    fi
    
    # Vault erişim kontrolü
    if ! vault status &> /dev/null; then
        print_error "Vault'a erişilemiyor. Vault'un çalıştığından emin olun."
        exit 1
    fi
    
    # Zitadel erişim kontrolü
    if ! curl -s http://localhost:8080/debug/ready &> /dev/null; then
        print_error "Zitadel'e erişilemiyor. Zitadel'in çalıştığından emin olun."
        print_info "Başlatmak için: ./start-vault-zitadel.sh"
        exit 1
    fi
    
    print_success "Tüm ön koşullar karşılanıyor"
}

setup_admin_token() {
    print_info "Admin token kurulumu..."
    
    if ! "$SCRIPT_DIR/terraform/setup-admin-token.sh"; then
        print_error "Admin token kurulumu başarısız"
        exit 1
    fi
    
    print_success "Admin token kuruldu"
}

prepare_terraform() {
    print_info "Terraform hazırlanıyor..."
    
    cd "$TERRAFORM_DIR"
    
    # Terraform variables hazırla
    if ! "$SCRIPT_DIR/terraform/get-zitadel-token.sh"; then
        print_error "Terraform variables hazırlanamadı"
        exit 1
    fi
    
    # Terraform init
    if ! terraform init; then
        print_error "Terraform init başarısız"
        exit 1
    fi
    
    print_success "Terraform hazırlandı"
}

plan_organizations() {
    print_info "Organization planı oluşturuluyor..."
    
    cd "$TERRAFORM_DIR"
    
    if ! terraform plan; then
        print_error "Terraform plan başarısız"
        exit 1
    fi
    
    print_success "Plan oluşturuldu"
}

create_organizations() {
    print_info "Organizasyonlar oluşturuluyor..."
    
    cd "$TERRAFORM_DIR"
    
    if ! terraform apply -auto-approve; then
        print_error "Terraform apply başarısız"
        exit 1
    fi
    
    print_success "Organizasyonlar oluşturuldu"
}

show_results() {
    print_info "Sonuçlar gösteriliyor..."
    
    cd "$TERRAFORM_DIR"
    
    echo ""
    echo -e "${BLUE}📋 Oluşturulan Organizasyonlar:${NC}"
    terraform output -json organizations | jq -r 'to_entries[] | "• \(.key): \(.value.name) (\(.value.domain))"'
    
    echo ""
    echo -e "${BLUE}🔑 OAuth Uygulamaları:${NC}"
    terraform output -json oauth_applications | jq -r 'to_entries[] | "• \(.key): Client ID = \(.value.client_id)"'
    
    echo ""
    echo -e "${BLUE}👤 Admin Kullanıcıları:${NC}"
    terraform output -json admin_users | jq -r 'to_entries[] | "• \(.key): \(.value.login_name)"'
    
    echo ""
    print_info "Detaylı bilgi için: cd terraform && terraform output"
}

save_credentials() {
    print_info "Kimlik bilgileri Vault'a kaydediliyor..."
    
    cd "$TERRAFORM_DIR"
    
    # OAuth credentials'ları Vault'a kaydet
    terraform output -json oauth_applications | jq -r 'to_entries[] | @base64' | while read -r line; do
        data=$(echo "$line" | base64 -d)
        key=$(echo "$data" | jq -r '.key')
        client_id=$(echo "$data" | jq -r '.value.client_id')
        client_secret=$(echo "$data" | jq -r '.value.client_secret')
        org_id=$(echo "$data" | jq -r '.value.org_id')
        project_id=$(echo "$data" | jq -r '.value.project_id')
        
        vault kv put "secret/saas/$key/oauth" \
            client_id="$client_id" \
            client_secret="$client_secret" \
            org_id="$org_id" \
            project_id="$project_id" \
            issuer_url="http://localhost:8080" \
            auth_url="http://localhost:8080/oauth/v2/authorize" \
            token_url="http://localhost:8080/oauth/v2/token" \
            userinfo_url="http://localhost:8080/oidc/v1/userinfo"
    done
    
    print_success "Kimlik bilgileri Vault'a kaydedildi"
}

destroy_organizations() {
    print_warning "Tüm organizasyonlar silinecek!"
    read -p "Devam etmek istediğinizden emin misiniz? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "İşlem iptal edildi"
        exit 0
    fi
    
    cd "$TERRAFORM_DIR"
    
    if ! terraform destroy -auto-approve; then
        print_error "Terraform destroy başarısız"
        exit 1
    fi
    
    print_success "Organizasyonlar silindi"
}

show_help() {
    echo "Kullanım: $0 [komut]"
    echo ""
    echo "Komutlar:"
    echo "  create    - Organizasyonları oluştur (varsayılan)"
    echo "  plan      - Sadece planı göster"
    echo "  destroy   - Organizasyonları sil"
    echo "  status    - Mevcut durumu göster"
    echo "  help      - Bu yardımı göster"
    echo ""
    echo "Örnekler:"
    echo "  $0                # Organizasyonları oluştur"
    echo "  $0 plan          # Planı göster"
    echo "  $0 destroy       # Organizasyonları sil"
}

main() {
    print_header
    
    case "${1:-create}" in
        "create")
            check_prerequisites
            setup_admin_token
            prepare_terraform
            plan_organizations
            create_organizations
            show_results
            save_credentials
            print_success "SaaS organizasyonları başarıyla oluşturuldu!"
            ;;
        "plan")
            check_prerequisites
            setup_admin_token
            prepare_terraform
            plan_organizations
            ;;
        "destroy")
            check_prerequisites
            prepare_terraform
            destroy_organizations
            ;;
        "status")
            cd "$TERRAFORM_DIR"
            if [ -f "terraform.tfstate" ]; then
                show_results
            else
                print_info "Henüz organizasyon oluşturulmamış"
            fi
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Bilinmeyen komut: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"