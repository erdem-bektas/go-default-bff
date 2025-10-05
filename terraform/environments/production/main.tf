# Production Terraform Configuration for Hetzner Server
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    zitadel = {
      source  = "zitadel/zitadel"
      version = "~> 1.0"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "~> 3.0"
    }
  }
  
  # Remote backend for production state
  backend "s3" {
    # Hetzner Object Storage veya AWS S3 kullanabilirsiniz
    bucket = "your-terraform-state-bucket"
    key    = "zitadel-saas/production/terraform.tfstate"
    region = "eu-central-1"
    
    # Hetzner Object Storage için
    # endpoint = "https://fsn1.your-project.hetzner-objects.com"
    # skip_credentials_validation = true
    # skip_metadata_api_check = true
    # skip_region_validation = true
  }
}

# Production Variables
variable "zitadel_domain" {
  description = "Production Zitadel domain"
  type        = string
  default     = "auth.yourdomain.com"
}

variable "zitadel_port" {
  description = "Zitadel port (443 for HTTPS)"
  type        = string
  default     = "443"
}

variable "zitadel_insecure" {
  description = "Use insecure connection (false for production)"
  type        = bool
  default     = false
}

variable "vault_address" {
  description = "Vault server address"
  type        = string
  default     = "https://vault.yourdomain.com"
}

variable "hetzner_server_ip" {
  description = "Hetzner server IP address"
  type        = string
}

# Vault Provider for credential management
provider "vault" {
  address = var.vault_address
  # Token will be provided via VAULT_TOKEN environment variable
}

# Get Zitadel admin token from Vault
data "vault_generic_secret" "zitadel_admin" {
  path = "secret/zitadel/admin"
}

# Zitadel Provider
provider "zitadel" {
  domain           = var.zitadel_domain
  insecure         = var.zitadel_insecure
  port             = var.zitadel_port
  token            = data.vault_generic_secret.zitadel_admin.data["admin_pat"]
}

# Production SaaS Organizations
variable "production_saas_organizations" {
  description = "Production SaaS organizations"
  type = map(object({
    name        = string
    domain      = string
    admin_email = string
    admin_password = string
    features = object({
      login_policy_allow_register = bool
      login_policy_allow_username_password = bool
      login_policy_allow_external_idp = bool
      password_complexity_policy = object({
        min_length    = number
        has_uppercase = bool
        has_lowercase = bool
        has_number    = bool
        has_symbol    = bool
      })
    })
  }))
  
  # Production organizasyonları - güvenli şifreler kullanın
  default = {
    saas1 = {
      name        = "Production SaaS 1"
      domain      = "saas1.yourdomain.com"
      admin_email = "admin@saas1.yourdomain.com"
      admin_password = "ProductionSecurePassword123!@#"
      features = {
        login_policy_allow_register = false  # Production'da kapalı
        login_policy_allow_username_password = true
        login_policy_allow_external_idp = true
        password_complexity_policy = {
          min_length    = 12  # Production'da daha güçlü
          has_uppercase = true
          has_lowercase = true
          has_number    = true
          has_symbol    = true
        }
      }
    }
    saas2 = {
      name        = "Production SaaS 2"
      domain      = "saas2.yourdomain.com"
      admin_email = "admin@saas2.yourdomain.com"
      admin_password = "ProductionSecurePassword456!@#"
      features = {
        login_policy_allow_register = false
        login_policy_allow_username_password = true
        login_policy_allow_external_idp = true
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

# Use production organizations
locals {
  saas_organizations = var.production_saas_organizations
}