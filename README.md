**Please note**: If you believe you have found a security issue, _please responsibly disclose_ by contacting us at [security@chanzuckerberg.com](mailto:security@chanzuckerberg.com).

----

# Introduction
AWS-OIDC is a command-line utility tool for generating temporary AWS STS credentials from an OIDC application. This works by:
- opening a browser window with the Identity Provider URL. this helps offboard the heavy logic around authentication + MFA to browser
- doing a local redirection to a temporary server on localhost to return the credentials back to our process
- Verifying flow with PKCE/public client
- Redeeming an id_token with the appropriate scopes
- Exchanging that token for temporary STS credentials

We also included a config generation web service that displays an AWS-OIDC-based Configuration file for authorized clients. The authorization requires an Okta Identity Provider, an AWS organizations role, and AWS worker roles for the accounts needed in the Config file.

# Install (Linux, macOS)
We recommend using [homebrew](https://brew.sh/):
```
brew tap chanzuckerberg/tap
brew install aws-oidc
```

## WSL2
We have tested on WSL2 Ubuntu-18. Make sure you've [upgraded](https://docs.microsoft.com/en-us/windows/wsl/install-win10#step-5---set-wsl-2-as-your-default-version) to WSL2. A couple extra steps are required:
```
sudo apt update && sudo apt install xdg-utils
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
Sets up the webserver that clients ping to set up their AWS Config. 

### configure
Will query the aws config service (serve-config command) to help populate your `~/.aws/config`. It will guide you through the process of setting this up.

### env
Env is primarily here to assist when running docker locally. It requires your `~/.aws/config` to be configured through `aws-oidc configure`. You can run the following to test it out:

```
docker run -it --env-file <(aws-oidc env --profile <your aws profile>) amazon/aws-cli sts get-caller-identity
```

### token
Prints the OIDC tokens to stdout in JSON format. Useful for debugging or piping into other tools.
``` bash
$ aws-oidc token --issuer-url=<issuer url> --client-id=<client ID>
{"version":1,"id_token":"...","access_token":"...","expiry":"2026-02-25T15:38:01Z"}
```

### check-token-valid
Checks whether the cached OIDC token is present and valid. Prints `valid` or `invalid` to stdout and exits with a non-zero status if the token is missing or expired. This does **not** trigger a refresh flow.
``` bash
$ aws-oidc check-token-valid --issuer-url=<issuer url> --client-id=<client ID>
valid
```

### check-refresh-ttl
Prints the remaining time-to-live of the cached refresh token by calling the issuer's introspection endpoint. Useful for monitoring when a user will need to re-authenticate.
``` bash
$ aws-oidc check-refresh-ttl --issuer-url=<issuer url> --client-id=<client ID>
2159h59m57.088655s
```

### version
Prints the version of aws-oidc to stdout.

## Distributed / NFS Environments

In environments where many hosts share a home directory over NFS (e.g. HPC clusters, shared compute nodes), the default OIDC token cache at `~/.cache/oidc-cli/` can cause problems:

- `O_APPEND` writes are not atomic on NFS, so concurrent processes on different hosts can overwrite each other's data.
- `flock`-based file locking is unreliable across NFS clients.
- Atomic `rename` may not be safe across NFS server frontends.

Use the `--node-local-cache` flag (or the `AWS_OIDC_NODE_LOCAL_CACHE` environment variable) to store the OIDC token cache and lock files on node-local disk instead. The value should be a directory that is **not** on an NFS mount and **unique per user** (e.g. `/tmp/oidc-cache-$(id -u)` or a path on a local SSD):

``` bash
$ aws-oidc creds-process --node-local-cache "/tmp/oidc-cache-$(id -u)" \
    --issuer-url=<issuer url> --client-id=<client ID> --aws-role-arn=<role ARN>
```

Or equivalently, via the environment variable:

``` bash
$ export AWS_OIDC_NODE_LOCAL_CACHE="/tmp/oidc-cache-$(id -u)"
$ aws-oidc creds-process --issuer-url=<issuer url> --client-id=<client ID> --aws-role-arn=<role ARN>
```

``` bash
$ aws-oidc exec --node-local-cache "/tmp/oidc-cache-$(id -u)" --profile <your profile> -- aws sts get-caller-identity
```

The environment variable is useful for setting this once in your shell profile (e.g. `.bashrc`, `.zshrc`) so every invocation uses node-local storage automatically. If both the flag and the environment variable are set, the flag takes precedence.

On first use, the local cache is bootstrapped by copying the existing token from the default (NFS) cache. All subsequent reads, writes, and lock operations use the local directory, avoiding cross-host contention entirely.

This flag is available on all subcommands (`token`, `exec`, `creds-process`, `configure`, `check-token-valid`, `check-refresh-ttl`).

# More docs
See [docs](docs) for more docs.

# Contributing
We use standard go tools + makefiles to build aws-oidc. Getting started should be as simple as-

1. install go
1. Clone this repo from `git@github.com:chanzuckerberg/aws-oidc.git`
1. `make setup && make`

We follow the [Contributor Conduct](https://www.contributor-covenant.org/version/2/0/code_of_conduct/).

## Releases
Each time a change gets merged to main, GitHub Actions triggers a release process. The build gets pushed to our [chanzuckerberg/homebrew-tap](https://github.com/chanzuckerberg/homebrew-tap) repo so you can pull the latest version using `brew`.

# Copyright
Copyright 2019-2021, Chan Zuckerberg Initiative, LLC

For our license, see [LICENSE](LICENSE).

## Code of Conduct

This project adheres to the Contributor Covenant [code of conduct](https://github.com/chanzuckerberg/.github/blob/master/CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code. 
Please report unacceptable behavior to [opensource@chanzuckerberg.com](mailto:opensource@chanzuckerberg.com).

