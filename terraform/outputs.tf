# Output organization information
output "organizations" {
  description = "Created organizations information"
  value = {
    for key, org in zitadel_org.saas_orgs : key => {
      id          = org.id
      name        = org.name
      domain      = var.saas_organizations[key].domain
      admin_email = var.saas_organizations[key].admin_email
    }
  }
}

# Output OAuth application credentials
output "oauth_applications" {
  description = "OAuth application credentials for each SaaS"
  value = {
    for key, app in zitadel_application_oidc.saas_apps : key => {
      client_id     = app.client_id
      client_secret = app.client_secret
      project_id    = app.project_id
      org_id        = app.org_id
      redirect_uris = app.redirect_uris
      issuer_url    = "http://${var.zitadel_domain}:${var.zitadel_port}"
      auth_url      = "http://${var.zitadel_domain}:${var.zitadel_port}/oauth/v2/authorize"
      token_url     = "http://${var.zitadel_domain}:${var.zitadel_port}/oauth/v2/token"
      userinfo_url  = "http://${var.zitadel_domain}:${var.zitadel_port}/oidc/v1/userinfo"
    }
  }
  sensitive = true
}

# Output project information
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

# Output admin user information
output "admin_users" {
  description = "Admin users for each organization"
  value = {
    for key, user in zitadel_human_user.org_admins : key => {
      id         = user.id
      username   = user.user_name
      email      = user.email
      org_id     = user.org_id
      login_name = "${user.user_name}@${var.saas_organizations[key].name}.${var.zitadel_domain}"
    }
  }
}