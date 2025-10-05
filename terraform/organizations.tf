# Create Organizations for each SaaS project
resource "zitadel_org" "saas_orgs" {
  for_each = var.saas_organizations

  name = each.value.name
}

# Create Admin Users for each organization
resource "zitadel_human_user" "org_admins" {
  for_each = var.saas_organizations

  org_id = zitadel_org.saas_orgs[each.key].id

  user_name         = "admin"
  first_name        = "Admin"
  last_name         = "User"
  nick_name         = "Admin"
  display_name      = "Organization Admin"
  email             = each.value.admin_email
  is_email_verified = true

  initial_password = each.value.admin_password
}

# Grant ORG_OWNER role to admin users
resource "zitadel_org_member" "org_owners" {
  for_each = var.saas_organizations

  org_id  = zitadel_org.saas_orgs[each.key].id
  user_id = zitadel_human_user.org_admins[each.key].id
  roles   = ["ORG_OWNER"]
}

# Create OAuth Applications for each SaaS
resource "zitadel_application_oidc" "saas_apps" {
  for_each = var.saas_organizations

  org_id                      = zitadel_org.saas_orgs[each.key].id
  project_id                  = zitadel_project.saas_projects[each.key].id
  name                        = "${each.value.name} App"
  redirect_uris               = ["http://${each.value.domain}/auth/callback"]
  post_logout_redirect_uris   = ["http://${each.value.domain}/"]
  response_types              = ["OIDC_RESPONSE_TYPE_CODE"]
  grant_types                 = ["OIDC_GRANT_TYPE_AUTHORIZATION_CODE", "OIDC_GRANT_TYPE_REFRESH_TOKEN"]
  app_type                    = "OIDC_APP_TYPE_WEB"
  auth_method_type            = "OIDC_AUTH_METHOD_TYPE_BASIC"
  version                     = "OIDC_VERSION_1_0"
  dev_mode                    = true
  access_token_type           = "OIDC_TOKEN_TYPE_BEARER"
  access_token_role_assertion = false
  id_token_role_assertion     = false
  id_token_userinfo_assertion = false
  clock_skew                  = "0s"
  additional_origins          = []
}

# Create Projects for each SaaS
resource "zitadel_project" "saas_projects" {
  for_each = var.saas_organizations

  org_id                   = zitadel_org.saas_orgs[each.key].id
  name                     = "${each.value.name} Project"
  project_role_assertion   = true
  project_role_check       = true
  has_project_check        = true
  private_labeling_setting = "PRIVATE_LABELING_SETTING_ALLOW_LOGIN_USER_RESOURCE_OWNER_POLICY"
}

# Create custom roles for each project
resource "zitadel_project_role" "user_roles" {
  for_each = var.saas_organizations

  org_id       = zitadel_org.saas_orgs[each.key].id
  project_id   = zitadel_project.saas_projects[each.key].id
  role_key     = "user"
  display_name = "Standard User"
  group        = "users"
}

resource "zitadel_project_role" "admin_roles" {
  for_each = var.saas_organizations

  org_id       = zitadel_org.saas_orgs[each.key].id
  project_id   = zitadel_project.saas_projects[each.key].id
  role_key     = "admin"
  display_name = "Administrator"
  group        = "admins"
}