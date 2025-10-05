# Production Security Policies

# Login Policies for each organization (Production settings)
resource "zitadel_login_policy" "org_login_policies" {
  for_each = local.saas_organizations
  
  org_id                      = zitadel_org.saas_orgs[each.key].id
  allow_username_password     = each.value.features.login_policy_allow_username_password
  allow_register              = each.value.features.login_policy_allow_register
  allow_external_idp          = each.value.features.login_policy_allow_external_idp
  force_mfa                   = true   # Production'da MFA zorunlu
  passwordless_type          = "PASSWORDLESS_TYPE_ALLOWED"  # Passwordless enable
  hide_password_reset        = false
  ignore_unknown_usernames   = true    # Security: don't reveal if user exists
  default_redirect_uri       = "https://${each.value.domain}/"
  
  # Multi-factor authentication settings
  force_mfa_local_only       = false
  
  # Session settings
  multi_factors              = ["MFA_TYPE_OTP_SMS", "MFA_TYPE_OTP_EMAIL", "MFA_TYPE_TOTP"]
}

# Password Complexity Policies (Strong for production)
resource "zitadel_password_complexity_policy" "org_password_policies" {
  for_each = local.saas_organizations
  
  org_id       = zitadel_org.saas_orgs[each.key].id
  min_length   = each.value.features.password_complexity_policy.min_length
  has_uppercase = each.value.features.password_complexity_policy.has_uppercase
  has_lowercase = each.value.features.password_complexity_policy.has_lowercase
  has_number   = each.value.features.password_complexity_policy.has_number
  has_symbol   = each.value.features.password_complexity_policy.has_symbol
}

# Password Age Policy (Production)
resource "zitadel_password_age_policy" "org_password_age_policies" {
  for_each = local.saas_organizations
  
  org_id      = zitadel_org.saas_orgs[each.key].id
  max_age_days = 90   # Password expires after 90 days
  expire_warn_days = 7  # Warn 7 days before expiry
}

# Privacy Policies (Production URLs)
resource "zitadel_privacy_policy" "org_privacy_policies" {
  for_each = local.saas_organizations
  
  org_id        = zitadel_org.saas_orgs[each.key].id
  tos_link      = "https://${each.value.domain}/legal/terms"
  privacy_link  = "https://${each.value.domain}/legal/privacy"
  help_link     = "https://${each.value.domain}/help"
  support_email = each.value.admin_email
  docs_link     = "https://${each.value.domain}/docs"
}

# Lockout Policies (Strict for production)
resource "zitadel_lockout_policy" "org_lockout_policies" {
  for_each = local.saas_organizations
  
  org_id              = zitadel_org.saas_orgs[each.key].id
  max_password_attempts = 3   # Stricter for production
  max_otp_attempts    = 3
  show_lockout_failure = false  # Don't reveal lockout details
}

# Domain Policies (Production settings)
resource "zitadel_domain_policy" "org_domain_policies" {
  for_each = local.saas_organizations
  
  org_id                    = zitadel_org.saas_orgs[each.key].id
  user_login_must_be_domain = true   # Enforce domain-based logins
  validate_org_domains      = true   # Validate organization domains
  smtp_sender_address_matches_instance_domain = true
}

# Label Policy (Branding for production)
resource "zitadel_label_policy" "org_label_policies" {
  for_each = local.saas_organizations
  
  org_id              = zitadel_org.saas_orgs[each.key].id
  primary_color       = "#1976d2"  # Customize per SaaS
  background_color    = "#fafafa"
  warn_color          = "#f57c00"
  font_color          = "#000000"
  primary_color_dark  = "#1565c0"
  background_color_dark = "#303030"
  warn_color_dark     = "#ff9800"
  font_color_dark     = "#ffffff"
  
  hide_login_name_suffix = false
  error_msg_popup       = true
  disable_watermark     = true  # Remove Zitadel branding
}

# Notification Policy (Production email settings)
resource "zitadel_notification_policy" "org_notification_policies" {
  for_each = local.saas_organizations
  
  org_id              = zitadel_org.saas_orgs[each.key].id
  password_change     = true
  
  # Email notifications will be configured via SMTP settings
}

# Custom Text for each organization (Localization)
resource "zitadel_custom_text" "org_custom_texts" {
  for_each = local.saas_organizations
  
  org_id   = zitadel_org.saas_orgs[each.key].id
  language = "en"
  
  login_texts = {
    "login.title" = "Sign in to ${each.value.name}"
    "login.description" = "Welcome to ${each.value.name}. Please sign in to continue."
    "login.loginname.label" = "Email or Username"
    "login.password.label" = "Password"
    "login.button.login" = "Sign In"
  }
  
  password_texts = {
    "password.title" = "Set Password"
    "password.description" = "Set a secure password for your ${each.value.name} account."
  }
  
  mfa_texts = {
    "mfa.title" = "Two-Factor Authentication"
    "mfa.description" = "Please complete two-factor authentication to access ${each.value.name}."
  }
}