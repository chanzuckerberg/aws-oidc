locals {
  // This is a map where keys are the names of the oidc apps
  //      and values are the set of authorized okta groups.
  aws_oidc_apps = {
    "readonly" : [okta_group.readonly],
    "admin": [okta_group.admin],
  }
}

module "aws-oidc" {
  source   = "./okta-aws-oidc-app"
  for_each = local.aws_oidc_apps

  app_name          = each.key
  authorized_groups = [for group in each.value : group.id]
}


resource "okta_group" "readonly" {
  name        = "readonly"
  description = "This group authorizes users for readonly access"

  // In this example we hard-code users for brevity.
  // In practice, we've learned that you're better off using okta rules instead.
  // see https://registry.terraform.io/providers/oktadeveloper/okta/latest/docs/resources/group_rule
  users = ["foo", "bar"]
}

resource "okta_group" "admin" {
  name        = "admin"
  description = "This group authorizes for admin access"

  users = ["foo", "baz"]
}
