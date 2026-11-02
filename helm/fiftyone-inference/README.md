# fiftyone-inference

<!-- markdownlint-disable-next-line line-length -->
![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v2.25.1](https://img.shields.io/badge/AppVersion-v2.25.1-informational?style=flat-square)

The FiftyOne inference service: FiftyOne Enterprise sends it embeddings,
predictions, segmentation and LLM calls, and it routes them to model servers
and to your LLM endpoint. With an LLM, it also runs the shared agentic
labeling service.

FiftyOne Enterprise sends its model calls to the inference service, which
routes each one to a backend: a model server from the
[fiftyone-model-server](../fiftyone-model-server/README.md) chart, or your LLM
endpoint. Install this chart in the namespace of FiftyOne Enterprise.

## What to know before installing

- **Licenses, one per org.** The service answers to the FiftyOne Enterprise
  licenses in `licenseSecrets`, the Secrets the fiftyone-teams-app chart
  mounts for CAS (`fiftyoneLicenseSecrets`, key `license`), one per CAS org.
  Each call is admitted against the license of the org it names: FiftyOne
  Enterprise names the caller's CAS org. A call that names no org is a 401,
  and one whose org has no license here a 403. An unexpired license admits
  every call of its org, an expired one none, while the other orgs keep
  working. After a license expires, a renewed one in its Secret is read
  without a restart. A new org's license needs a restart.
- **Shared token.** Every caller sends one token, which the service checks.
  Set `sharedToken.value`, which creates the Secret `<release>-token`, or
  `sharedToken.secretName` for your own Secret with the key `token`. The token
  must be at least 32 characters, or the service refuses to start:

  ```shell
  kubectl create secret generic inference-token \
    --namespace fiftyone-teams \
    --from-literal=token="$(openssl rand -hex 32)"
  ```

  The token proves a caller is part of the deployment, nothing more. Any holder
  can call as any organization, and change which models an organization may
  use. Give it only to FiftyOne Enterprise.
- **Token rotation.** The service reads the token at start only, and accepts
  one token. Change the Secret, then restart the inference service, the
  labeler and the FiftyOne Enterprise pods together: a caller with the old
  token is refused from the restart on. A change to `sharedToken.value`
  restarts the inference service and the labeler.
- **Model servers.** The chart ships the declarations of every model the
  voxel51 model server image serves (`catalog/model-server.yaml`), so a
  chart upgrade brings new models with no values edit. It points them at
  the fiftyone-model-server release `modelServer.name` in this namespace,
  `model-server` by default, with `modelServer.replicas` pods, 1 by
  default: one URL per pod, `http://<name>-<i>:8080` for i from 0. With
  `modelServer.replicas: 3`, these are `http://model-server-0:8080`,
  `http://model-server-1:8080` and `http://model-server-2:8080`. Set the same
  `replicas` as that release. For model servers in another namespace or
  cluster, list one URL per box in `modelServer.urls`, which is then used
  instead. A box need not preload every
  model: the service asks each box which models it holds (`GET /models`), and
  for each model prefers the boxes that hold it. Set `allowedModels` on each
  box, or a box answers that it holds any model and loads one on its first
  request. The model server admits only pods labelled `app: inference`, which
  the inference pod is. Set `modelServer.enabled: false` to declare no model server.
- **Your backends.** `backends` adds your own to the service's `BACKENDS`
  setting, in YAML, such as an LLM.
- **LLM and the agentic labeler.** A backend of `kind: openai` is an LLM
  endpoint: any OpenAI compatible API, such as vLLM, Azure OpenAI or OpenAI.
  Each of its models needs `tasks: [generation]`. With one, the chart also
  runs the shared agentic labeling service (`<release>-labeler`), with no
  other setting. It calls the LLM through the inference service, with the
  shared token, and picks how from the backends:
  - A backend with `openaiHosted: true` and `files: true` (OpenAI's own API):
    the model `gpt-6-sol-responses` on the Responses API, with its examples
    as uploaded files (`remote-gpt-6-sol.json`). Declare that model with
    `wireFormat: openai-responses`, as in the first example below.
  - Otherwise: the first model of the first `kind: openai` backend on Chat
    Completions, with images inline (`remote-vllm-openai.json`, and
    `LABELER_ENGINE_CONFIG` naming the model). Declare that model with
    `wireFormat: openai-chat` and a vision model behind it, and with
    `limits.maxImagesPerRequest`, which the install requires. It is the
    labeler's only limit: images are sent inline, so it bounds how many
    examples a test may carry and each request under the service's
    `maxBodyBytes`. The labeler does not read `contextWindow`. It
    constrains classification with a JSON schema (`response_format`), which
    the host must support. Its prompts and parsing
    are written for Gemma, so check the labels on your model.

  The install fails when the backend the labeler uses declares no such
  model. Put the API key in a Secret, and set its variable with `secretEnv`.
- **Labeler access.** The labeler checks the shared token
  (`LABELER_AUTH_MODE: shared-token`), then trusts the organization its
  callers name. A NetworkPolicy also admits only pods in this namespace:
  every FiftyOne Enterprise pod that calls it, delegated operator jobs
  included, must run in this namespace.
- **Model server token.** When the fiftyone-model-server release sets a
  `token`, set `modelServer.tokenSecret` to the same Secret (key `token`):
  the shipped backend then sends it as a bearer. Probes stay on `/healthz`
  and `/readyz`, which need no token. Anything that scrapes a model server's
  `/metrics` must send it. `modelServer.tokenSecret` is read in this
  namespace: for model servers in another namespace or cluster, copy their
  token Secret here. To change the token, change the Secret, then restart
  the model servers and the inference service: the inference service reads
  it at start only, and the model server chart restarts its pods only for a
  change to its `token.value`.
- **Labeler media.** The labeler fetches media itself, by default only from
  public addresses, over https on port 443. For media in a private object
  store, set in `labeler.env`:
  - `LABELER_FETCH_ALLOWED_HOSTS`: the store's host names
  - `LABELER_FETCH_ALLOWED_NETWORKS`: the store's private address ranges
  - `LABELER_FETCH_ALLOWED_PORTS`: its port, when not 443
  - `LABELER_FETCH_ALLOW_HTTP`: `"true"` for plain http

  Lists are JSON strings, not YAML lists. A MinIO at `http://minio:9000`:

  ```yaml
  labeler:
    env:
      LABELER_FETCH_ALLOWED_HOSTS: '["minio"]'
      LABELER_FETCH_ALLOWED_NETWORKS: '["10.0.0.0/8"]'
      LABELER_FETCH_ALLOWED_PORTS: '[9000]'
      LABELER_FETCH_ALLOW_HTTP: "true"
  ```

- **Scheduling.** `nodeSelector`, `tolerations`, `affinity`,
  `podSecurityContext` and `securityContext` apply to both pods.
- **One replica each.** The default `memory` store and the rate limits are per
  replica, and a labeler job lives in the pod that took it.

## Install

```shell
helm install inference ./helm/fiftyone-inference \
  --namespace fiftyone-teams \
  --set sharedToken.secretName=inference-token
```

With an LLM, the agentic labeler runs too. OpenAI's own API:

```yaml
sharedToken:
  secretName: inference-token
backends:
  - kind: openai
    id: openai
    baseUrl: https://api.openai.com/v1
    auth:
      mode: bearer-token
      tokenEnv: OPENAI_API_KEY
    openaiHosted: true
    files: true
    models:
      - name: gpt-6-sol-responses
        backendModelId: gpt-6-sol
        contextWindow: 1050000
        tasks: [generation]
        wireFormat: openai-responses
        allowedTemplateKwargs: []
        limits:
          maxImagesPerRequest: 1500
secretEnv:
  OPENAI_API_KEY:
    secretName: openai
    secretKey: apiKey
```

Any other OpenAI compatible API, here a vLLM server. The labeler calls
`qwen3-vl`, the first model:

```yaml
sharedToken:
  secretName: inference-token
backends:
  - kind: openai
    id: vllm
    baseUrl: http://vllm.internal:8000/v1
    auth:
      mode: bearer-token
      tokenEnv: VLLM_API_KEY
    models:
      - name: qwen3-vl
        backendModelId: Qwen/Qwen3-VL-32B-Instruct
        tasks: [generation]
        wireFormat: openai-chat
        limits:
          maxImagesPerRequest: 64
secretEnv:
  VLLM_API_KEY:
    secretName: vllm
    secretKey: apiKey
```

## Connect FiftyOne Enterprise

Set these in the values of the fiftyone-teams-app chart, for a release named
`inference` with the token in the Secret `inference-token`. No chart change is
needed: every component takes `env` and `secretEnv`. FiftyOne Enterprise must
run in the namespace of this chart. The short Service names below resolve
there. Their full names are `inference.<namespace>.svc.cluster.local` and
`inference-labeler.<namespace>.svc.cluster.local`, such as
`http://inference.fiftyone-teams.svc.cluster.local:8080`.

| Variable                   | Value                                                       |
| -------------------------- | ----------------------------------------------------------- |
| `FIFTYONE_INFERENCE_URL`   | `http://inference:8080`                                     |
| `FIFTYONE_INFERENCE_TOKEN` | the shared token, from its Secret                           |
| `LABELER_SERVICE_URL`      | `http://inference-labeler:8080`, with an LLM only           |
| `LABELER_SERVICE_AUTH`     | `static-token`, with an LLM only: sends the shared token    |

```yaml
appSettings:
  env: &inferenceEnv
    FIFTYONE_INFERENCE_URL: http://inference:8080
    LABELER_SERVICE_URL: http://inference-labeler:8080
    LABELER_SERVICE_AUTH: static-token
  secretEnv: &inferenceSecretEnv
    FIFTYONE_INFERENCE_TOKEN:
      secretName: inference-token
      secretKey: token
pluginsSettings:
  env: *inferenceEnv
  secretEnv: *inferenceSecretEnv
delegatedOperatorDeployments:
  template:
    env: *inferenceEnv
    secretEnv: *inferenceSecretEnv
delegatedOperatorJobTemplates:
  template:
    env: *inferenceEnv
    secretEnv: *inferenceSecretEnv
```

`fiftyone-app` serves the App's model features, and the plugins and delegated
operators run the model operators, so each needs the variables.
`pluginsSettings` applies with dedicated plugins only.

**The org of each call.** FiftyOne Enterprise names the caller's CAS org
(`user.organization.id`) as the tenant of each call. Never set
`FIFTYONE_TENANT_ID`: it overrides that org, and makes FiftyOne Enterprise
act as a SaaS tenant. `FIFTYONE_INFERENCE_TENANT_ID` is only the fallback for
calls with no user, such as builtin services started by the reconcile:

- Several orgs: leave it unset.
- One org: add `FIFTYONE_INFERENCE_TENANT_ID: <org id>` to the
  `inferenceEnv` anchor above, set to that org's id, the `organizationId` of
  its license (the same as `user.organization.id`), never a made up name.
  Do not repeat `delegatedOperatorJobTemplates` or
  `delegatedOperatorDeployments` elsewhere in the values: Helm keeps the last
  of two equal keys, and the other would drop the anchors.

Register service orchestrators through
`delegatedOperatorJobTemplates.serviceOrchestrators`, so their pods get the
template's env. A service pod with no template gets only the `valueFrom` and
`envFrom` env of `teams-api`, and the block above sets nothing on
`apiSettings`.

## Requirements

Kubernetes: `>=1.31-0`

## Values

<!-- markdownlint-disable MD060 -->

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity of both pods. |
| backends | list | `[]` | Your own backends, added to the shipped model server backend in `BACKENDS`. A backend of `kind: openai` is an LLM endpoint (any OpenAI compatible API), and turns on the agentic labeler. On an OpenAI backend with `openaiHosted` and `files`, the labeler uses its model `gpt-6-sol-responses` (`wireFormat: openai-responses`). Otherwise it uses the first model of the first `kind: openai` backend, which needs `wireFormat: openai-chat` and `limits.maxImagesPerRequest`. Either needs `generation` in its tasks. A backend names its token's variable in `auth.tokenEnv`, set from a Secret with `secretEnv`. See the README for LLM examples. |
| env | object | `{}` | Extra environment variables of the inference service, by name. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy of the inference service and the labeler. |
| image.repository | string | `"voxel51/inference"` | Inference service image repository. |
| image.tag | string | `""` | Inference service image tag. Defaults to the chart `appVersion`. |
| imagePullSecrets | list | `[]` | Image pull secrets for a private registry. |
| labeler.env | object | `{"LABELER_MAX_IN_FLIGHT_TESTS":"1"}` | Labeler environment variables, by name. They win over the ones the chart sets. The labeler fetches media only from public addresses on port 443 by default. For media in a private object store, set `LABELER_FETCH_ALLOWED_HOSTS`, `LABELER_FETCH_ALLOWED_NETWORKS`, `LABELER_FETCH_ALLOWED_PORTS` and, for plain http, `LABELER_FETCH_ALLOW_HTTP: "true"`. Lists are JSON strings, such as `'["minio.internal"]'`, not YAML lists. |
| labeler.image.repository | string | `"voxel51/agentic-labeler-cpu"` | Labeler image repository. |
| labeler.image.tag | string | `""` | Labeler image tag. Defaults to the chart `appVersion`. |
| labeler.resources | object | `{"limits":{},"requests":{}}` | Labeler container resources. |
| licenseSecrets | list | `["fiftyone-license"]` | Existing Secrets holding the FiftyOne Enterprise licenses, one per org, under the key `license`: the ones the fiftyone-teams-app chart mounts for CAS (its `fiftyoneLicenseSecrets`). Each call is admitted against the license of the org it names: unexpired, it admits every call, expired, none. A new org's license needs a restart. |
| maxBodyBytes | int | `67108864` | Largest request body in bytes. Images are sent inline, so this is large. |
| modelServer.enabled | bool | `true` | Declare the model server backend this chart ships: every model the voxel51 model server image serves (`catalog/model-server.yaml`). A chart upgrade brings the new ones. Off leaves only `backends`. |
| modelServer.name | string | `"model-server"` | Name of the fiftyone-model-server release in this namespace: its release name, or its `nameOverride`. |
| modelServer.replicas | int | `1` | Its `replicas`. The model servers are then `http://<name>-<i>:8080` for i from 0, one per pod. |
| modelServer.tokenSecret | string | `""` | Existing Secret holding the model servers' bearer token under the key `token`: the fiftyone-model-server chart's `token` (its Secret `<name>-token` from `token.value`, or your own). Empty: the model servers check no token. |
| modelServer.urls | list | `[]` | The model servers, one URL per box, used instead of `name` and `replicas` when set: for model servers in another namespace or cluster. |
| nodeSelector | object | `{}` | Node selector of both pods. |
| podSecurityContext | object | `{}` | Pod security context of both pods, added to `runAsNonRoot: true`, `runAsUser: 1000` and `runAsGroup: 1000`. Those cannot be set to false or 0. |
| resources | object | `{"limits":{},"requests":{}}` | Inference service container resources. |
| secretEnv | object | `{}` | Environment variables of the inference service from existing Secrets, such as the LLM token a backend's `auth.tokenEnv` names. |
| securityContext | object | `{}` | Container security context of both pods, added to `allowPrivilegeEscalation: false`. |
| sharedToken.secretName | string | `""` | Existing Secret holding the shared token under the key `token`. Every caller sends this token: FiftyOne Enterprise and the labeler. |
| sharedToken.value | string | `""` | The shared token, used to create the Secret `<release>-token` when `sharedToken.secretName` is empty. At least 32 characters, or the service refuses to start. Generate one with `openssl rand -hex 32`. |
| storeMode | string | `"memory"` | Where the catalog and per organization model settings live. `memory` loses the settings at restart. `mongo` needs `MONGO_URI` in `secretEnv`. |
| tolerations | list | `[]` | Tolerations of both pods, for tainted nodes for example. |

<!-- markdownlint-enable MD060 -->
