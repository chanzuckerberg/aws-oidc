**Please note**: If you believe you have found a security issue, _please responsibly disclose_ by contacting us at [security@chanzuckerberg.com](mailto:security@chanzuckerberg.com).

----

# Introduction
AWS-OIDC is a command-line utility tool for generating temporary AWS STS credentials from an OIDC application. This works by:
- opening a browser window with the Identity Provider URL. this helps offboard the heavy logic around authentication + MFA to browser
- doing a local redirection to a temporary server on localhost to return the credentials back to our process
- Verifying flow with PKCE/public client
- Redeeming an id_token with the appropriate scopes
- Exchanging that token for temporary STS credentials

We also included a config generation web service that displays an AWS-OIDC-based Configuration file for authorized clients. The authorization requires an Okta Identity Provider, an AWS master role, and AWS worker roles for the accounts needed in the Config file.

# Install
```
brew tap chanzuckerberg/tap
brew install aws-oidc
```

# Command-Line Tools
### creds-process
Authenticates into AWS and prints structured AWS credentials to stdout. The stdout output is based on [AWS Configuration for External Processes](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html).
``` bash
$ aws-oidc creds-process --issuer-url=<issuer url> --client-id=<client ID> --aws-role-arn=<AWS role you want credentials for>
{
  "Version": 1,
  "AccessKeyId": "an AWS access key",
  "SecretAccessKey": "your AWS secret access key",
  "SessionToken": "the AWS session token for temporary credentials",
  "Expiration": "ISO8601 timestamp when the credentials expire"
}
```

### exec
Executes a command with AWS credentials loaded in the environment
``` bash
$ aws-oidc exec --issuer-url=<issuer url> --client-id=<client ID> --aws-role-arn=<AWS role you want credentials for>   -- aws sts get-caller-identity
{
	“UserId”: <...>
	“Account”: <Account from that role-arn flag>
	“Arn:”: <AWS STS ARN for the role-arn flag>
}
```
### serve-config
Deploys a service that displays an AWS Config file for any authorized visitor (see [Deployment Requirements](#deployment-requirements))

### version
Prints the version of aws-oidc to stdout.


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
AWS_READER_ROLE_NAME: role name that can run AWS List Roles in any account in your AWS Organization

AWS_MASTER_ROLE_ARNS: a list of role ARNs that can list accounts in your AWS Organization

# Contributing
We use standard go tools + makefiles to build aws-oidc. Getting started should be as simple as-

1. install go
1. Clone this repo from `git@github.com:chanzuckerberg/aws-oidc.git`
1. `make setup && make`

We follow the [Contributor Conduct](https://www.contributor-covenant.org/version/2/0/code_of_conduct/).

# Copyright
Copyright 2019-2020, Chan Zuckerberg Initiative, LLC

For our license, see [LICENSE](LICENSE).
