# fiftyone-model-server

<!-- markdownlint-disable-next-line line-length -->
![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

The FiftyOne model server: the GPU boxes the FiftyOne inference service
sends embeddings, predictions and segmentation to. One pod is one box, in
the same cluster as FiftyOne Enterprise or in a separate GPU cluster.
Preview: the image is not published yet.

The FiftyOne inference service routes embeddings, predictions
and segmentation to model servers.
This chart deploys them,
one box per pod.
The inference service reaches a box only by URL,
so a box can run in a separate GPU cluster.

## What to know before installing

- **The model server checks no token by default.**
  To require one, set either:
  - `token.value`
  - `token.secretName`, for your own Secret with the key `token`

  Every call then needs `Authorization: Bearer <token>`,
  except the probes (`/healthz` and `/readyz`).
  This includes `/metrics`,
  so anything that scrapes it must send the token.
  Give the same Secret to the fiftyone-inference chart
  in `modelServer.tokenSecret`,
  and it sends the token for you.

  To change the token, change the Secret.
  Then restart the model servers and the inference service.
  Both read the token only at start.
  A change to `token.value` restarts only the model servers.

  With or without a token,
  only the inference service should reach the model server:
  - **Same cluster.**
    The default NetworkPolicy admits only pods labeled `app: inference`
    in this namespace.
    For an inference service in another namespace,
    add a `namespaceSelector` to the peer in `networkPolicy.from`.
  - **Another cluster.**
    Set `service.type: LoadBalancer`
    and set `service.loadBalancerSourceRanges`
    to the calling cluster's pod range.
    Add your cloud's annotation for an internal load balancer.
    Each pod gets its own load balancer.
    The NetworkPolicy then admits those ranges too.
    Set `service.externalTrafficPolicy: Local`
    so the NetworkPolicy sees the caller's real address.
- **One box per pod.**
  `replicas` sets how many pods run
  (default 1).
  The inference service limits how many video runs each box takes,
  and it keeps a run on the box that started it.
  So pods don't share one URL.
  Each pod has its own Service, `<name>-<i>`,
  with `i` counting from 0.
  `<name>` is the release name or `nameOverride`.
  For example, release `model-server` with `replicas: 3` gives
  `http://model-server-0:8080`,
  `http://model-server-1:8080`
  and `http://model-server-2:8080`.
  List them all in the fiftyone-inference chart
  with `modelServer.replicas` or `modelServer.urls`.
  `replicas` doesn't autoscale pods.
  To autoscale, see [Autoscaling](#autoscaling).
- **You must set `preload`.**
  The pod reports Ready only after a model loads.
  Nothing answers until preload finishes,
  so the first start includes downloading the weights.
- **Weights.**
  By default, the weights go to an `emptyDir`:
  - A new pod downloads them again.
  - They count toward the pod's ephemeral storage,
    so request enough of it.
    GKE Autopilot gives 1Gi otherwise.

  To keep the weights on a claim instead,
  set `weights.persistence.enabled`.
  Each pod gets one claim,
  `weights-<name>-<i>`:
  - When you scale in,
    the claims of the removed pods stay.
    When you scale out again,
    new pods reuse them.
    Delete a claim yourself once you no longer need it.
  - `helm uninstall` keeps the claims too,
    and a reinstall uses them again.
  - `weights.persistence.existingClaim` supports one pod only,
    because pods can't share a ReadWriteOnce claim.
  - You can't change `weights.persistence`
    `enabled`, `size`, `storageClass` or `existingClaim`
    after install.
    To change them, run
    `kubectl delete statefulset <name> --cascade=orphan`,
    then upgrade.
    Existing claims keep their old size until you resize them yourself.
- **GPU.**
  `gpu.enabled` requests `gpu.count` GPUs.
  Use `nodeSelector` and `tolerations` to place pods on your GPU nodes.
  Small models such as SigLIP work on CPU.
  SAM2 and video models are too slow on CPU
  for the inference service's timeouts.

## Install

```shell
helm install model-server ./helm/fiftyone-model-server \
  --namespace fiftyone-teams \
  --set image.repository=<repository> \
  --set image.tag=<tag> \
  --set 'preload[0]=siglip-base-patch16-224-torch'
```

## Autoscaling

Set `autoscaling.enabled`
to let a HorizontalPodAutoscaler set the number of pods.
It keeps the count between `autoscaling.minReplicas`
and `autoscaling.maxReplicas`.
By default, it scales on CPU use,
so set a CPU request in `resources`.
To scale on other metrics, set `autoscaling.metrics`.

Point the inference chart at the pods by DNS
with `modelServer.dns`.
The inference service resolves the headless Service `<name>-headless`,
which lists the ready pods.
New pods join its pool and stopping pods leave it,
with no change to its values.

When a pod stops,
because it scaled in or got replaced:

1. It waits `drain.preStopSeconds`
   while callers stop sending it new work.
1. It finishes its current work,
   for up to `drain.terminationGracePeriodSeconds`.

Turning on `autoscaling.enabled`
for a release that ran with `replicas`
stops every pod but the first.
The upgrade removes the StatefulSet's replica count,
which then defaults to 1.
The other pods drain and restart,
and their models load again.
Set `autoscaling.minReplicas` to at least the old `replicas`
to bring them back as soon as they stop,
rather than when load calls for them.
Upgrade at a quiet time.

Rollouts replace one pod at a time,
and each pod waits for its runs.
So give `helm upgrade --wait` a longer `--timeout`
than its default 5 minutes.

## Serve another zoo model

The image declares the models it ships with.
To serve another zoo model:

1. Declare it in `declarations`.
1. If you set `allowedModels`, add it there too.
   A model left out of that list drops out of the catalog.

The inference service adds the model to its catalog within about a minute,
and the App lists it.
You don't need to rebuild the image.
This example adds CLIP as an embedding model
next to the image's models:

```yaml
allowedModels:
  - clip-vit-base32-torch
  - qwen3-vl-embedding-2b-torch
  - segment-anything-2-hiera-large-image-torch
  - segment-anything-2-hiera-large-video-torch
  - siglip-base-patch16-224-torch
declarations:
  - name: clip-vit-base32-torch
    tasks: [embedding, prediction]
    dimensions: 512
    maxMediaPerRequest: 32
    supportsText: true
    supportsLogits: true
    modelKwargs: [classes, text_prompt]
    tags: [embeddings, zero-shot]
```

This works for zoo models that meet both conditions:

- The model server runs the model's task:
  embedding, prediction or propagation.
- The image contains the model's packages.

The model server won't start
if a name isn't in the zoo
or a task isn't one it runs.

## Requirements

Kubernetes: `>=1.31-0`

## Values

<!-- markdownlint-disable MD060 -->

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity. |
| allowedMediaPrefixes | list | `[]` | Media locations this box fetches from, as scheme and host prefixes such as `gs://my-bucket`. Empty means any URL, including internal addresses. |
| allowedModelKwargs | list | `[]` | Model arguments this box passes to the zoo loader. Empty means any. Set it when anything besides the inference service can reach the box: some loader arguments are filesystem paths. |
| allowedModels | list | `[]` | Models this box serves at all. Empty means any zoo model, which lets a caller fill the disk with weights. Set it to the same list as the inference service declares for this box. |
| automountServiceAccountToken | bool | `false` | Mount the service account token. Off: the model server calls no Kubernetes API. |
| autoscaling.enabled | bool | `false` | Scale the pods with a HorizontalPodAutoscaler instead of `replicas`. Point the inference chart at them by DNS (`modelServer.dns`), so boxes join and leave its pool as they come and go. |
| autoscaling.maxReplicas | int | `4` | Most pods. A Service `<name>-<i>` exists for each, ready for the pod. |
| autoscaling.metrics | list | `[]` | Metrics to scale on instead, as the HorizontalPodAutoscaler's `metrics` (autoscaling/v2). Pod or external metrics, such as GPU use, need a metrics adapter in the cluster. |
| autoscaling.minReplicas | int | `1` | Fewest pods. |
| autoscaling.targetCPUUtilizationPercentage | int | `70` | Average CPU use to keep across pods, in percent of each pod's CPU request, which `resources` must then set. Used when `metrics` is empty. |
| declarations | list | `[]` | Declarations of zoo models the image does not declare, in the fields of the model server's `catalog.py` (`name`, `tasks` and the rest): the inference service adds them to its catalog, and the App lists them, with no rebuild. One named as a built-in replaces it. Allow each in `allowedModels` when that is set. See the README for an example. |
| drain.preStopSeconds | int | `45` | Seconds a stopping pod keeps serving before it stops taking new connections: callers that find boxes by DNS stop sending it new work first. Longer than the inference service's DNS refresh (10 s) plus the cluster DNS cache (often 30 s). |
| drain.terminationGracePeriodSeconds | int | `600` | Seconds a stopping pod gets to finish what it is running, video propagations included, before it is killed. |
| env | object | `{"FIFTYONE_SAM2_FEATURE_CACHE_SIZE":"16","FIFTYONE_SAM2_LAZY_FRAMES":"1","FIFTYONE_SAM2_PROPAGATION_CONCURRENCY":"4","FIFTYONE_SAM2_VIDEO_CACHE_SIZE":"4"}` | Environment variables, by name. The defaults are the SAM2 video settings of the SaaS boxes: frames decoded on demand rather than the whole clip, and the propagation and cache sizes of a 48 GB GPU. Set one to `null` to drop it. |
| gpu.count | int | `1` | GPUs requested when `gpu.enabled`. |
| gpu.enabled | bool | `false` | Request GPUs. Off runs on CPU, which is enough for small models such as SigLIP and too slow for SAM2 or video. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| image.repository | string | `nil` | Container image repository. Required: the image is not published yet. |
| image.tag | string | `nil` | Container image tag. Required. |
| imagePullSecrets | list | `[]` | Image pull secrets for a private registry. |
| maxLoadedModels | int | `6` | Models kept loaded at once, least recently used evicted first. |
| nameOverride | string | `""` | Overrides the release name in every object's name. |
| networkPolicy.enabled | bool | `true` | Admit only `networkPolicy.from` to the box. The model server checks no token, so leave this on unless the cluster enforces nothing. |
| networkPolicy.from | list | `[{"podSelector":{"matchLabels":{"app":"inference"}}}]` | NetworkPolicy peers allowed to call the box: by default the inference service's pods in this namespace. Add a `namespaceSelector` to the peer for an inference service in another namespace. With a LoadBalancer, the `service.loadBalancerSourceRanges` are admitted as well. |
| nodeSelector | object | `{}` | Node selector, for GPU nodes for example. |
| podAnnotations | object | `{}` | Pod annotations. |
| podSecurityContext | object | `{}` | Pod security context. |
| preload | list | `[]` | Models loaded at startup, by zoo name. Required: `/readyz` answers 503 until one is loaded, so with none the pod never becomes Ready. Name the models the inference service routes to this box. |
| replicas | int | `1` | Model server pods. Each is its own box, reached at its own Service `<name>-<i>` (`http://<name>-<i>:8080`, i from 0): list every URL in the inference chart. Ignored with `autoscaling.enabled`. |
| resources | object | `{"limits":{},"requests":{}}` | Container resources. With `gpu.enabled`, `nvidia.com/gpu` is set to `gpu.count` in requests and limits. |
| securityContext | object | `{}` | Container security context. |
| service.annotations | object | `{}` | Service annotations, such as `networking.gke.io/load-balancer-type: Internal`. |
| service.externalTrafficPolicy | string | `"Local"` | `externalTrafficPolicy` of a LoadBalancer. `Local` keeps the caller's address, which the NetworkPolicy needs to admit it by range. |
| service.loadBalancerSourceRanges | list | `[]` | Source ranges a LoadBalancer admits, such as the calling cluster's pod range. Without `token`, the model server checks no token, so this is its access control. The NetworkPolicy admits the same ranges. |
| service.type | string | `"ClusterIP"` | `ClusterIP` for callers in this cluster. `LoadBalancer` for an inference service in another cluster, which then requires `service.loadBalancerSourceRanges`. |
| serviceAccountName | string | `nil` | Service account the pod runs as. |
| startupProbe.failureThreshold | int | `90` | Startup probe attempts, 10 s apart. Nothing listens until the preload is done, downloads included, so the default allows 15 minutes. |
| token.secretName | string | `""` | Existing Secret holding a bearer token under the key `token`. When a token is set, every call but `/healthz` and `/readyz` needs `Authorization: Bearer <token>`, `/metrics` included. Give the same Secret to the fiftyone-inference chart (`modelServer.tokenSecret`). |
| token.value | string | `""` | The token, used to create the Secret `<name>-token` when `token.secretName` is empty. Empty, with no `secretName`, means no token. |
| tolerations | list | `[]` | Tolerations, for tainted GPU nodes for example. |
| weights.persistence.enabled | bool | `false` | Keep downloaded weights on a PersistentVolumeClaim per pod, so a restart does not download them again and the node's disk is not filled. Off uses an emptyDir. |
| weights.persistence.existingClaim | string | `nil` | Use this claim instead of a claim per pod. With `replicas: 1` only: a ReadWriteOnce claim cannot be shared. |
| weights.persistence.size | string | `"30Gi"` | Size of the created claim. |
| weights.persistence.storageClass | string | `nil` | Storage class of the created claim. Empty is the cluster default. |
| weights.sizeLimit | string | `nil` | emptyDir size limit, when persistence is off. |

<!-- markdownlint-enable MD060 -->
