resource "okta_app_oauth" "aws_oidc" {
  label         = "aws_oidc ${var.app_name}"
  type          = "native"
  grant_types   = ["authorization_code", "refresh_token"]
  redirect_uris = formatlist("http://localhost:%d", range(49152, 49152 + 64))
  login_uri     = "http://localhost:49152"
  omit_secret   = true
  response_types = [
    "code",
  ]

  // Use PKCE for public clients
  token_endpoint_auth_method = "none"
}

resource okta_app_group_assignments oauth-aws_oidc {
  app_id    = okta_app_oauth.aws_oidc.id
  group_ids = distinct(var.authorized_groups)
}
