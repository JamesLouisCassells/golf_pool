# GitOps Handoff

This repo is now in "ready to hand off into the cluster repo" shape.
The remaining deployment work is mostly in Dan's `gitops` repo and the cluster, not in this app repo.

## What To Copy Into `dgunzy/gitops`

App resources from this repo:

- `deploy/app/namespace.yaml`
- `deploy/app/deployment.yaml`
- `deploy/app/service.yaml`
- `deploy/app/httproute.yaml`
- `deploy/app/api-configmap.yaml`
- `deploy/app/web-configmap.yaml`

Optional Flux image automation examples from this repo:

- `deploy/gitops/image-repositories.example.yaml`
- `deploy/gitops/image-policies.example.yaml`

If Dan wants this app to follow the same structure already used in `dgunzy/gitops/apps/masters-pool/`, the likely target directory is:

- `apps/masters-pool/`

## Values To Replace In GitOps

Before applying these manifests in the real cluster, replace the placeholders below.

### HTTPRoute

In `deploy/app/httproute.yaml`:

- replace `masters-pool.example.com` with the real hostname
- confirm:
  - `parentRefs.name: main-gateway`
  - `parentRefs.namespace: envoy-gateway-system`
  - `parentRefs.sectionName: https`

Those values were chosen to match the shared gateway pattern already present in Dan's repo.

### Images

In `deploy/app/deployment.yaml`:

- API image: `ghcr.io/jameslouiscassells/masters-pool-api`
- Web image: `ghcr.io/jameslouiscassells/masters-pool-web`

The deployment uses Flux image policy markers already:

- `{"$imagepolicy": "flux-system:masters-pool-api"}`
- `{"$imagepolicy": "flux-system:masters-pool-web"}`

If Dan uses Flux image automation, the example policy files in this directory are designed to match those markers.

## Secret Shape

This app expects a Kubernetes secret named:

- `masters-pool-api-secrets`

Use `deploy/app/api-secrets.example.yaml` as the shape reference only.
Do not commit real secret values in this repo.

Required keys:

- `DATABASE_URL`
- `CLERK_SECRET_KEY`
- `CLERK_JWKS_URL`
- `CLERK_ISSUER`
- `CLERK_AUTHORIZED_PARTIES`
- `CLERK_EMAIL_CLAIM`
- `CLERK_NAME_CLAIM`
- `CLERK_ADMIN_CLAIM`
- `CLERK_ADMIN_VALUE`
- `GOLF_API_KEY`

Likely real cluster differences:

- `DATABASE_URL` should point at the in-cluster Postgres service Dan chooses
- `CLERK_AUTHORIZED_PARTIES` should use the real production host
- Clerk issuer/JWKS values should use the real Clerk tenant domain

## ConfigMap Shape

The API config map currently provides:

- `HTTP_ADDR`
- `GOLF_PROVIDER`
- `GOLF_API_BASE_URL`
- `GOLF_API_HOST`

The web config map currently provides:

- nginx config serving the Vue SPA
- proxy rules for `/api/*`
- proxy rule for `/healthz`

## Flux Tagging Convention

The publish workflow in this repo now emits:

- `sha-<shortsha>`
- `main-<github_run_number>`
- `latest` on `main`
- semver tags on `v*` git tags

For Flux automation, use the `main-<number>` tags.
Those tags are monotonic and safe for numerical image policies.

The example `ImagePolicy` resources in this directory match:

- pattern: `^main-(?P<run>[0-9]+)$`
- policy: highest numeric run wins

## Suggested GitOps Steps

1. Copy the app manifests into `dgunzy/gitops/apps/masters-pool/`.
2. Replace the hostname in `httproute.yaml`.
3. Add the real external secret wiring for `masters-pool-api-secrets`.
4. If desired, add the Flux image repository/policy resources in `flux-system`.
5. Point the deployment image fields at the published GHCR images.
6. Commit to `dgunzy/gitops` and let Flux reconcile.
7. Confirm the app becomes reachable through the shared gateway host.

## Verification Checklist

After the gitops change is applied, verify:

- the `HTTPRoute` is accepted by the gateway
- the `masters-pool` deployment becomes ready
- both containers start:
  - `api`
  - `web`
- `/healthz` returns success through the service/pod path
- the public site loads through the real hostname
- `/api/config/:year` works through the deployed route
- Clerk redirect / authorized party settings match the real hostname

## Known Remaining Cluster-Side Risk

This repo cannot prove the final deployment by itself.
The remaining unknowns are cluster-owned:

- exact external secret wiring
- exact database service/credentials
- exact hostname/TLS setup
- whether Dan wants Flux image automation resources in the app path or elsewhere in `gitops`

That is the final integration step left.
