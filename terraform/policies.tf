# Login Policies for each organization
resource "zitadel_login_policy" "org_login_policies" {
  for_each = var.saas_organizations

  org_id                        = zitadel_org.saas_orgs[each.key].id
  user_login                    = each.value.features.login_policy_allow_username_password
  allow_register                = each.value.features.login_policy_allow_register
  allow_external_idp            = each.value.features.login_policy_allow_external_idp
  force_mfa                     = false
  force_mfa_local_only          = false
  passwordless_type             = "PASSWORDLESS_TYPE_NOT_ALLOWED"
  hide_password_reset           = false
  ignore_unknown_usernames      = false
  default_redirect_uri          = "http://${each.value.domain}/"
  password_check_lifetime       = "240h"
  external_login_check_lifetime = "240h"
  mfa_init_skip_lifetime        = "720h"
  second_factor_check_lifetime  = "18h"
  multi_factor_check_lifetime   = "12h"
}

# Password Complexity Policies
resource "zitadel_password_complexity_policy" "org_password_policies" {
  for_each = var.saas_organizations

  org_id        = zitadel_org.saas_orgs[each.key].id
  min_length    = each.value.features.password_complexity_policy.min_length
  has_uppercase = each.value.features.password_complexity_policy.has_uppercase
  has_lowercase = each.value.features.password_complexity_policy.has_lowercase
  has_number    = each.value.features.password_complexity_policy.has_number
  has_symbol    = each.value.features.password_complexity_policy.has_symbol
}

# Privacy Policies
resource "zitadel_privacy_policy" "org_privacy_policies" {
  for_each = var.saas_organizations

  org_id        = zitadel_org.saas_orgs[each.key].id
  tos_link      = "http://${each.value.domain}/terms"
  privacy_link  = "http://${each.value.domain}/privacy"
  help_link     = "http://${each.value.domain}/help"
  support_email = each.value.admin_email
}

# Lockout Policies
resource "zitadel_lockout_policy" "org_lockout_policies" {
  for_each = var.saas_organizations

  org_id                = zitadel_org.saas_orgs[each.key].id
  max_password_attempts = 5
}

# Domain Policies (Branding)
resource "zitadel_domain_policy" "org_domain_policies" {
  for_each = var.saas_organizations

  org_id                                      = zitadel_org.saas_orgs[each.key].id
  user_login_must_be_domain                   = false
  validate_org_domains                        = false
  smtp_sender_address_matches_instance_domain = false
}