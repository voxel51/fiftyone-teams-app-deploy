<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img alt="Voxel51 Logo" src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img alt="Voxel51 FiftyOne" src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# Configuring the Agentic Labeler Service

<!-- toc -->

- [Overview](#overview)
- [Model downloads](#model-downloads)
  - [Applying changes](#applying-changes)
  - [Network access](#network-access)
  - [Hugging Face token (recommended)](#hugging-face-token-recommended)
  - [Keep the model between restarts (recommended)](#keep-the-model-between-restarts-recommended)
  - [Offline and air-gapped clusters](#offline-and-air-gapped-clusters)
- [Troubleshooting](#troubleshooting)

<!-- tocstop -->

## Overview

The Agentic Labeler service labels images with a vision-language model
running on a GPU.
The chart already defines it on the `gpuServiceOrc`
[service orchestrator](../../docs/configuring-service-orchestrator.md),
so there is nothing to install.
To start it, go to *Settings* -> *Services* in the FiftyOne Enterprise UI.

For GPU node setup and sizing, see
[Configuring Service Orchestrators](../../docs/configuring-service-orchestrator.md#kubernetes)
and
[Leveraging GPU Workloads](./configuring-gpu-workloads.md).

## Model downloads

The service does not ship with its model.
Each time the service starts, it downloads the model
(about 24 GB) from [Hugging Face](https://huggingface.co),
unless it is already stored on a volume
(see [Keep the model between restarts](#keep-the-model-between-restarts-recommended)).
Plan for the first start to take several minutes.

The default models are public.
You do not need a Hugging Face account to download them.

### Applying changes

The settings below go in your `values.yaml`,
under `delegatedOperatorJobTemplates.serviceOrchestrators.gpuServiceOrc`.
They also apply to the `annotation-ai` service, which runs on the same
orchestrator. This is safe.
After changing them:

1. Run `helm upgrade`.
1. In the FiftyOne Enterprise UI, go to *Settings* -> *Services*,
   then stop and start `agentic-labeler`.

### Network access

Service pods need outbound HTTPS access to
`huggingface.co` and `*.hf.co`.
Allow this traffic in any `NetworkPolicy`, firewall, or egress gateway
for your release namespace.

If your network uses a proxy, set it on `gpuServiceOrc`.
`NO_PROXY` must include `teams-api` so the service can still reach the
FiftyOne API inside the cluster:

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      env:
        HTTPS_PROXY: http://proxy.example.com:3128  # Modify as needed
        https_proxy: http://proxy.example.com:3128  # Modify as needed
        NO_PROXY: teams-api,localhost,127.0.0.1,.svc,.cluster.local
        no_proxy: teams-api,localhost,127.0.0.1,.svc,.cluster.local
```

If the cluster cannot reach Hugging Face at all, see
[Offline and air-gapped clusters](#offline-and-air-gapped-clusters).

### Hugging Face token (recommended)

A Hugging Face token is not required, but we recommend one.
Without it, downloads can be slow enough for the service to time out
on start.

1. Create a read-only
   [Hugging Face access token](https://huggingface.co/docs/hub/security-tokens).
1. Store it in a Kubernetes `Secret` in your release namespace
   (the `namespace.name` in your `values.yaml`):

    ```shell
    kubectl create secret generic huggingface-token \
      --namespace <namespace.name> \
      --from-literal=token=<your-token>
    ```

1. Reference it in your `values.yaml`, then
   [apply the changes](#applying-changes):

    ```yaml
    delegatedOperatorJobTemplates:
      serviceOrchestrators:
        gpuServiceOrc:
          secretEnv:
            HF_TOKEN:
              secretName: huggingface-token
              secretKey: token
    ```

Do not put the token in
`delegatedOperatorJobTemplates.serviceOrchestrators.gpuServiceOrc.services.agentic-labeler.env`.
Values there are not treated as secrets.

### Keep the model between restarts (recommended)

Without a volume, the model is downloaded on every start and stored
on the node's local disk.
To download it only once, create a `PersistentVolumeClaim`
in your release namespace:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: agentic-labeler-hf-cache
spec:
  accessModes:
    - ReadWriteOnce  # See the note below
  resources:
    requests:
      storage: 50Gi
```

Mount it on `gpuServiceOrc`, then [apply the changes](#applying-changes):

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      podSecurityContext:
        fsGroupChangePolicy: OnRootMismatch
      volumes:
        - name: hf-cache
          persistentVolumeClaim:
            claimName: agentic-labeler-hf-cache
      volumeMounts:
        - name: hf-cache
          mountPath: /home/voxel51/.cache/huggingface
```

The `annotation-ai` service mounts the same volume.
If `agentic-labeler` and `annotation-ai` may run at the same time on
different nodes, use `ReadWriteMany` and a storage class that supports it.

### Offline and air-gapped clusters

If the cluster cannot reach Hugging Face, pre-load the model onto the
volume from
[Keep the model between restarts](#keep-the-model-between-restarts-recommended),
then turn on offline mode.
The cluster also needs the `voxel51/agentic-labeler` image.
Mirror it to your registry the same way as your other FiftyOne Enterprise
images.

1. Save this pod as `model-download.yaml`.
   Replace `<version>` with your FiftyOne Enterprise release
   (for example, `v2.25.1`), and update `image` and `imagePullSecrets`
   if you use a mirror registry.
   The pod needs temporary access to Hugging Face,
   or to an internal Hugging Face mirror set with `HF_ENDPOINT`.

    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: agentic-labeler-model-download
    spec:
      restartPolicy: Never
      imagePullSecrets:
        - name: regcred
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: download
          image: voxel51/agentic-labeler:<version>
          command:
            - python3
            - -c
            - |
              from huggingface_hub import snapshot_download as d
              d("google/gemma-4-31B-it-qat-w4a16-ct")
              d("google/gemma-4-31B-it-assistant")
          env:
            - name: HF_TOKEN
              valueFrom:
                secretKeyRef:
                  name: huggingface-token
                  key: token
                  optional: true
            # - name: HF_ENDPOINT
            #   value: https://hf-mirror.example.com  # Your mirror
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          volumeMounts:
            - name: hf-cache
              mountPath: /home/voxel51/.cache/huggingface
      volumes:
        - name: hf-cache
          persistentVolumeClaim:
            claimName: agentic-labeler-hf-cache
    ```

    The second model is a small helper model that speeds up labeling.

1. Run it in your release namespace and wait for it to finish:

    ```shell
    kubectl apply -n <namespace.name> -f model-download.yaml
    kubectl wait -n <namespace.name> --for=jsonpath='{.status.phase}'=Succeeded \
      pod/agentic-labeler-model-download --timeout=1h
    kubectl delete -n <namespace.name> pod/agentic-labeler-model-download
    ```

1. Turn on offline mode, then [apply the changes](#applying-changes):

    ```yaml
    delegatedOperatorJobTemplates:
      serviceOrchestrators:
        gpuServiceOrc:
          env:
            HF_HUB_OFFLINE: "1"
    ```

If you changed `LABELER_CONFIG_FILE` to a different model,
download the models named by `model_id` and `speculative_config.model`
in that config file instead.

## Troubleshooting

| Error or symptom | What to do |
| --- | --- |
| `Deployment planner timed out after 120s` | The service could not reach Hugging Face quickly enough. Check network and proxy access to `huggingface.co`, and set `HF_TOKEN`. If downloads work but are slow, raise the limit as shown below this table. |
| `Could not resolve the weights of target model` | The service cannot reach Hugging Face and the model is not downloaded. Check network and proxy access, or follow [Offline and air-gapped clusters](#offline-and-air-gapped-clusters). |
| The service pod is stopped before it becomes ready | The first start, including the download, took longer than `FIFTYONE_SERVICE_POD_READY_TIMEOUT_S` (see [Broker settings](../../docs/configuring-service-orchestrator.md#broker-settings)). Keep the model on a volume, or raise the timeout. |
| Every start is as slow as the first | The model is not being kept between restarts. Add the [volume](#keep-the-model-between-restarts-recommended). |
| The pod is evicted for low ephemeral storage | Without a volume, the model is stored on the node's local disk. Add the [volume](#keep-the-model-between-restarts-recommended). |

To raise the startup time limit (in seconds), add this to your
`values.yaml`, then [apply the changes](#applying-changes):

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      env:
        LABELER_LAUNCH_PLAN_TIMEOUT_S: "300"
```
