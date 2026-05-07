# Using `aws-oidc serve-imds` as a workload credential helper

`aws-oidc serve-imds` runs as a long-lived process that brokers Okta OAuth 2.0 `client_credentials` into AWS STS credentials and serves them on a localhost endpoint that mimics EC2 IMDSv2. Unmodified workloads (Prometheus, AWS CLI, anything using the AWS SDK) pick up the credentials via the standard credential chain — they don't need to know `aws-oidc` exists.

This document covers the operator-side setup (Okta + AWS), the producer-side runtime patterns (systemd, Kubernetes sidecar), security expectations, and the most common failure modes.

## Audience

Operators rolling this out for a non-EKS workload that needs to call AWS (e.g. Prometheus remote-writing to Amazon Managed Prometheus from a non-AWS host). For interactive human authentication, use `aws-oidc creds-process` / `aws-oidc exec` instead — `serve-imds` is the unattended-workload analogue.

## How it fits together

```
┌─── workload host (HPC, non-EKS K8s, anything outside AWS) ───────┐
│                                                                  │
│   client_secret on disk                                          │
│           │                                                      │
│           ▼                                                      │
│   ┌──────────────────────────┐    Okta CUSTOM Auth Server        │
│   │ aws-oidc serve-imds      │ ─► POST /oauth2/<id>/v1/token     │
│   │   (Okta client_creds  →  │ ◄─ access_token JWT               │
│   │    STS AssumeRoleWith-   │                                   │
│   │    WebIdentity  →        │ ─► AWS STS                        │
│   │    cached  →             │ ◄─ short-lived AWS credentials    │
│   │    serves IMDSv2         │                                   │
│   │    on 127.0.0.1:9911)    │                                   │
│   └─────────────┬────────────┘                                   │
│                 │ AWS SDK credential chain                       │
│                 │   AWS_EC2_METADATA_SERVICE_ENDPOINT=...        │
│                 ▼                                                │
│   ┌─────────────────────────┐                                    │
│   │ Prometheus / AWS CLI /  │                                    │
│   │ any AWS-SDK workload    │                                    │
│   │   sigv4 { region: ... } │                                    │
│   └─────────────┬───────────┘                                    │
└─────────────────┼────────────────────────────────────────────────┘
                  │ SigV4-signed AWS API call
                  ▼
                AWS service (e.g. AMP remote_write)
```

## Prerequisites — Okta side

You need:

1. **An Okta API Service application** (Service App, OAuth 2.0 Client Credentials grant) per workload identity:
   - Note the `client_id` and `client_secret`.
   - Client authentication method: `client_secret_post` (default).
2. **A Custom Authorization Server** (NOT the Org / default Okta authorization server):
   - The Org AS endpoint at `/oauth2/v1/token` issues **opaque** access tokens whose structure is "subject to change at any time without notice" per Okta documentation. AWS STS cannot validate them.
   - The Custom AS endpoint is `/oauth2/<auth_server_id>/v1/token`. Tokens issued here are RS256-signed JWTs with the audience you configure.
   - Configure the Custom AS's single `audiences` field to a stable URL like `https://aws-sts.<your-app>`.
3. **A custom scope on the Custom AS** (e.g. `aws-m2m-access`) with `consent=IMPLICIT`, granted to the Service app.
4. **An access policy rule** on the Custom AS allowing the Service app to use the `client_credentials` grant for the custom scope.

Reference walkthrough: [okta-aws-cli M2M Command Requirements](https://github.com/okta/okta-aws-cli#m2m-command-requirements). Their setup uses `private_key_jwt` rather than `client_secret_post` but the AWS-side configuration is identical.

## Prerequisites — AWS side

You need three IAM resources in the account that holds the target role:

1. **An IAM OIDC identity provider**:
   - URL: the Custom AS URL with no trailing slash and exact casing — e.g. `https://<org>.okta.com/oauth2/aus123abc`.
   - Client ID list: contains your Custom AS audience value, e.g. `["https://aws-sts.<your-app>"]`.
2. **A target IAM role** with a trust policy allowing `sts:AssumeRoleWithWebIdentity` from the OIDC provider, conditioned on the audience and (preferably) the Service app's client_id:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Principal": {
         "Federated": "arn:aws:iam::ACCT:oidc-provider/<org>.okta.com/oauth2/aus123abc"
       },
       "Action": "sts:AssumeRoleWithWebIdentity",
       "Condition": {
         "StringEquals": {
           "<org>.okta.com/oauth2/aus123abc:aud": "https://aws-sts.<your-app>",
           "<org>.okta.com/oauth2/aus123abc:sub": "0oaSERVICEAPPCLIENTID"
         }
       }
     }]
   }
   ```
   Use `:aud` and `:sub`. Do **not** use `:azp` — Okta's default token profile does not emit an `azp` claim (it uses `cid` for the client ID instead, and the IAM trust-policy condition keys cover `aud`/`sub`/`amr`/`auth_time`/`iat`/`iss`).
3. **The role's permission policy** should grant only what the workload needs. For example, AMP remote-write would be `aps:RemoteWrite` on a specific workspace ARN.

### The audience-three-way coupling

The audience value MUST be **identical** in three places:

- The Custom AS's `audiences` field (Okta stamps this into the JWT's `aud` claim).
- The IAM OIDC provider's `client_id_list` (AWS validates `aud` against this list).
- The role trust policy's `Condition` on `<issuer>:aud`.

A single mismatch produces an `InvalidIdentityToken` / "Token does not match expected audience" error from STS, which is hard to debug downstream. **Decide on the audience value first, before configuring anything.**

### R1 verification recipe

Before deploying the helper to a real workload, verify the round-trip works in your tenant. From a shell with AWS credentials that can call STS:

```bash
# 1. Get an access token from your Custom AS.
TOKEN=$(curl -s -X POST \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "scope=$SCOPE" \
  "https://$OKTA_DOMAIN/oauth2/$AUTH_SERVER_ID/v1/token" \
  | jq -r .access_token)

