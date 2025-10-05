# Production Organizations Configuration

# Create Organizations for each SaaS project
resource "zitadel_org" "saas_orgs" {
  for_each = local.saas_organizations
  
  name = each.value.name
}

# Create Admin Users for each organization
resource "zitadel_human_user" "org_admins" {
  for_each = local.saas_organizations
  
  org_id = zitadel_org.saas_orgs[each.key].id
  
  user_name    = "admin"
  first_name   = "Admin"
  last_name    = "User"
  nick_name    = "Admin"
  display_name = "Organization Admin"
  email        = each.value.admin_email
  is_email_verified = true
  
  initial_password = each.value.admin_password
  password_change_required = false
}

# Grant ORG_OWNER role to admin users
resource "zitadel_org_member" "org_owners" {
  for_each = local.saas_organizations
  
  org_id  = zitadel_org.saas_orgs[each.key].id
  user_id = zitadel_human_user.org_admins[each.key].id
  roles   = ["ORG_OWNER"]
}

# Create Projects for each SaaS
resource "zitadel_project" "saas_projects" {
  for_each = local.saas_organizations
  
  org_id                   = zitadel_org.saas_orgs[each.key].id
  name                     = "${each.value.name} Project"
  project_role_assertion   = true
  project_role_check       = true
  has_project_check        = true
  private_labeling_setting = "PRIVATE_LABELING_SETTING_ALLOW_LOGIN_USER_RESOURCE_OWNER_POLICY"
}

# Create OAuth Applications for each SaaS (Production URLs)
resource "zitadel_application_oidc" "saas_apps" {
  for_each = local.saas_organizations
  
  org_id                      = zitadel_org.saas_orgs[each.key].id
  project_id                  = zitadel_project.saas_projects[each.key].id
  name                        = "${each.value.name} App"
  redirect_uris               = [
    "https://${each.value.domain}/auth/callback",
    "https://api.${each.value.domain}/auth/callback"
  ]
  post_logout_redirect_uris   = [
    "https://${each.value.domain}/",
    "https://${each.value.domain}/logout"
  ]
  response_types              = ["OIDC_RESPONSE_TYPE_CODE"]
  grant_types                 = ["OIDC_GRANT_TYPE_AUTHORIZATION_CODE", "OIDC_GRANT_TYPE_REFRESH_TOKEN"]
  app_type                    = "OIDC_APP_TYPE_WEB"
  auth_method_type            = "OIDC_AUTH_METHOD_TYPE_BASIC"
  version                     = "OIDC_VERSION_1_0"
  dev_mode                    = false  # Production'da false
  access_token_type           = "OIDC_TOKEN_TYPE_BEARER"
  access_token_role_assertion = true   # Production'da role assertion enable
  id_token_role_assertion     = true
  id_token_userinfo_assertion = false
  clock_skew                  = "0s"
  additional_origins          = [
    "https://${each.value.domain}",
    "https://api.${each.value.domain}",
    "https://admin.${each.value.domain}"
  ]
}

# Create custom roles for each project
resource "zitadel_project_role" "user_roles" {
  for_each = local.saas_organizations
  
  org_id     = zitadel_org.saas_orgs[each.key].id
  project_id = zitadel_project.saas_projects[each.key].id
  role_key   = "user"
  display_name = "Standard User"
  group       = "users"
}

resource "zitadel_project_role" "admin_roles" {
  for_each = local.saas_organizations
  
  org_id     = zitadel_org.saas_orgs[each.key].id
  project_id = zitadel_project.saas_projects[each.key].id
  role_key   = "admin"
  display_name = "Administrator"
  group       = "admins"
}

resource "zitadel_project_role" "manager_roles" {
  for_each = local.saas_organizations
  
  org_id     = zitadel_org.saas_orgs[each.key].id
  project_id = zitadel_project.saas_projects[each.key].id
  role_key   = "manager"
  display_name = "Manager"
  group       = "managers"
}

# Store OAuth credentials in Vault
resource "vault_generic_secret" "saas_oauth_credentials" {
  for_each = local.saas_organizations
  
  path = "secret/production/saas/${each.key}/oauth"
  
  data_json = jsonencode({
    org_id        = zitadel_org.saas_orgs[each.key].id
    project_id    = zitadel_project.saas_projects[each.key].id
    client_id     = zitadel_application_oidc.saas_apps[each.key].client_id
    client_secret = zitadel_application_oidc.saas_apps[each.key].client_secret
    issuer_url    = "https://${var.zitadel_domain}"
    auth_url      = "https://${var.zitadel_domain}/oauth/v2/authorize"
    token_url     = "https://${var.zitadel_domain}/oauth/v2/token"
    userinfo_url  = "https://${var.zitadel_domain}/oidc/v1/userinfo"
    jwks_url      = "https://${var.zitadel_domain}/oauth/v2/keys"
    domain        = each.value.domain
    environment   = "production"
  })
}