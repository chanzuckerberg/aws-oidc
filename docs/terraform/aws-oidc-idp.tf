locals {
  // CHANGEME
  okta_tenant = "foobar"
  okta_url = "https://${local.okta_tenant}.okta.com"
}

// https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc_verify-thumbprint.html
// dynamically fetch the cert fingerprint to make sure we're robust to cert rotation
data "tls_certificate" "okta" {
  url = local.okta_url
}

resource "aws_iam_openid_connect_provider" "idp" {
  url = local.okta_url

  client_id_list = [for app in module.aws-oidc: app.client_id]

  thumbprint_list = [
    data.tls_certificate.okta.certificates.0.sha1_fingerprint,
  ]
}
