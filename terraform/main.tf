# Zitadel Provider Configuration - terraform bloğu terraform.tf'e taşındı

# Zitadel Provider - Variable'dan token alacak
provider "zitadel" {
  domain   = var.zitadel_domain
  insecure = var.zitadel_insecure
  port     = var.zitadel_port
  token    = var.zitadel_admin_token
}

# Variables
variable "zitadel_domain" {
  description = "Zitadel domain"
  type        = string
  default     = "localhost"
}

variable "zitadel_insecure" {
  description = "Use insecure connection"
  type        = bool
  default     = true
}

variable "zitadel_port" {
  description = "Zitadel port"
  type        = string
  default     = "8080"
}

variable "zitadel_admin_token" {
  description = "Zitadel admin token"
  type        = string
  sensitive   = true
  default     = "vEkmHTLRcebGEUttOsaeHD-WJX8hCQs7evoMPNimtBzXIt3ShytTOPS2m9JfpeGvMMBCV8c"
}

# SaaS Organizations Configuration
variable "saas_organizations" {
  description = "SaaS organizations to create"
  type = map(object({
    name           = string
    domain         = string
    admin_email    = string
    admin_password = string
    features = object({
      login_policy_allow_register          = bool
      login_policy_allow_username_password = bool
      login_policy_allow_external_idp      = bool
      password_complexity_policy = object({
        min_length    = number
        has_uppercase = bool
        has_lowercase = bool
        has_number    = bool
        has_symbol    = bool
      })
    })
  }))
  default = {
    sp1 = {
      name           = "SaaS Project 1"
      domain         = "sp1.localhost"
      admin_email    = "admin@sp1.localhost"
      admin_password = "SP1AdminPass123!"
      features = {
        login_policy_allow_register          = true
        login_policy_allow_username_password = true
        login_policy_allow_external_idp      = false
        password_complexity_policy = {
          min_length    = 8
          has_uppercase = true
          has_lowercase = true
          has_number    = true
          has_symbol    = true
        }
      }
    }
    sp2 = {
      name           = "SaaS Project 2"
      domain         = "sp2.localhost"
      admin_email    = "admin@sp2.localhost"
      admin_password = "SP2AdminPass123!"
      features = {
        login_policy_allow_register          = false
        login_policy_allow_username_password = true
        login_policy_allow_external_idp      = true
        password_complexity_policy = {
          min_length    = 10
          has_uppercase = true
          has_lowercase = true
          has_number    = true
          has_symbol    = true
        }
      }
    }
  }
}