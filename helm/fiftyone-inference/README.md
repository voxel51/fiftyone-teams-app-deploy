# fiftyone-inference

<!-- markdownlint-disable-next-line line-length -->
![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v2.25.1](https://img.shields.io/badge/AppVersion-v2.25.1-informational?style=flat-square)

The FiftyOne inference service: FiftyOne Enterprise sends it embeddings,
predictions, segmentation and LLM calls, and it routes them to model servers
and to your LLM endpoint. With an LLM, it also runs the shared agentic
labeling service.

FiftyOne Enterprise sends its model calls to the inference service.
The service routes each call to a backend:
a model server from the
[fiftyone-model-server](../fiftyone-model-server/README.md) chart,
or your LLM endpoint.
Install this chart in the same namespace as FiftyOne Enterprise.

## What to know before installing

- **Licenses, one per org.**
  The service reads the FiftyOne Enterprise licenses in `licenseSecrets`.
  These are the Secrets the fiftyone-teams-app chart mounts for CAS
  (`fiftyoneLicenseSecrets`, key `license`),
  one per CAS org.
  FiftyOne Enterprise names the caller's CAS org on each call,
  and the service checks the call against that org's license:
  - A call that names no org gets a 401.
  - A call whose org has no license here gets a 403.
  - An unexpired license admits every call of its org.
    An expired license admits none,
    and the other orgs keep working.

  When you renew an expired license in its Secret,
  the service reads it without a restart.
  A license for a new org needs a restart.
- **Shared token.**
  Every caller sends one token, and the service checks it.
  Set one of these:
  - `sharedToken.value`, which creates the Secret `<release>-token`
  - `sharedToken.secretName`, for your own Secret with the key `token`

  The token must be at least 32 characters,
  or the service won't start:

  ```shell
  kubectl create secret generic inference-token \
    --namespace fiftyone-teams \
    --from-literal=token="$(openssl rand -hex 32)"
  ```

  The token only proves that a caller is part of the deployment.
  Anyone who has it can call as any organization
  and change which models an organization can use.
  Give it only to FiftyOne Enterprise.
- **Token rotation.**
  The service reads the token only at start,
  and it accepts only one token.
  To rotate it, change the Secret.
  Then restart the inference service,
  the labeler,
  and the FiftyOne Enterprise pods together.
  After the restart,
  the service refuses any caller that still sends the old token.
  A change to `sharedToken.value` restarts the inference service and the labeler
  for you.
- **Model servers.**
  The service asks the model servers which models they serve (`GET /models`).
  When you upgrade a model server,
  its new models appear with no change to these values.
  Point the service at your model servers in one of three ways:
  - **Same namespace (default).**
    The chart uses the fiftyone-model-server release named `modelServer.name`
    (default `model-server`)
    with `modelServer.replicas` pods (default 1).
    It builds one URL per pod,
    `http://<name>-<i>:8080`,
    with `i` counting from 0.
    For example, `modelServer.replicas: 3` gives
    `http://model-server-0:8080`,
    `http://model-server-1:8080`
    and `http://model-server-2:8080`.
    Set `replicas` to match that release.
  - **Another namespace or cluster.**
    List one URL per box in `modelServer.urls`.
    The chart then uses this list instead.
  - **Autoscaled model servers.**
    When the model server chart uses `autoscaling`,
    set `modelServer.dns` instead.
    Each ready pod of its headless Service counts as a box.
    Pods that the autoscaler adds or removes join or leave the pool
    with no change here.

  A box doesn't need to preload every model.
  The service asks each box which models it holds (`GET /models`)
  and prefers the boxes that hold the model a call needs.
  Set `allowedModels` on each box.
  Otherwise the box answers that it holds any model
  and loads a model on its first request.

  The model server admits only pods labeled `app: inference`.
  The inference pod has this label.
  To declare no model server, set `modelServer.enabled: false`.
- **Your backends.**
  Use `backends` to add your own backends, such as an LLM,
  to the service's `BACKENDS` setting, in YAML.
- **LLM and the agentic labeler.**
  A backend of `kind: openai` is an LLM endpoint.
  It can be any OpenAI-compatible API,
  such as vLLM, Azure OpenAI or OpenAI.
  Each of its models needs `tasks: [generation]`.
  When you add one,
  the chart also runs the shared agentic labeling service (`<release>-labeler`).
  You don't need any other setting.
  The labeler calls the LLM through the inference service with the shared token.
  It picks how to call the LLM from your backends:
  - **OpenAI's own API.**
    For a backend with `openaiHosted: true` and `files: true`,
    the labeler calls the model `gpt-6-sol-responses` on the Responses API.
    It sends its examples as uploaded files
    (`remote-gpt-6-sol.json`).
    Declare that model with `wireFormat: openai-responses`,
    as in the first example below.
  - **Any other backend.**
    The labeler calls the first model of the first `kind: openai` backend
    on Chat Completions,
    with images inline
    (`remote-vllm-openai.json`,
    with `LABELER_ENGINE_CONFIG` naming the model).
    Declare that model with `wireFormat: openai-chat`,
    with a vision model behind it.
    You must also set `limits.maxImagesPerRequest`,
    or the install fails.
    This is the labeler's only limit.
    Because the labeler sends images inline,
    this limit bounds how many examples a test can carry.
    It also keeps each request under the service's `maxBodyBytes`.
    The labeler doesn't read `contextWindow`.
    The labeler constrains classification with a JSON schema
    (`response_format`),
    so the host must support that.
    We wrote the labeler's prompts and parsing for Gemma,
    so check the labels it produces on your model.

  The install fails when the backend the labeler uses
  doesn't declare the model it needs.
  Put the API key in a Secret,
  and set its variable with `secretEnv`.
- **Labeler access.**
  The labeler checks the shared token
  (`LABELER_AUTH_MODE: shared-token`),
  then trusts the organization its callers name.
  A NetworkPolicy also admits only pods in this namespace.
  So every FiftyOne Enterprise pod that calls the labeler
  must run in this namespace,
  including delegated operator jobs.
- **Model server token.**
  When the fiftyone-model-server release sets a `token`,
  set `modelServer.tokenSecret` to the same Secret (key `token`).
  The built-in model server backend then sends it as a bearer token.
  - The probes stay on `/healthz` and `/readyz`,
    which need no token.
  - Anything that scrapes a model server's `/metrics` must send the token.
  - The chart reads `modelServer.tokenSecret` from this namespace.
    For model servers in another namespace or cluster,
    copy their token Secret here.

  To change the token, change the Secret.
  Then restart the model servers and the inference service.
  The inference service reads the token only at start.
  The model server chart restarts its pods only when `token.value` changes.
- **Labeler media.**
  The labeler fetches media itself.
  By default, it fetches only from public addresses,
  over https on port 443.
  For media in a private object store,
  set these in `labeler.env`:
  - `LABELER_FETCH_ALLOWED_HOSTS`: the store's host names
  - `LABELER_FETCH_ALLOWED_NETWORKS`: the store's private address ranges
  - `LABELER_FETCH_ALLOWED_PORTS`: the store's port, when not 443
  - `LABELER_FETCH_ALLOW_HTTP`: `"true"` for plain http

  Write lists as JSON strings, not YAML lists.
  For example, for a MinIO at `http://minio:9000`:

  ```yaml
  labeler:
    env:
      LABELER_FETCH_ALLOWED_HOSTS: '["minio"]'
      LABELER_FETCH_ALLOWED_NETWORKS: '["10.0.0.0/8"]'
      LABELER_FETCH_ALLOWED_PORTS: '[9000]'
      LABELER_FETCH_ALLOW_HTTP: "true"
  ```

- **Scheduling.**
  `nodeSelector`, `tolerations`, `affinity`,
  `podSecurityContext` and `securityContext`
  apply to both pods.
- **One replica each.**
  Run one replica of each pod.
  The default `memory` store and the rate limits are per replica,
  and a labeler job lives in the pod that took it.

## Install

```shell
helm install inference ./helm/fiftyone-inference \
  --namespace fiftyone-teams \
  --set sharedToken.secretName=inference-token
```

With an LLM, the agentic labeler runs too.
This example uses OpenAI's own API:

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

This example uses another OpenAI-compatible API,
a vLLM server.
The labeler calls `qwen3-vl`,
the first model:

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

Set these values in the fiftyone-teams-app chart.
This example uses a release named `inference`
with the token in the Secret `inference-token`.
FiftyOne Enterprise must run in the same namespace as this chart,
so the short Service names below resolve.
Their full names are
`inference.<namespace>.svc.cluster.local`
and `inference-labeler.<namespace>.svc.cluster.local`,
such as `http://inference.fiftyone-teams.svc.cluster.local:8080`.

```yaml
inference:
  url: http://inference:8080
  tokenSecret: inference-token
  labeler:
    # With an LLM only
    url: http://inference-labeler:8080
```

The fiftyone-teams-app chart then sets these variables on:

- `fiftyone-app`, which serves the App's model features
- the dedicated plugins
- each delegated operator deployment
- each delegated operator job template
- each service orchestrator

All but `fiftyone-app` run the model operators.
A variable you set in a component's own `env` or `secretEnv` wins.

| Variable                       | Value                                                     |
| ------------------------------ | --------------------------------------------------------- |
| `FIFTYONE_INFERENCE_URL`       | `inference.url`                                           |
| `FIFTYONE_INFERENCE_TOKEN`     | the key `token` of the Secret `inference.tokenSecret`     |
| `FIFTYONE_INFERENCE_TENANT_ID` | `inference.tenantId`, when set                            |
| `LABELER_SERVICE_URL`          | `inference.labeler.url`, with an LLM only                 |
| `LABELER_SERVICE_AUTH`         | `static-token`, with an LLM only: sends the shared token  |
| `LABELER_TEST_TRANSPORT`       | `operator`, with an LLM only: Run test uses this labeler  |
| `LABELER_RUN_MODE`             | `job`, with an LLM only: Run is a delegated operator      |
| `LABELER_DELEGATION_TARGET`    | `agentic-labeler`, with an LLM only: the job that runs it |

### The org of each call

FiftyOne Enterprise sends the caller's CAS org (`user.organization.id`)
as the tenant of each call.
Never set `FIFTYONE_TENANT_ID`.
It overrides that org
and makes FiftyOne Enterprise act as a SaaS tenant.

`FIFTYONE_INFERENCE_TENANT_ID` is only the fallback for calls with no user,
such as built-in services that the reconcile starts:

- **Several orgs:** leave `inference.tenantId` empty.
  If you set it, every labeler run also goes out as that one org.
- **One org:** set `inference.tenantId` to that org's ID.
  This is the `organizationId` of its license,
  the same as `user.organization.id`.
  Never use a made-up name.

### Labeler runs

With an LLM, each Run is a Kubernetes Job on CPU
from the job template `agentic-labeler`.
The Job calls the LLM through this service.
Add the job template
and turn off the chart's service orchestrators.
This service replaces them,
and a per-user labeler service left running takes runs too.

The labeler image tag is this chart's `appVersion`,
the same as for `labeler.image`.
The example below fits OpenAI with files.
For any other backend,
copy `LABELER_CONFIG_FILE` and `LABELER_ENGINE_CONFIG`
from the `<release>-labeler` Deployment instead.
Then add this key to that engine config,
so the Job names the org of each run.
The shared labeler gets the org from its callers,
but the Job needs this key:

```text
"default_headers_env": {
  "x-voxel-tenant-id": "FIFTYONE_INFERENCE_TENANT_ID",
  "x-voxel-principal-id": "FIFTYONE_INFERENCE_PRINCIPAL_ID"
}
```

```yaml
delegatedOperatorJobTemplates:
  jobs:
    agentic-labeler:
      image:
        repository: voxel51/agentic-labeler-cpu
        tag: v2.25.1
      resources:
        limits:
          cpu: 2
          memory: 8Gi
        requests:
          cpu: 2
          memory: 8Gi
      env:
        LABELER_CONFIG_FILE: /app/configs/remote-gpt-6-sol.json
        LABELER_INTAKE: none
        LABELER_TENANCY: env
        MALLOC_ARENA_MAX: "2"
  serviceOrchestrators:
    cpuServiceOrc:
      enabled: false
    gpuServiceOrc:
      enabled: false
```

The Job also gets the inference variables
and the template's `env` and `secretEnv`.
Each run names the org and user who started it.
The chart registers the Job as an orchestrator in a post-upgrade hook.
Argo CD runs that hook only after a healthy sync.
Before you upgrade,
stop any per-user labeler service that's still running.

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
| modelServer.dns | bool | `false` | Find the model servers by DNS instead of `replicas`, `urls` or `endpointsSecret`: each ready pod of the release `name`'s headless Service `<name>-headless` is a box, so pods its autoscaler adds and removes join and leave the pool with no change here. Needs an inference image with ENDPOINTS_DNS. |
| modelServer.enabled | bool | `true` | Declare the model server backend: the service reads its models from the model servers, so a model server upgrade brings the new ones. Off leaves only `backends`. |
| modelServer.endpointsSecret | string | `""` | Existing Secret holding the model servers in an endpoint file, under the key `endpoints.json`, used instead of `name`, `replicas` and `urls` when set: for model servers behind a gateway. The service re-reads it, so a box list change needs no restart. Each entry names the backend `model-server`: `{"endpoints": [{"id": "box-0", "backend": "model-server", "url": "http://..."}]}`. |
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
