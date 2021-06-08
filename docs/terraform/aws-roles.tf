module "readonly" {
  source            = "github.com/chanzuckerberg/cztack//aws-iam-role-readonly?ref=v0.35.0"
  oidc = [
    {
      idp_arn : aws_iam_openid_connect_provider.idp.arn,
      client_ids : [
        module.aws-oidc["readonly"].client_id,
        // We can also authorize admin for the readonly
        module.aws-oidc["admin"].client_id,
      ],
      provider : "${local.okta_tenant}.okta.com",
    }
  ]
}

module "poweruser" {
  source = "github.com/chanzuckerberg/cztack//aws-iam-role-poweruser?ref=v0.35.0"
  oidc = [
    {
      idp_arn : aws_iam_openid_connect_provider.idp.arn,
      client_ids : [module.aws-oidc["admin"].client_id],
      provider : "${local.okta_tenant}.okta.com",
    }
  ]
}
