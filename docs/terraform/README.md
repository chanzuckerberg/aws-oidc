This dir has some basic (but hopefully representative) terraform configuration for how to wire aws-oidc up.

We use Okta but you should be able to achieve a similar configuration with another OIDC compliant idp.



# AWS Config Generation
We also have a service that helps users generate their `~/.aws/config` and discover which roles are available to them across all your AWS accounts. Unlike the rest of the project, this service currently relies on Okta as the IDP. It requires Okta Authentication.

This server will (on a schedule) scan your Okta instance and AWS accounts to see which users are authorized for which roles.

## Okta configuration Service
There are a couple of pieces that need to be configured for the AWS Config Generation service to work.
### Okta
We create an Okta oauth app to authenticate to the Okta API with least privileges. We use an RSA key as the authentication mechanism. The Config Generation service needs access to this RSA key. Steps to configure the service:

- Generate an RSA key. We added a convenience `rsa-keygen` aws-oidc command that will generate a key for you and write it to the current working directory as `rsa`. It will also echo back the public key in a format that Okta can consume.

```
➜  aws-oidc rsa-keygen | jq .
{
  "use": "sig",
  "kty": "RSA",
  "kid": "SHA256:Ty5t0WBhdzzvJRwoCcc1tow3UZ08CDS2UwEGCE6iuiY",
  "alg": "RS256",
  "n": "sy_gjdJHT1G1IiuY_J7KVm9NM1peSD1oNJzBmiJb3g4ueYr4fcRbRFG8ZDRWJ1_EpctL_NDooBKOgV9TMLKkx2no1x0yp3lJA1potslhamedXo5lWQll5VSb5o5TIKw3SXILyXs4jsGTsciRt7zFqgADtWEWle_Hj-kUiM2KfIhILc5OtQwIZ_CYKYkM_0NHfhcKD28-gN-94G2MBvQO7N2sNopOsFOzl7pWsWcFH9eDV9yUHLkZXhmZizyffkYcedTxwp45EM7kE4-zdt-OpFH5rfWzr22XlmA0mf2m93MoRugkGyp-bJ1IbdUV3hC3Q2w8QGM3i8114RNjKHYSmXRK9uoX0SUP4tMsRNu2BdSZHYLVuuaExs0SV5pAEcphRWcGXMyk-af8sTXJsRVu7KCDBwj0DQ35VE2Z1Igo_RanOEpTYp27n2cP3AcfMIGEZLgeodRrgJZpgRyKoQ_QeE0N3oIfpqr5N58f6lG5vSaQt_-BfskQ3NyEj0Sbd_HJhWqamNtbBzDUZWVMTbB_1eMmqP5MV8ZQpPXUFCNp7l22EM1FWICwPrD1Fu43CAxcASGt1C5Yt08eo3MyQOjRLitS3e3GqQsNlhKlhjrlyclsQ6Aeex6CcRvKvZaYH6dtRng73kicl0h43HMupvUoKu5PNYvcQYTC3_-q_kgbY4k",
  "e": "AQAB"
}
```

- Config Generation Service will need the private key as the `OKTA_PRIVATE_KEY` environment variable.

