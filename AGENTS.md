<!-- markdownlint-disable line-length -->

# AGENTS.md — Instructions for AI Coding Agents

This file is for AI agents (Claude Code, Codex, Cursor, Gemini CLI, Copilot,
Aider, Windsurf, or any other agent a customer brings) helping a self-hosted
customer deploy **FiftyOne Enterprise** from this repository. If you are an
agent reading this repo to help someone deploy, follow it as a runbook: work
through the numbered steps in order, and don't move past a step whose **Gate**
hasn't passed.

If you are a human maintaining this repo, see [Maintaining this file](#maintaining-this-file).

## How to invoke this runbook

Point your agent's working directory at this repo (or the relevant
`docker/`/`helm/` subdirectory), start a fresh session so it actually loads
this file, and give it a prompt like one of these — they tell it which entry
point to use.

### Agent-guided deployment

> Using AGENTS.md in this repo as your runbook, help me deploy FiftyOne
> Enterprise from scratch. Start at Step 0 and ask me the intake questions
> before touching any file.

The agent should start at [Step 0](#step-0--determine-deployment-shape) and
work through Steps 0-4 in order, gate by gate.

### Agent-guided upgrade

> Using AGENTS.md in this repo as your runbook, help me upgrade our existing
> FiftyOne Enterprise deployment. We're currently running \<current version\>
> on \<Docker Compose / Helm\> and want to move to \<target version\>.

An upgrade is **not** Step 0 run again from a blank slate — the deployment
shape already exists. Instead the agent should:

1. Confirm the *current* feature set (plugins mode, delegated-operator
   count/type, GPU services, telemetry, etc.) by reading the live
   `.env`/`compose.override.yaml` or `helm get values` — not by re-asking
   Step 0's questions as if nothing were deployed yet — and flag anything
   that looks like drift from what Step 3 recommends.
2. Read [`docker/docs/upgrading.md`](./docker/docs/upgrading.md) or
   [`helm/docs/upgrading.md`](./helm/docs/upgrading.md) for the specific
   version jump and apply every version-specific step it calls out.
3. Set `FIFTYONE_DATABASE_ADMIN=true` only for the duration of the migration
   window, then set it back to `false` (see
   [Non-negotiable rules](#non-negotiable-rules)).
4. For Helm, preview the change first — e.g. the `helm diff` plugin already
   noted in [`helm/README.md`](./helm/README.md#example-with-valuesyaml) —
   rather than applying blind.
5. Re-run [Step 4 — Final validation](#step-4--final-validation) once the
   upgrade completes.

## Non-negotiable rules

Read these before touching any file. They exist because violating them has
caused real customer incidents.

- **This file routes and gates — it does not replace the real docs.**
  [`docker/README.md`](./docker/README.md) and [`helm/README.md`](./helm/README.md)
  (plus their `docs/` subfolders) are the authoritative content for every
  command, env var, and version-specific note. If this file and those ever
  disagree, the subdirectory docs win — open an issue on this repo so the
  drift gets fixed.
- **Never fabricate or guess** at credentials, license file contents, MongoDB
  URIs, DNS records, or TLS certificates. If a value isn't already present in
  the customer's environment or `.env`/`values.yaml`, stop and ask the human
  operator for it.
- **Never duplicate the full `helm/values.yaml`** into a customer's override
  file and hand-edit it in place. Use the minimal-overlay pattern: chart
  defaults (untouched) + a small `my-overrides.yaml` with only the values that
  differ + a separate, untracked secrets file. A full duplicated copy silently
  drifts from the chart on every upgrade and is the single most common cause
  of customer upgrade incidents in this product.
- **Never lose or regenerate `FIFTYONE_ENCRYPTION_KEY`** on an existing
  deployment. It encrypts stored cloud credentials; Voxel51 cannot recover it.
  Confirm it's backed up somewhere durable before doing anything else with an
  existing deployment.
- **Never leave `FIFTYONE_DATABASE_ADMIN=true` set** after a migration
  finishes — it must be temporary, applied only for the duration of an
  upgrade's migration window.
- **Don't run destructive or irreversible commands** — `helm uninstall`,
  `docker compose down -v`, dropping Mongo collections, deleting the license
  or encryption-key secret — without explicit confirmation from the human
  operator, even if a step seems to imply it.
- **If a Gate fails twice in a row on the same step, stop.** Don't keep
  guessing at fixes. Move to [Escalating to Voxel51](#escalating-to-voxel51),
  gather the diagnostic bundle, and let the human operator decide whether to
  retry, work around it, or escalate.
- **Never paste secrets into a GitHub issue** — license contents, connection
  strings, encryption keys, auth secrets. Redact them from any command output
  before it leaves the customer's environment.

## Step 0 — Determine deployment shape

Before editing any file, work through this intake with the human operator and
write down the answers (e.g. as a short checklist in your working notes) —
every later step is conditioned on this profile. This step exists because the
single biggest failure mode of an under-instructed agent deploying this repo
is silently skipping a feature that isn't in the bare-minimum path (this is
exactly what happened when an agent was pointed at this repo with no more
structure than "deploy this": it produced a deployment missing log access,
plugin installs, and delegated operators entirely).

| Question | Options | Where it matters later |
| --- | --- | --- |
| Deployment target? | Docker Compose / Kubernetes (Helm) | Which of Step 2's two paths to follow |
| Identity provider — *and* which protocol? | Which IdP do you use — Okta, Azure AD/Entra ID, Google Workspace, PingIdentity, ADFS, something else — or none at all (air-gapped/self-contained)? Always ask for the **protocol** too (SAML vs. OIDC/OAuth2), since that, not the vendor, picks the auth mode. See "Auth mode, explained" below | `docker/README.md` Step 3 / `helm/fiftyone-teams-app/README.md` |
| Plugins? | Builtin only / Shared / **Dedicated (standard — see below)** | `docker/docs/configuring-plugins.md`, `helm/docs/configuring-plugins.md` |
| Delegated operators (background compute)? | **Some form is standard — see below.** Ask: how many orchestrators do you want, and what type/where should each run? E.g. always-on `teams-do` workers on a VM/on-prem host or in Kubernetes, or on-demand executors on Anyscale, Databricks, or whichever cloud (AWS/GCP/Azure) Kubernetes you already run | `docs/configuring-on-demand-orchestrator.md` + `docs/orchestrators/*` |
| GPU-backed workloads? | Yes/No — needed for delegated operators doing embeddings/inference, and *required* for either builtin service below | `docker/docs/configuring-gpu-workloads.md`, `helm/docs/configuring-gpu-workloads.md` |
| Agentic Labeler (few-shot VLM auto-labeling)? | On/Off — builtin GPU service, needs 24GB+ VRAM (Ampere or newer) | `docker/docs/configuring-agentic-labeler.md`, `docs/configuring-service-orchestrator.md` |
| Annotation AI (SAM2-assisted segmentation)? | On/Off — builtin GPU service, needs 16GB+ VRAM | `docs/configuring-service-orchestrator.md` |
| Multimodal datasets (large per-sample modalities, e.g. sensor streams/point clouds)? | Yes/No | `docker/docs/configuring-multimodal.md`, `helm/docs/configuring-multimodal.md` |
| Telemetry (per-service metrics + log tailing in-app)? | On by default — confirm the customer doesn't want it disabled | `docker/docs/configuring-telemetry.md`, `docker/README.md` §Telemetry |
| HA `teams-api` (multiple replicas)? | Yes/No — Helm only, needs an RWX volume | `helm/docs/configure-ha-teams-api.md` |
| Snapshot archival to cold storage? | Yes/No | `docker/docs/configuring-snapshot-archival.md`, `helm/docs/configuring-snapshot-archival.md` |
| Workload Identity Federation (GCP/AWS)? | Yes/No — Helm only | `helm/docs/configure-workload-identity-federation.md` |
| Custom plugin/DO images (extra Python deps)? | Yes/No | `docs/custom-plugins.md` |
| Custom MongoDB permissions (non-root DB role)? | Yes/No | `docs/custom-mongodb-permissions.md` |
| Corporate proxy in the network path? | Yes/No | `docker/docs/configuring-proxies.md`, `helm/docs/configuring-proxies.md` |
| Air-gapped (no egress to Docker Hub / GHCR / public PyPI)? | Yes/No | Adds items to Step 1's gate — see below |

**Follow-up questions, conditioned on the answers above.** Ask each one only
when its trigger answer was selected; every one of them feeds a Step 1
prerequisite row, so collect them *before* directing anyone to Step 2.

| If Step 0 selected… | Also ask | Feeds |
| --- | --- | --- |
| Shared or Dedicated plugins, or always-on delegated operators | Where does shared plugin storage live, how much capacity, and which pods get read/write vs. read-only access? (Helm: which StorageClass and access mode; Docker: which host path) | Step 1 shared-storage row |
| Any deployment | What is the network path to MongoDB, the IdP, the container registry, and the ingress — including egress restrictions, allowlists, or a proxy in between? | Step 1 network-connectivity row |
| HA `teams-api` | Which RWX-capable StorageClass will back the shared PVC, and is it provisioned in this cluster today? | Step 1 RWX row |
| Snapshot archival | Which cold-storage bucket/path, and what credentials or role grant the deployment write access to it? | Step 1 snapshot-storage row |
| Workload Identity Federation | Which cloud service account is being federated, and does it already carry the IAM permissions the workloads need? | Step 1 Workload Identity row |
| Air-gapped | Where does the Helm chart itself come from — an internal OCI registry or chart mirror, since `helm repo add https://helm.fiftyone.ai` needs egress? | Step 1 air-gapped chart-source row |

**Auth mode, explained:** this isn't a vendor choice — almost any IdP (Okta,
Azure AD/Entra ID, Google Workspace, PingIdentity, OneLogin, ADFS, etc.) can
work with either mode. The actual decision comes down to:

- **Do you need SAML?** Only `legacy-auth` supports it — it's brokered
  through Auth0, which supports SAML plus a wide range of OIDC/OAuth2
  providers. If the customer's identity team says "we're connecting via
  SAML," they need `legacy-auth`.
- **Do you need to avoid depending on Auth0's cloud service** — air-gapped,
  or a policy against external auth dependencies? Use `internal-auth` — CAS
  handles authentication directly, with no external Auth0 hop. It only
  supports OIDC/OAuth2 connections, **not SAML**.
- **Otherwise** (network egress is fine, the IdP will connect via OIDC/OAuth2,
  and SAML isn't required): either mode works technically, but
  `internal-auth` is simpler to operate since it has no external dependency.

If the customer doesn't already know which protocol their IdP will use,
that's a question for their identity/IT team — don't guess at it.

**Standard recommendation:** default the plugins answer to **Dedicated
Plugins**, and plan on setting up **some form of delegated operators** —
unless the customer states a specific reason to run with none at all (e.g. no
long-running or compute-heavy background jobs planned, or a hard resource
constraint on the host/cluster). Both are already the recommended production
configuration according to the docs themselves. But *how many* orchestrators
and *what type/where* each runs (always-on vs. on-demand; on-prem/VM vs.
their existing cloud's Kubernetes vs. Databricks/Anyscale) is entirely the
customer's call — ask, don't assume. Docker Compose's default first-launch
command (see Step 2) happens to include always-on workers; treat that as a
starting point to confirm with the customer, not a decision already made on
their behalf.

## Step 1 — Prerequisites & access gate

Work through the row for your Step 0 answers. **Every required row must be
verified before Step 2 starts** — don't begin writing `.env`/`values.yaml`
until this gate passes.

| Prerequisite | Applies to | Verify with |
| --- | --- | --- |
| Docker + Docker Compose installed | Docker | `docker compose version` |
| `kubectl` + Helm installed, cluster reachable, `kubeVersion >= 1.31` | Helm | `kubectl version`, `helm version`, `kubectl get nodes` |
| License file received from Voxel51 | Both | File exists and is readable; will be mounted per `docker/README.md` Step 2 or `secret-license.template.yaml` for Helm |
| Docker Hub credentials (or GHCR/OCI access for Helm) actually pull | Both | `utils/validate-docker-pulls.sh` (already in this repo) — run it against the images your Step 0 profile needs |
| MongoDB reachable, meets [version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version) (8.0+ recommended) | Both | Connect with the intended `FIFTYONE_DATABASE_URI` / `mongodbConnectionString` using `mongosh` or equivalent before it's ever put in an env file |
| DNS record(s) you control for ingress | Both | You can create/modify the record now, even if it isn't pointed at anything yet |
| TLS/SSL certificate mechanism decided | Both | cert-manager + ClusterIssuer (Helm/GKE example), customer-provided certs, or a load balancer that terminates TLS |
| GPU host(s) available | Only if Step 0 selected GPU workloads, Agentic Labeler, or Annotation AI | `nvidia-smi` on the host, or `kubectl get nodes -o json \| grep nvidia.com/gpu` in-cluster |
| Shared plugin/DO storage exists, is sized, and is writable by the right pods | Shared or Dedicated plugins, or always-on delegated operators | Helm: `kubectl get pvc -n <namespace>` shows the claim `Bound` with the intended capacity; write a file from `teams-plugins` and read it back read-only from `teams-do`. Docker: the host path exists with the intended mode |
| Network path open to MongoDB, IdP, registry, and ingress (proxy/allowlist accounted for) | Both | From the deployment host/pod network: reach the Mongo endpoint, the IdP discovery URL, and the registry — don't assume egress that hasn't been tested |
| RWX-capable StorageClass provisioned for the shared `teams-api` volume | Only if Step 0 selected HA `teams-api` | `kubectl get storageclass`, then bind a test PVC with `ReadWriteMany` and mount it from two pods at once |
| Cold-storage bucket/path reachable and writable for archived snapshots | Only if Step 0 selected snapshot archival | Write and delete a test object at the archive path using the same credentials/role the deployment will use |
| Federated service account carries the IAM permissions the workloads need | Only if Step 0 selected Workload Identity Federation | Impersonate/assume the federated identity and perform one real read *and* write against the buckets the deployment will touch |
| *(air-gapped only)* Internal registry mirrors all required images | Air-gapped | Pull one image through the mirror end-to-end before proceeding |
| *(air-gapped only)* Internal Helm chart source available | Air-gapped + Helm | `helm repo add voxel51 https://helm.fiftyone.ai` needs egress — confirm an internal OCI registry or chart mirror instead, and `helm pull` the target chart version through it |
| *(air-gapped only)* Internal CA distributed; `NODE_EXTRA_CA_CERTS` path known | Air-gapped | `curl` an HTTPS internal endpoint from a test container using that CA |
| *(air-gapped only)* Internal PyPI/package mirror reachable, if plugins/DO images `pip install` anything | Air-gapped | `pip install <pkg> --index-url <mirror>` from a test container |

If any required box can't be checked, stop and ask the human operator to
resolve it (or route it via [Escalating to Voxel51](#escalating-to-voxel51) if
it's a Voxel51-side blocker like a missing license or credentials) before
writing any configuration.

## Step 2 — Guided deployment

Follow the path matching Step 0's deployment-target answer. Each numbered
item's **Gate** must pass before moving to the next.

### Docker Compose path

Mirrors [`docker/README.md`](./docker/README.md) — read that file for full
command/flag detail; this is the ordered checklist with gates.

1. Set up MongoDB (Docker README Step 1) → **Gate:** already covered by Step 1
   of this file.
2. Prepare the license file (Docker README Step 2) → **Gate:** file exists at
   `LOCAL_LICENSE_FILE_DIR/license`, mode `644`.
3. Choose auth mode per Step 0's answer (see
   ["Auth mode, explained"](#step-0--determine-deployment-shape) above if it
   hasn't been nailed down yet), `cd legacy-auth` or `cd internal-auth`
   (Docker README Step 3).
4. Configure `.env` (from `env.template`) and `compose.override.yaml` (Docker
   README Step 4) → **Gate:** `BASE_URL`/`AUTH0_BASE_URL`, `FIFTYONE_API_URI`,
   `FIFTYONE_DATABASE_URI`, `FIFTYONE_ENCRYPTION_KEY`, `LOCAL_LICENSE_FILE_DIR`,
   `FIFTYONE_AUTH_SECRET` are all set to real values, none left as
   placeholders.
5. Initial deployment (Docker README Step 5). A fresh install does **not**
   need `FIFTYONE_DATABASE_ADMIN=true` (v2.9+ auto-initializes) — leave it
   `false`. Launch with dedicated plugins per the Step 0 standard
   recommendation. For delegated operators, use whatever Step 0 settled on —
   don't default to always-on without checking:
   - **Always-on `teams-do` workers** (the customer wants one or more
     always-on workers, on their VM/on-prem host or in Kubernetes): include
     `compose.delegated-operators.yaml` below.
   - **On-demand executors only** (Anyscale, Databricks, or in-cluster
     Kubernetes Jobs): omit `compose.delegated-operators.yaml` and follow
     `docs/configuring-on-demand-orchestrator.md` instead.
   - **Both**: include `compose.delegated-operators.yaml` for the always-on
     workers, then layer the on-demand executor on top per its doc.
   - **None** (the customer stated a specific reason): omit
     `compose.delegated-operators.yaml` and skip the on-demand docs too.

   ```shell
   docker compose \
     -f compose.dedicated-plugins.yaml \
     -f compose.delegated-operators.yaml \
     -f compose.override.yaml \
     up -d
   ```

   (omit `-f compose.delegated-operators.yaml` if the customer chose
   on-demand-only or none)

   → **Gate:** `docker compose ps` shows every container *your* Step 0 profile
   expects to be `Up` — always `fiftyone-app`, `teams-app`, `teams-api`, and
   `teams-cas`; `teams-do-*` only if always-on delegated operators were
   selected; `teams-plugins` only for Dedicated Plugins; `*-telemetry`
   sidecars and `telemetry-redis` only while telemetry is enabled (it is on by
   default, so expect them unless Step 0 turned it off). See
   [Basic Health Assessment](./docker/README.md#basic-health-assessment). Then:

   ```shell
   curl -Iv http://localhost:3030/cas/api    # expect HTTP/1.1 200 OK
   curl -Iv http://localhost:8000/health     # expect HTTP/1.1 200 OK
   curl -Iv http://localhost:3000/api/hello  # expect HTTP/1.1 200 OK
   ```

   If any container is not `Up`, follow
   [Troubleshooting Unhealthy Containers](./docker/README.md#troubleshooting-unhealthy-containers)
   and [`docker/docs/known-issues.md`](./docker/docs/known-issues.md) before
   proceeding.
6. Configure SSL & reverse proxy (Docker README Step 6) → **Gate:**
   `curl -I https://<your-domain>` returns `HTTP/2 200` from outside the host.
7. Configure the IdP / CAS (Docker README Step 7).
8. Initial CAS setup — add the first admin user, enable auto-join if wanted
   (Docker README Step 8) → **Gate:** the Super Admin UI at
   `https://<domain>/cas/configurations` accepts the `FIFTYONE_AUTH_SECRET`
   API key and the admin user is listed under `/cas/admins`.
9. Test end-user login (Docker README Step 9) → **Gate:** a non-admin user
   can log in at `https://<domain>` and lands on the FiftyOne Enterprise home
   page.
10. Work through Step 3 of this file for every feature Step 0 selected beyond
    the two standard defaults.
11. Run [Step 4 — Final validation](#step-4--final-validation).

### Kubernetes / Helm path

Mirrors [`helm/README.md`](./helm/README.md) and
[`helm/fiftyone-teams-app/README.md`](./helm/fiftyone-teams-app/README.md) —
read those for full value-key detail; this is the ordered checklist with
gates. **Unlike Docker Compose, the base Helm install does *not* default to
dedicated plugins or delegated operators** — that gap must be closed
explicitly in item 5 below, per the Step 0 standard recommendation.

1. Create the license secret (see `secret-license.template.yaml` at repo
   root) → **Gate:** `kubectl get secret <license-secret> -n <namespace>`
   exists.
2. Build your `values.yaml` / minimal overlay (never a full copy — see
   [Non-negotiable rules](#non-negotiable-rules)), setting at minimum
   `secret.fiftyone.mongodbConnectionString`, `cookieSecret`, `encryptionKey`,
   `fiftyoneAuthSecret`, and `teamsAppSettings.dnsName`/ingress host.
3. `helm repo add voxel51 https://helm.fiftyone.ai && helm repo update` →
   `helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f
   ./values.yaml` (add `cert-manager`/MongoDB-operator install steps first if
   this is a from-scratch cluster — see the
   [Full GKE Deployment Example](./helm/README.md#a-full-deployment-example-on-gke))
   → **Gate:** `kubectl rollout status deployment -n <namespace>` reports
   success for every deployment (`teams-app`, `teams-api`, `teams-cas`,
   `fiftyone-app`).
4. TLS/DNS → **Gate:** `curl -I https://<your-domain>` returns `HTTP/2 200`
   (cert issuance can take up to ~15 min; if stuck, delete the TLS secret and
   let cert-manager recreate it).
5. **Close the plugins/DO gap** — apply
   [Recommended Post-Installation Configuration](./helm/docs/post-install-recommended-configuration.md)
   to turn on Dedicated Plugins, plus whatever delegated-operator count and
   type Step 0 settled on (always-on, on-demand, or both), via a
   `helm upgrade` with your overlay updated, not a fresh `values.yaml` copy →
   **Gate:**
   [Verifying Your Setup](./helm/docs/post-install-recommended-configuration.md#verifying-your-setup)
   passes for the parts Step 0 selected — a `teams-plugins` pod is `Running`
   only if Dedicated Plugins were chosen, and a test delegated operation
   completes instead of sitting `QUEUED` only if the profile configures
   delegated operators at all (always-on *or* on-demand; for on-demand, the
   job lands on the configured executor rather than a `teams-do` pod). Skip
   whichever check the profile doesn't include.
6. Work through Step 3 of this file for every other feature Step 0 selected.
7. Run [Step 4 — Final validation](#step-4--final-validation).

## Step 3 — Feature configuration (routed by Step 0)

Only open the doc for a feature Step 0 actually selected — don't
proactively configure features nobody asked for.

| Feature | Docker doc | Helm doc |
| --- | --- | --- |
| Dedicated plugins (standard) | `docker/docs/configuring-plugins.md` | `helm/docs/configuring-plugins.md` |
| Delegated operators — always-on workers (count/platform per Step 0: VM, on-prem host, or Kubernetes) | `docker/docs/configuring-delegated-operators.md` | `helm/docs/configuring-delegated-operators.md` |
| Delegated operators — on-demand executors (count/platform per Step 0: Anyscale, Databricks, or K8s Jobs) | `docs/configuring-on-demand-orchestrator.md` + `docs/orchestrators/*` | `helm/docs/configuring-delegated-operators.md` (`delegatedOperatorJobTemplates`) |
| GPU workloads | `docker/docs/configuring-gpu-workloads.md` | `helm/docs/configuring-gpu-workloads.md` |
| Agentic Labeler (GPU builtin service) | `docker/docs/configuring-agentic-labeler.md` | `docs/configuring-service-orchestrator.md` (`serviceOrchestrators.gpuServiceOrc`) |
| Annotation AI / SAM2 (GPU builtin service) | `docs/configuring-service-orchestrator.md` | `docs/configuring-service-orchestrator.md` |
| Multimodal datasets | `docker/docs/configuring-multimodal.md` | `helm/docs/configuring-multimodal.md` |
| Telemetry (on by default; tuning/disabling) | `docker/docs/configuring-telemetry.md` | see chart `values.yaml` telemetry keys |
| Shared storage for plugins/DO (PVC/NFS) | n/a (host volume) | `helm/docs/plugins-storage.md` |
| HA `teams-api` | n/a | `helm/docs/configure-ha-teams-api.md` (requires RWX volume) |
| Snapshot archival | `docker/docs/configuring-snapshot-archival.md` | `helm/docs/configuring-snapshot-archival.md` |
| Workload Identity Federation | n/a | `helm/docs/configure-workload-identity-federation.md` |
| Custom plugin/DO images | `docs/custom-plugins.md` | `docs/custom-plugins.md` |
| Custom MongoDB permissions | `docs/custom-mongodb-permissions.md` | `docs/custom-mongodb-permissions.md` |
| Proxy configuration | `docker/docs/configuring-proxies.md` | `helm/docs/configuring-proxies.md` |
| Exposing `teams-api` for SDK access | `docker/docs/expose-teams-api.md` | `helm/docs/expose-teams-api.md` |
| AWS/EKS-specific deployment | n/a | `helm/docs/aws-deployment-guide.md` + `cloudformation/*` |

GPU sizing note: both builtin services (Agentic Labeler, Annotation AI) need a
GPU host and have real minimum VRAM requirements — see
[Accelerator sizing](./docs/configuring-service-orchestrator.md#accelerator-sizing)
before promising either feature works on whatever GPU happens to be available.

## Step 4 — Final validation

Regardless of path, the deployment isn't done until this passes:

Replace both `REPLACE_ME` values before running this — the placeholders are
sentinels, not shell syntax:

```shell
export FIFTYONE_API_URL="https://REPLACE_ME_API_URL"
export FIFTYONE_API_KEY="REPLACE_ME_API_KEY"   # generate one from the Enterprise UI  # pragma: allowlist secret
python -c 'import fiftyone.management as fom; fom.test_api_connection()'
# Expect: "API connection succeeded"
```

If the deployment's `.env` already holds a working API URL and key, source it
instead of retyping them (`set -a; . ./.env; set +a`).

Full prerequisites and troubleshooting: [`docs/validating-deployment.md`](./docs/validating-deployment.md).

## Escalating to Voxel51

Split by issue type — don't default to one channel for everything:

- **GitHub Issues** (this repo, `voxel51/fiftyone-teams-app-deploy/issues`):
  chart or compose bugs, incorrect/missing documentation, template defects —
  anything that would affect other customers too, not just this deployment.
- **The customer's existing Voxel51 support channel** (Slack Connect, support
  email, or portal — whichever was already established for this account):
  anything account-specific or actively blocking this deployment (license
  issues, credential problems, environment-specific failures).

**Before opening either, gather a diagnostic bundle:**

- Which Step/Gate in this file failed, and its exact output.
- Chart version or image tags in use — read them from
  `helm/fiftyone-teams-app/Chart.yaml` (`version`/`appVersion`) or
  `docker/common-services.yaml` (per-service image tags); don't guess or rely
  on a version mentioned elsewhere in this file, since that will go stale
  before this doc is next updated.
- `helm get values <release> -n <namespace>` output, or the customer's
  `.env`/`compose.override.yaml` — **redacted** of secrets, license contents,
  and connection strings.
- Relevant `kubectl describe pod/deploy` + `kubectl logs`, or
  `docker compose logs <service>`, for the failing component.

**Never paste secrets, license keys, or connection strings into a public
GitHub issue** — redact them from every command's output first.

## Maintaining this file

This file is deliberately thin on content it doesn't own: env var tables,
full command reference, and version-specific upgrade notes stay in
`docker/README.md`, `helm/README.md`, and their `docs/` subfolders
(`docker/docs/upgrading.md`, `helm/docs/upgrading.md`). Update this file when
the *shape, gate, or escalation* structure changes — a new deployment path, a
new standard-recommended feature, a new required prerequisite — not for every
edit to those other docs.
