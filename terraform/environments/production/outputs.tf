# Production Outputs

# Organization information
output "organizations" {
  description = "Created organizations information"
  value = {
    for key, org in zitadel_org.saas_orgs : key => {
      id     = org.id
      name   = org.name
      domain = local.saas_organizations[key].domain
      admin_email = local.saas_organizations[key].admin_email
      environment = "production"
    }
  }
}

# OAuth application credentials (sensitive)
output "oauth_applications" {
  description = "OAuth application credentials for each SaaS"
  value = {
    for key, app in zitadel_application_oidc.saas_apps : key => {
      client_id     = app.client_id
      project_id    = app.project_id
      org_id        = app.org_id
      redirect_uris = app.redirect_uris
      issuer_url    = "https://${var.zitadel_domain}"
      auth_url      = "https://${var.zitadel_domain}/oauth/v2/authorize"
      token_url     = "https://${var.zitadel_domain}/oauth/v2/token"
      userinfo_url  = "https://${var.zitadel_domain}/oidc/v1/userinfo"
      jwks_url      = "https://${var.zitadel_domain}/oauth/v2/keys"
      domain        = local.saas_organizations[key].domain
      environment   = "production"
    }
  }
  sensitive = true
}

# Project information
output "projects" {
  description = "Created projects information"
  value = {
    for key, project in zitadel_project.saas_projects : key => {
      id     = project.id
      name   = project.name
      org_id = project.org_id
    }
  }
}

# Admin user information
output "admin_users" {
  description = "Admin users for each organization"
  value = {
    for key, user in zitadel_human_user.org_admins : key => {
      id       = user.id
      username = user.user_name
      email    = user.email
      org_id   = user.org_id
      login_name = "${user.user_name}@${local.saas_organizations[key].name}.${var.zitadel_domain}"
    }
  }
}

# Vault paths for credentials
output "vault_paths" {
  description = "Vault paths where credentials are stored"
  value = {
    for key, _ in local.saas_organizations : key => {
      oauth_path = "secret/production/saas/${key}/oauth"
      vault_ui_url = "${var.vault_address}/ui/vault/secrets/secret/show/production/saas/${key}/oauth"
    }
  }
}

# Production URLs
output "production_urls" {
  description = "Production URLs for each SaaS"
  value = {
    for key, _ in local.saas_organizations : key => {
      domain = local.saas_organizations[key].domain
      app_url = "https://${local.saas_organizations[key].domain}"
      api_url = "https://api.${local.saas_organizations[key].domain}"
      admin_url = "https://admin.${local.saas_organizations[key].domain}"
      auth_url = "https://${var.zitadel_domain}/oauth/v2/authorize"
      zitadel_console = "https://${var.zitadel_domain}/ui/console/orgs/${zitadel_org.saas_orgs[key].id}"
    }
  }
}

# Security summary
output "security_summary" {
  description = "Security configuration summary"
  value = {
    mfa_enabled = true
    password_min_length = 12
    password_expiry_days = 90
    max_login_attempts = 3
    domain_validation = true
    https_only = true
    environment = "production"
  }
}

# Deployment information
output "deployment_info" {
  description = "Deployment information"
  value = {
    terraform_workspace = terraform.workspace
    zitadel_domain = var.zitadel_domain
    vault_address = var.vault_address
    hetzner_server_ip = var.hetzner_server_ip
    total_organizations = length(local.saas_organizations)
    deployment_time = timestamp()
  }
}