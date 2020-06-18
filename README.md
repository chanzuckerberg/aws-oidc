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
Executes a command with AWS credentials loaded in the environment. Requires your `~/.aws/config` to be managed through `aws-oidc configure`.
``` bash
$ aws-oidc exec --profile <your profile> -- aws sts get-caller-identity
{
	“UserId”: <...>
	“Account”: <Account from that role-arn flag>
	“Arn:”: <AWS STS ARN for the role-arn flag>
}
```
### serve-config
Deploys a service that displays an AWS Config file for any authorized visitor (see [Deployment Requirements](#deployment-requirements))

### configure
Will query your aws config service (serve-config command) to help populate your `~/.aws/config`. It will guide you through the process of setting this up.

### env
Env is primarily here to assist when running docker locally. It requires your `~/.aws/config` to be configured through `aws-oidc configure`. You can run the following to test it out:

```
docker run -it --env-file <(go run main.go env --profile <your aws profile>) amazon/aws-cli sts get-caller-identity
```

### version
Prints the version of aws-oidc to stdout.


# More docs
See [docs](docs) for more docs.

# Contributing
We use standard go tools + makefiles to build aws-oidc. Getting started should be as simple as-

1. install go
1. Clone this repo from `git@github.com:chanzuckerberg/aws-oidc.git`
1. `make setup && make`

We follow the [Contributor Conduct](https://www.contributor-covenant.org/version/2/0/code_of_conduct/).

# Copyright
Copyright 2019-2020, Chan Zuckerberg Initiative, LLC

For our license, see [LICENSE](LICENSE).