# 2. Decode the JWT and confirm `aud` and `sub`.
echo "$TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq

# 3. Exchange it via STS.
aws sts assume-role-with-web-identity \
  --role-arn "$ROLE_ARN" \
  --role-session-name "$CLIENT_ID" \
  --web-identity-token "$TOKEN"
```

If step 3 returns 200, the Okta + AWS configuration is correct and `serve-imds` will work. If it returns `InvalidIdentityToken`, check the audience-three-way-coupling above before debugging anything else.

## Producer-side: Linux + systemd

For an HPC head node or any standalone Linux host:

```ini
# /etc/systemd/system/aws-oidc-serve-imds.service
[Unit]
Description=aws-oidc serve-imds (workload OIDC -> AWS STS -> IMDS)
Wants=network-online.target chronyd.service
After=network-online.target chronyd.service

[Service]
Type=simple
DynamicUser=yes
SupplementaryGroups=aws-oidc-secret-readers
ExecStart=/usr/local/bin/aws-oidc serve-imds
EnvironmentFile=/etc/aws-oidc/imds.env
Restart=on-failure
RestartSec=5s

ProtectSystem=strict
ProtectHome=yes
NoNewPrivileges=yes
ReadOnlyPaths=/etc/aws-oidc

[Install]
WantedBy=multi-user.target
```

```bash
# /etc/aws-oidc/imds.env  (mode 0644, contains no secret)
AWS_OIDC_IMDS_CLIENT_ID=0oa...
AWS_OIDC_IMDS_CLIENT_SECRET_FILE=/etc/aws-oidc/client_secret
AWS_OIDC_IMDS_ISSUER_URL=https://example.okta.com/oauth2/aus123abc
AWS_OIDC_IMDS_AWS_ROLE_ARN=arn:aws:iam::123456789012:role/argus-amp-producer-bruno
AWS_OIDC_IMDS_SCOPE=aws-m2m-access
AWS_OIDC_IMDS_AWS_REGION=us-west-2
```

The `client_secret` file should be `0600`, owned by `root:aws-oidc-secret-readers`. The helper runs under `DynamicUser=yes` and is added to that group via `SupplementaryGroups`, so it can read the secret without root.

For Prometheus on the same host, edit its systemd unit:

```ini
[Service]
Environment=AWS_EC2_METADATA_SERVICE_ENDPOINT=http://127.0.0.1:9911
Environment=AWS_REGION=us-west-2

[Unit]
After=aws-oidc-serve-imds.service
Requires=aws-oidc-serve-imds.service
```

Drop the `access_key` / `secret_key` fields from the `sigv4` block in `prometheus.yml`. Leave `region`. The AWS SDK chain inside Prometheus will discover credentials via the local IMDS endpoint.

## Producer-side: Kubernetes sidecar (non-EKS)

For a Prometheus pod in a non-EKS cluster (e.g. CoreWeave):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-oidc-imds
  namespace: monitoring
type: Opaque
stringData:
  client_secret: "<okta-client-secret>"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-oidc-imds
  namespace: monitoring
data:
  AWS_OIDC_IMDS_CLIENT_ID: "0oa..."
  AWS_OIDC_IMDS_CLIENT_SECRET_FILE: "/etc/aws-oidc/client_secret"
  AWS_OIDC_IMDS_ISSUER_URL: "https://example.okta.com/oauth2/aus123abc"
  AWS_OIDC_IMDS_AWS_ROLE_ARN: "arn:aws:iam::123456789012:role/argus-amp-producer-cw"
  AWS_OIDC_IMDS_SCOPE: "aws-m2m-access"
  AWS_OIDC_IMDS_AWS_REGION: "us-west-2"
---
# Sidecar container in the Prometheus pod
spec:
  containers:
    - name: prometheus
      image: prom/prometheus:v2.x
      env:
        - name: AWS_EC2_METADATA_SERVICE_ENDPOINT
          value: http://127.0.0.1:9911
        - name: AWS_REGION
          value: us-west-2
      # ... your existing prometheus config

    - name: aws-oidc-serve-imds
      image: <your-registry>/aws-oidc:latest
      args: ["serve-imds"]
      envFrom:
        - configMapRef:
            name: aws-oidc-imds
      volumeMounts:
        - name: client-secret
          mountPath: /etc/aws-oidc
          readOnly: true
      readinessProbe:
        httpGet:
          path: /readyz
          port: 9911
      livenessProbe:
        httpGet:
          path: /healthz
          port: 9911

  volumes:
    - name: client-secret
      secret:
        secretName: aws-oidc-imds
        items:
          - key: client_secret
            path: client_secret
            mode: 0o600
```

