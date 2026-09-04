# Tailscale OIDC enrollment for agent workspace pods

Agent workspace pods join the tailnet by exchanging a Kubernetes projected service-account
token for a Tailscale machine key. The token carries a custom audience
(`api.tailscale.com/<client-id>`) that Tailscale's token-exchange endpoint validates
against a registered OIDC trust relationship.

## Current state (manual, needs codifying)

The OIDC trust relationship for the rdev (dev-central) cluster was created by hand in the
[Tailscale admin console](https://login.tailscale.com/admin/settings/oauth) and is not yet
managed by Terraform. The client ID is `TjHt1v2bSH11CNTRL-kz7ofAGpBJ11CNTRL`.

This needs to be codified before it can be reproduced, rotated, or deployed to other
clusters.

## What needs to be added to biohub-ai-infra

Create a new workspace in
`terraform/envs/<env>/tailscale-federated-identity-agent-rdev/` (or equivalent) using the
existing `modules/tailscale-federated-identity` module:

```hcl
module "tailscale-federated-identity" {
  source      = "../../../modules/tailscale-federated-identity"
  description = "Agent workspace pods on dev-central EKS — OIDC token exchange for tag:mantis-shrimp"
  issuer_url  = "https://oidc.eks.us-west-2.amazonaws.com/id/A72DBBC7C19B4B419EEA3497A3A4CC03"
  subject     = "system:serviceaccount:aws-oidc:*"
  scopes      = ["openid"]
  tags        = ["tag:mantis-shrimp"]
}
```

Key inputs:

| Field | Value |
|---|---|
| `issuer_url` | `https://oidc.eks.us-west-2.amazonaws.com/id/A72DBBC7C19B4B419EEA3497A3A4CC03` |
| `subject` | `system:serviceaccount:aws-oidc:*` (all workspace service accounts in the operator namespace) |
| `scopes` | `["openid"]` |
| `tags` | `["tag:mantis-shrimp"]` |

The module outputs `client_id`. Once applied, replace the placeholder value in
`aws-oidc/.infra/rdev/values.yaml`:

```yaml
- name: TAILSCALE_TOKEN_AUDIENCE
  value: "api.tailscale.com/<output.client_id>"
```

## How the enrollment works

1. The operator projects a service-account token with audience
   `api.tailscale.com/<client-id>` into each workspace pod at
   `/var/run/secrets/tailscale.com/token`.
2. `docker/agent/entrypoint.sh` decodes the `aud` claim from the token's JWT payload to
   recover the bare `<client-id>`, then calls:
   ```
   tailscale up --client-id=<client-id> --id-token=<token> \
     --advertise-tags=tag:mantis-shrimp --hostname=agent-<name>-<workspace>
   ```
3. Tailscale validates the token against the registered OIDC trust relationship and issues
   a machine key. The pod appears in the tailnet as `tag:mantis-shrimp`.
4. The Claude `PreToolUse` hook (`/etc/claude-code/ssh-guard.sh`) blocks any SSH invocation
   whose login user does not match `AGENT_SSH_USER` (the owner's email local part).

## ACL and SSH rules

The tailnet policy (in `biohub-ai-infra/terraform/envs/auth-*/tailscale-policy/policy.jsonc`)
must include:

- `tag:mantis-shrimp` in `tagOwners` so the trust relationship can mint nodes with that tag.
- An IP grant allowing `tag:mantis-shrimp` port 22 on login nodes.
- An SSH `accept` rule (not `check` — SSO re-auth is not supported for tag sources) allowing
  `tag:mantis-shrimp` to connect to login nodes as `autogroup:nonroot`.
