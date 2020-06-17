# Deployment Requirements
Deploying the web service requires a few things:
A master role with permission to run [List Accounts](https://docs.aws.amazon.com/cli/latest/reference/organizations/list-accounts.html) in the AWS Organization
A reader role in each account with permission to run [List Roles](https://docs.aws.amazon.com/cli/latest/reference/iam/list-roles.html) in the accounts
An Okta Identity Provider with a private key, client ID, and issuer URL.

This deployment relies on a working identity provider, which will provide the ID Token needed for identifying any clients that try to interact with the server. The aws-oidc docker image includes [chamber](https://github.com/segmentio/chamber/), which we use for loading sensitive environment variables.

Using the latest version of aws-oidc, run `aws-oidc serve-config --web-server-port=8080`

Ping localhost:8080/health to make sure your service is up and running.

## Environment Variables for Deploying
### Okta Identity Provider:
OKTA_PRIVATE_KEY: the private key from the Okta

OKTA_SERVICE_CLIENT_ID: The client ID of the Okta Client that manages Okta apps for your clients

OKTA_CLIENT_ID: the client ID of the Okta Identity Provider that verifies your clients

OKTA_ISSUER_URL: the URL of the identity provider

You can create create those values using [this tutorial](https://developer.okta.com/docs/guides/create-an-api-token/overview/)


###  AWS Config Generation:
AWS_READER_ROLE_NAME: role name that can run AWS List Roles in within each account in your AWS Organization

AWS_MASTER_ROLE_ARNS: a list of role ARNs that can list accounts in your AWS Organizatio

### Skipping roles
You can tag AWS Roles with "aws-oidc/skip-role" if you don't want serve-config to return this role to users.