The two containers share the pod's network namespace, so Prometheus reaches the helper at `127.0.0.1:9911` without crossing any pod or host boundary. No other pod can reach the helper.

## Security notes

- **`client_secret` hygiene.** The file MUST be mode `0600`. The helper warns at startup if it's looser; do not ignore that warning. The secret is never written to logs (`workload_oidc.Config.LogValue` redacts it at all slog levels).
- **Bind address.** Default is `127.0.0.1`. The helper warns at startup if you set it to a non-loopback address. K8s sidecar deployments are safe because the pod's network namespace contains only the helper and the workload.
- **TLS.** Not used on the local IMDS endpoint, matching real EC2 IMDS behavior. The trust boundary is "anything that can reach 127.0.0.1 on this host can fetch credentials" — same as real EC2.
- **IMDSv2 token TTL.** Default 6 hours, the AWS-documented maximum. Use `--require-imdsv2` to refuse v1-style requests for stricter posture.
- **Secret rotation requires restart.** The helper reads `client_secret` once at startup. If you rotate the Okta client_secret, restart the helper. Periodic re-read is not currently supported.
- **CloudTrail attribution.** The role session name passed to `AssumeRoleWithWebIdentity` is the OAuth `client_id`, which is also what Okta puts in the JWT's `sub` claim — so trust-policy conditions, the JWT, and the role session name all line up to the same identifier. CloudTrail's `userIdentity.userName` for downstream API calls shows `<role-name>/<client_id>`.

## Operational notes

- **Startup validation.** On startup, the helper performs one full Okta → STS round trip before binding the listener. Misconfiguration (bad client_secret, wrong audience, role trust policy mismatch, network) produces an immediate, clear error rather than a silent failure at first request.
- **Refresh.** The `aws.CredentialsCache` underneath refreshes credentials 5 minutes before expiry by default (`--refresh-before`). Concurrent IMDS requests during a refresh are coalesced into one Okta + STS round trip.
- **Graceful shutdown.** SIGTERM / SIGINT triggers `http.Server.Shutdown` with a 5-second drain window, after which the process exits cleanly.
- **Multiple identities on one host.** The helper serves one role per process. To serve multiple roles, run multiple processes on different `--port`s and point each consumer's `AWS_EC2_METADATA_SERVICE_ENDPOINT` accordingly.

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| `InvalidIdentityToken` / "Token does not match expected audience" | The Custom AS audience config doesn't match the IAM OIDC provider `client_id_list` or the trust policy's `aud` condition. Decode the JWT (`echo $TOKEN \| cut -d. -f2 \| base64 -d`) and compare. |
| `IDPCommunicationError` from STS | AWS can't fetch the OIDC discovery document. Check that the IAM OIDC provider URL (no trailing slash, exact casing) resolves and the JWKS endpoint is reachable from AWS. Custom Okta domains need a publicly trusted TLS chain. |
| `InvalidIdentityToken: Couldn't retrieve verification key` | The JWT is signed with a key that's not in the JWKS, or the issuer URL casing doesn't match. Or you're hitting the Org AS instead of a Custom AS — the helper warns about this at startup but check anyway. |
| `InvalidParameterException: 1 validation error detected: Value '...' at 'roleSessionName' failed to satisfy constraint` | The OAuth `client_id` contains characters STS rejects in `RoleSessionName`. Open an issue; the helper sanitizes session names but unusual `client_id`s can still land here. |
| Prometheus shows `failed to retrieve credentials` from IMDS | The helper isn't running, isn't bound to the right address/port, or `AWS_EC2_METADATA_SERVICE_ENDPOINT` isn't set in the Prometheus process. Check `journalctl -u aws-oidc-serve-imds` (Linux) or `kubectl logs ... -c aws-oidc-serve-imds` (K8s). |
| Helper logs `transient error, retrying` | Okta /token returned 5xx. Self-heals after up to 3 retries with exponential backoff; persistent failure produces `failed startup credential fetch` and the process exits. |

## What's NOT included

- This subcommand does not provision the Okta Service app, the AWS IAM OIDC provider, or the IAM role. Those are operator tasks (Terraform / Okta admin / AWS console).
- It does not rotate the `client_secret`. Restart the helper after rotation.
- It does not refresh the cert chain for self-hosted issuer URLs (none used by default).
