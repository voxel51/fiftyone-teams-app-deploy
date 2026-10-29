# fiftyone-model-server

<!-- markdownlint-disable-next-line line-length -->
![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

The FiftyOne model server: the GPU box the FiftyOne inference service sends
embeddings, predictions and segmentation to. One release is one box, in the
same cluster as FiftyOne Enterprise or in a separate GPU cluster.
Preview: the image is not published yet.

The FiftyOne inference service routes embeddings, predictions and
segmentation to model servers. This chart deploys one: install one release
per box. The inference service reaches a box by URL only, so a box can run in
a separate GPU cluster.

## What to know before installing

- **The model server checks no token.** Only the inference service should
  reach it:
  - In the same cluster, the default NetworkPolicy admits only pods labelled
    `app: inference` in this namespace. For an inference service in another
    namespace, add a `namespaceSelector` to the peer in `networkPolicy.from`.
  - From another cluster, use `service.type: LoadBalancer` with
    `service.loadBalancerSourceRanges` set to the calling cluster's pod range,
    and an internal load balancer annotation for your cloud. The
    NetworkPolicy then admits those ranges too, and
    `service.externalTrafficPolicy: Local` keeps the caller's address so it
    can.
- **One box per release.** The inference service limits how many video runs
  each box takes and keeps a run on the box that started it, so the chart
  runs one replica. Add capacity with another release, and list its URL in
  the inference service's backend.
- **`preload` is required.** The pod reports Ready only once a model is
  loaded. Nothing answers until the preload is done, so the first start
  includes downloading the weights.
- **Weights.** By default they go to an `emptyDir`: a new pod downloads them
  again, and they count as the pod's ephemeral storage, so request enough of
  it (GKE Autopilot gives 1Gi otherwise). Set `weights.persistence.enabled`
  to keep them on a claim instead.
- **GPU.** `gpu.enabled` requests `gpu.count` GPUs. Use `nodeSelector` and
  `tolerations` for your GPU nodes. On CPU, small models such as SigLIP work,
  but SAM2 and video models are too slow for the inference service's
  timeouts.

## Install

```shell
helm install model-server ./helm/fiftyone-model-server \
  --namespace fiftyone-teams \
  --set image.repository=<repository> \
  --set image.tag=<tag> \
  --set 'preload[0]=siglip-base-patch16-224-torch'
```

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
| env | object | `{}` | Extra environment variables, by name. |
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
| resources | object | `{"limits":{},"requests":{}}` | Container resources. With `gpu.enabled`, `nvidia.com/gpu` is set to `gpu.count` in requests and limits. |
| securityContext | object | `{}` | Container security context. |
| service.annotations | object | `{}` | Service annotations, such as `networking.gke.io/load-balancer-type: Internal`. |
| service.externalTrafficPolicy | string | `"Local"` | `externalTrafficPolicy` of a LoadBalancer. `Local` keeps the caller's address, which the NetworkPolicy needs to admit it by range. |
| service.loadBalancerSourceRanges | list | `[]` | Source ranges a LoadBalancer admits, such as the calling cluster's pod range. The model server checks no token, so this is its access control. The NetworkPolicy admits the same ranges. |
| service.type | string | `"ClusterIP"` | `ClusterIP` for callers in this cluster. `LoadBalancer` for an inference service in another cluster, which then requires `service.loadBalancerSourceRanges`. |
| serviceAccountName | string | `nil` | Service account the pod runs as. |
| startupProbe.failureThreshold | int | `90` | Startup probe attempts, 10 s apart. Nothing listens until the preload is done, downloads included, so the default allows 15 minutes. |
| tolerations | list | `[]` | Tolerations, for tainted GPU nodes for example. |
| weights.persistence.enabled | bool | `false` | Keep downloaded weights on a PersistentVolumeClaim, so a restart does not download them again and the node's disk is not filled. Off uses an emptyDir. |
| weights.persistence.existingClaim | string | `nil` | Use this claim instead of creating one. |
| weights.persistence.size | string | `"30Gi"` | Size of the created claim. |
| weights.persistence.storageClass | string | `nil` | Storage class of the created claim. Empty is the cluster default. |
| weights.sizeLimit | string | `nil` | emptyDir size limit, when persistence is off. |

<!-- markdownlint-enable MD060 -->
