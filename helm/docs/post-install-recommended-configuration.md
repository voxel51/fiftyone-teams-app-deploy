# Recommended Post-Installation Configuration

After completing the base Helm installation, we strongly recommend enabling the
following features for a production-ready FiftyOne Enterprise deployment:

1. [Dedicated Plugins Mode](#dedicated-plugins-mode)
2. [Delegated Operators](#delegated-operators)
3. [GPU Workloads (Optional)](#gpu-workloads-optional)

The default installation uses **builtin-only plugins** and has **no delegated
operator workers**. While this is sufficient to get started, it limits the
ability to install custom plugins and run long-running or compute-heavy tasks
in the background.

> **Note:** Steps 1 and 2 require shared storage (a PersistentVolumeClaim).
> See [Prerequisites](#prerequisites-shared-storage) before proceeding.

---

## Prerequisites: Shared Storage

Dedicated plugins and delegated operators share a common plugin directory
backed by a Kubernetes PersistentVolume (PV) and PersistentVolumeClaim (PVC).

If you do not already have shared storage configured, refer to
[Adding Shared Storage for FiftyOne Enterprise Plugins](./plugins-storage.md)
for steps on creating a PV and PVC.

The examples below assume:

- PVC name: `plugins-pvc`
- Plugin directory: `/opt/plugins`

---

## Dedicated Plugins Mode

By default, plugins run inside the `fiftyone-app` pod (builtin-only mode).
We recommend enabling **dedicated plugins mode**, which runs plugins in their
own `teams-plugins` pod. This provides:

- **Custom plugin support** — install plugins from the
  [FiftyOne Plugin Library](https://github.com/voxel51/fiftyone-plugins) or
  build your own
- **Resource isolation** — plugin workloads do not affect `fiftyone-app`
  stability or performance
- **Custom dependency support** — plugins with heavy ML dependencies (e.g.
  `torch`, `transformers`) are isolated from the main app

There are three plugin modes available:

| Mode | Description |
| --- | --- |
| Builtin Only (default) | Only builtin plugins shipped with FiftyOne Enterprise |
| Shared | Custom plugins run inside `fiftyone-app` — may starve the app |
| **Dedicated (recommended)** | Custom plugins run in a dedicated `teams-plugins` pod |

To enable dedicated plugins, add the following to your `values.yaml`:

```yaml
# Dedicated plugins pod
pluginsSettings:
  enabled: true
  env:
    FIFTYONE_PLUGINS_DIR: /opt/plugins

# teams-api also requires access to the plugins directory
apiSettings:
  env:
    FIFTYONE_PLUGINS_DIR: /opt/plugins
```

Then apply the changes:

```bash
helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
  -f values.yaml \
  -n <your-namespace>
```

Verify the `teams-plugins` pod is running:

```bash
kubectl get pods -n <your-namespace> | grep teams-plugins
```

For more details, see
[Configuring Plugins](./configuring-plugins.md).

---

## Delegated Operators

Delegated operators allow long-running or compute-heavy tasks — such as
computing embeddings, running model evaluations, importing datasets, or
annotation workflows — to be scheduled from the FiftyOne UI and executed in
the background on dedicated compute workers.

> **Note:** If you are using builtin-only plugin mode, omit the PVC volume
> mount from the configuration below. The `teamsDo` pod will only be able to
> execute builtin operators.

To enable delegated operators with dedicated plugins, add the following to
your `values.yaml`:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo:
      replicaCount: 3  # adjust based on your concurrency needs
      env:
        FIFTYONE_PLUGINS_DIR: /opt/plugins
      volumes:
        - name: plugins-vol
          persistentVolumeClaim:
            claimName: plugins-pvc
            readOnly: true
      volumeMounts:
        - name: plugins-vol
          mountPath: /opt/plugins
```

Then apply the changes:

```bash
helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
  -f values.yaml \
  -n <your-namespace>
```

Verify the `teams-do` pod is running:

```bash
kubectl get pods -n <your-namespace> | grep teams-do
```

For full configuration options, see
[Configuring Delegated Operators](./configuring-delegated-operators.md).

### On-Demand Delegated Operators

As an alternative to always-on workers, FiftyOne Enterprise v2.11.0+ supports
**on-demand delegated operators** that spin up compute pods only when a job is
scheduled, and tear them down when complete. This is more cost-efficient for
infrequent or GPU-intensive workloads.

See
[Configuring On-Demand Orchestrator](../../docs/configuring-on-demand-orchestrator.md)
for setup instructions.

---

## GPU Workloads (Optional)

If your team runs GPU-accelerated tasks (e.g. computing embeddings with
`@voxel51/brain`, model evaluation, or custom ML operators), you can schedule
`teams-do` pods on GPU-enabled nodes using `nodeSelector` and `tolerations`.

Add the following to your `delegatedOperatorDeployments.deployments.teamsDo`
config in `values.yaml`:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo:
      nodeSelector:
        <your-gpu-node-label-key>: <your-gpu-node-label-value>
        # e.g. node-type: gpu
      tolerations:
        - key: "nvidia.com/gpu"
          operator: "Exists"
          effect: "NoSchedule"
      resources:
        limits:
          nvidia.com/gpu: 1
        requests:
          nvidia.com/gpu: 1
```

For full details, see
[Configuring GPU Workloads](./configuring-gpu-workloads.md).

### Multiple Orchestrators

For deployments with mixed workloads, you can register multiple delegated
operator orchestrators — for example, one targeting GPU nodes and one
targeting CPU nodes — and route specific operators to the appropriate
orchestrator from the FiftyOne UI.

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDoGpu:
      replicaCount: 1
      env:
        FIFTYONE_PLUGINS_DIR: /opt/plugins
      nodeSelector:
        node-type: gpu
      tolerations:
        - key: "nvidia.com/gpu"
          operator: "Exists"
          effect: "NoSchedule"
      resources:
        limits:
          nvidia.com/gpu: 1
      volumes:
        - name: plugins-vol
          persistentVolumeClaim:
            claimName: plugins-pvc
            readOnly: true
      volumeMounts:
        - name: plugins-vol
          mountPath: /opt/plugins

    teamsDocpu:
      replicaCount: 2
      env:
        FIFTYONE_PLUGINS_DIR: /opt/plugins
      volumes:
        - name: plugins-vol
          persistentVolumeClaim:
            claimName: plugins-pvc
            readOnly: true
      volumeMounts:
        - name: plugins-vol
          mountPath: /opt/plugins
```

---

## Advanced: Custom Job Priorities

> **Note:** The Helm chart's `delegatedOperatorJobTemplates.jobs` does not
> currently support `priorityClassName` natively. To use Kubernetes
> [PriorityClasses](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)
> with delegated operator jobs — for example, to prevent DO workloads from
> contending with user-facing pods — define custom Jinja2 job templates via a
> ConfigMap and register them using the
> [FiftyOne Management SDK](https://docs.voxel51.com/enterprise/management_sdk.html).

See
[Configuring On-Demand Orchestrator](../../docs/configuring-on-demand-orchestrator.md)
for details on custom job templates.

---

## Verifying Your Setup

After applying all changes, verify that all expected pods are running:

```bash
kubectl get pods -n <your-namespace>
```

You should see:

- `teams-plugins-*` — dedicated plugins pod
- `teams-do-*` (one or more) — delegated operator worker pod(s)

You can also verify delegated operator registration from the FiftyOne Python
SDK:

```python
import fiftyone.operators.orchestrator as foo

orc_svc = foo.OrchestratorService()
for orc in orc_svc.list():
    print("{} \"{}\" {}".format(orc.instance_id, orc.description, orc.id))
```

---

## Getting Started with Plugins

Once dedicated plugins are enabled, you can install community plugins from the
[FiftyOne Plugin Library](https://github.com/voxel51/fiftyone-plugins) or
the [FiftyOne Docs](https://docs.voxel51.com/plugins/index.html).

Some recommended plugins to get started:

- [`@voxel51/brain`](https://github.com/voxel51/fiftyone-plugins/tree/main/plugins/brain)
  — compute embeddings and similarity indexes
- [`@voxel51/annotation`](https://github.com/voxel51/fiftyone-plugins/tree/main/plugins/annotation)
  — annotation workflows
- [`@voxel51/evaluation`](https://github.com/voxel51/fiftyone-plugins/tree/main/plugins/evaluation)
  — model evaluation panels
- [`@voxel51/zoo`](https://github.com/voxel51/fiftyone-plugins/tree/main/plugins/zoo)
  — access the FiftyOne Model Zoo

To use plugins with custom dependencies (e.g. `torch`, `transformers`), build
and use
[Custom Plugin Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).