- Create an Okta Oauth app and use the public key for authentication. A Terraform snippet to achieve this (using the above key) looks like:
```
// This allows our aws config generation service to interact with the Okta API
// The private key is stored in the param store
resource "okta_app_oauth" "aws-config-generation-okta-api" {
  label         = "aws-config-generation okta api"
  type          = "service"
  grant_types   = ["client_credentials"]
  redirect_uris = ["<where you host the service>/oauth2/callback"]
  login_uri     = "<where you host the service>"
  omit_secret   = true
  response_types = [
    "token",
  ]
  token_endpoint_auth_method = "private_key_jwt"

  jwks {
    kid = "SHA256:Ty5t0WBhdzzvJRwoCcc1tow3UZ08CDS2UwEGCE6iuiY"
    kty = "RSA"
    e   = "AQAB"
    n   = "sy_gjdJHT1G1IiuY_J7KVm9NM1peSD1oNJzBmiJb3g4ueYr4fcRbRFG8ZDRWJ1_EpctL_NDooBKOgV9TMLKkx2no1x0yp3lJA1potslhamedXo5lWQll5VSb5o5TIKw3SXILyXs4jsGTsciRt7zFqgADtWEWle_Hj-kUiM2KfIhILc5OtQwIZ_CYKYkM_0NHfhcKD28-gN-94G2MBvQO7N2sNopOsFOzl7pWsWcFH9eDV9yUHLkZXhmZizyffkYcedTxwp45EM7kE4-zdt-OpFH5rfWzr22XlmA0mf2m93MoRugkGyp-bJ1IbdUV3hC3Q2w8QGM3i8114RNjKHYSmXRK9uoX0SUP4tMsRNu2BdSZHYLVuuaExs0SV5pAEcphRWcGXMyk-af8sTXJsRVu7KCDBwj0DQ35VE2Z1Igo_RanOEpTYp27n2cP3AcfMIGEZLgeodRrgJZpgRyKoQ_QeE0N3oIfpqr5N58f6lG5vSaQt_-BfskQ3NyEj0Sbd_HJhWqamNtbBzDUZWVMTbB_1eMmqP5MV8ZQpPXUFCNp7l22EM1FWICwPrD1Fu43CAxcASGt1C5Yt08eo3MyQOjRLitS3e3GqQsNlhKlhjrlyclsQ6Aeex6CcRvKvZaYH6dtRng73kicl0h43HMupvUoKu5PNYvcQYTC3_-q_kgbY4k"
  }
}

```
- Grant the app the following Scopes (In Okta under the "Okta API Scopes" tab) `okta.apps.read` and `okta.users.read`. These are the 2 Okta API endpoints we use to match Okta users to AWS roles. Note these are read-only endpoints.

- Config Generation Service will need this App's `client_id` as the `OKTA_SERVICE_CLIENT_ID` environment variable.

- No users should be assigned to this application.

We also create a separate OIDC app for users to authenticate with the config generation service:

- Create an Okta OIDC App with the following configuration:
```
resource "okta_app_oauth" "aws-config-generation" {
  label         = "<your Okta naming convention, otherwise the URL where this service is hosted>"
  type          = "native"
  grant_types   = ["authorization_code", "refresh_token"]
  // port range by convention, to work around port race conditions.
  redirect_uris = formatlist("http://localhost:%d", range(49152, 49512 + 64))
  login_uri     = "http://localhost:49512"
  omit_secret   = true
  response_types = [
    "code",
  ]
  // If you want you can set a more "memorable" client_id, but not necessary.
  client_id = "aws-config"

  // Use PKCE for public clients
  token_endpoint_auth_method = "none"
}
```

- Config Generation Service will need this Application's `client_id` as the `OKTA_CLIENT_ID`.

- Everyone that needs access to AWS should be assigned to this Application.

- Your Okta Domain (Issuer URL) needs to be exposed as the `OKTA_ISSUER_URL` to the service.

### AWS
In addition to scanning Okta users and applications, we need to be able to list roles in AWS. We assume a role in the Organization Main account that lists all accounts in that organization. We then assume a role in each account that lists all roles in the account. We then parse the assume role policy for each role and match them with Okta data.

Note that each of these Roles must be Assumable from the AWS Config Generation Service's role.

- In each Main account in your AWS Organization, create a role with the following policy:
```hcl
data "aws_iam_policy_document" "aws-config-generation-organizations" {
  statement {
    sid = "ListAccounts"

    actions = ["organizations:ListAccounts"]

    resources = ["*"]
  }
}
```
- If you have multiple Organizations you can create this role in each. Note the ARNs for these roles. The Config Generation Service will look for a list in the `AWS_ORG_ROLE_ARNS` environment variables.

- You will also need to create a Role in each account. We assume all of these roles share the same name and will interpolate the `account_ids` returned from the previous step to generate an `arn` to assume. The name of these roles needs to be exposed as the `AWS_READER_ROLE_NAME` environment variable. These roles need the following policy:
```hcl
data "aws_iam_policy_document" "aws-config-generation-worker" {
  statement {
    sid = "ListRoles"

    actions = ["iam:ListRoles", "iam:ListRoleTags"]

    resources = ["*"]
  }
  statement {
    sid = "ListAliases"

    actions = ["iam:ListAccountAliases"]

    resources = ["*"]
  }
}
```
- You can now deploy the AWS Config Generation Service in your framework of choice. `aws-oidc serve-config` will launch the service, by default on port `8080`. You will need to set all the environment variables from previous steps.

- As an end user, running `aws-oidc configure --issuer-url <OKTA_ISSUER_URL> --client-id <OKTA_CLIENT_ID> --config-url <where you are hosting the configuration service>` will guide you through assembling your AWS Config in an aws-oidc compatible way.
