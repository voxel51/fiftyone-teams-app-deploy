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

# Configuring FiftyOne Enterprise Delegated Operators

<!-- toc -->

- [v2.14.0+](#v2140)
- [v2.7.0+](#v270)
- [Using `delegatedOperatorJobTemplates` for on-demand executors](#using-delegatedoperatorjobtemplates-for-on-demand-executors)
  - [Built-in Plugins](#built-in-plugins)
  - [Shared/Dedicated Plugins](#shareddedicated-plugins)
- [Using `delegatedOperatorDeployments` for always-on executors](#using-delegatedoperatordeployments-for-always-on-executors)
  - [Built-in Plugins](#built-in-plugins-1)
  - [Shared/Dedicated Plugins](#shareddedicated-plugins-1)
- [Examples](#examples)
  - [Map Merges](#map-merges)
  - [List Merges](#list-merges)
- [Migrating from `delegatedOperatorExecutorSettings` to `delegatedOperatorDeployments`](#migrating-from-delegatedoperatorexecutorsettings-to-delegatedoperatordeployments)
  - [Example](#example)
- [Prior to v2.7.0](#prior-to-v270)

<!-- tocstop -->

## v2.14.0+

> [!NOTE]
> `delegatedOperatorJobTemplates` and on-demand kubernetes orchestration
> are currently in beta and can be used by early adopters.

`delegatedOperatorJobTemplates` was added in version 2.14.0 which allows users
to create on-demand delegated operators utilizing
[kubernetes jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/).
`delegatedOperatorJobTemplates` enables you to create multiple job
templates that FiftyOne Enterprise can use to create Kubernetes jobs.

> [!NOTE]
> Using `delegatedOperatorJobTemplates` and on-demand kubernetes orchestration
> requires that you install the `kubernetes` python client into your
> `teams-api` image.
> See
> [the kubernetes orchestrator docs](../../docs/orchestrators/configuring-kubernetes-orchestrator.md)

FiftyOne Enterprise 2.14+ deploys the `teams-do-cpu-default`
delegated operator `Deployment` by default.
Configuring the delegated operator has
[not changed](#using-delegatedoperatordeployments-for-always-on-executors).
The `teams-do-cpu-default` deployment can be
disabled by setting
`delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled=false`
in the `values.yaml` file:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDoCpuDefault:
      enabled: false
```

## v2.7.0+

As of version 2.7.0, `delegatedOperatorExecutorSettings`
has been deprecated in favor of `delegatedOperatorDeployments`.
`delegatedOperatorExecutorSettings` has been marked for deletion
for versions released after May 31st, 2025.

`delegatedOperatorDeployments` enables you to deploy multiple instances
of delegated operators targeting different hardware or use-cases.

## Using `delegatedOperatorJobTemplates` for on-demand executors

The values in `delegatedOperatorJobTemplates.template` will be applied to
every job instance under `delegatedOperatorJobTemplates.jobs`.
Per job values set in
`delegatedOperatorJobTemplates.jobs.<JOB>`
will override the template.

For overlapping settings, the following rules apply:

1. When the setting is a `map` object, the
   `delegatedOperatorJobTemplates.template.<SETTING>` will be key-wise merged
   with `delegatedOperatorJobTemplates.jobs.<JOB>.<SETTING>`.
   For duplicate setting definitions, the
   `delegatedOperatorJobTemplates.jobs.<JOB>.<SETTING>` will
   take precedence.

1. When the setting is a `list` object, the
   `delegatedOperatorJobTemplates.template.<SETTING>` will be ignored and the
   `delegatedOperatorJobTemplates.jobs.<JOB>.<SETTING>` will
   take precedence.

See
[examples](#examples)
for more information.

To enable delegated operators, add an object to `delegatedOperatorJobTemplates.jobs`:

```yaml
delegatedOperatorJobTemplates:
  jobs:
    teamsDoCpuDefaultK8s: {}
```

The helm chart will create a `ConfigMap` with an `teamsDoCpuDefaultK8s.yaml` entry.
This entry will be mounted onto the API as a file
at `/tmp/do-targets/teamsDoCpuDefaultK8s.yaml`.
Delegated operators can be added to any of the three existing
[plugin modes](./configuring-plugins.md).

### Built-in Plugins

For built-in plugins, no additional configuration
is needed.

```yaml
delegatedOperatorJobTemplates:
  jobs:
    teamsDoCpuDefaultK8s: {}
```

### Shared/Dedicated Plugins

For shared/dedicated plugins, mount the plugins'
`PersistentVolumeClaim` with `Read` access
at `FIFTYONE_PLUGINS_DIR`.

This can be done by modifying `values.yaml` in one
of these two ways:

1. The template (applies to all instances):

    ```yaml
    delegatedOperatorJobTemplates:
      jobs:
        teamsDoCpuDefaultK8s: {}
      template:
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

1. Or, per instance:

    ```yaml
    delegatedOperatorJobTemplates:
      jobs:
        teamsDoCpuDefaultK8s:
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

See
[Adding Shared Storage for FiftyOne Enterprise Plugins](./plugins-storage.md)
for configuring persistent volumes and claims.

Optionally, the delegated operation run logs may be
uploaded to a network-mounted file system or cloud storage path
available to this deployment.
Logs are uploaded in the format
`<configured_path>/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log`

In `values.yaml`, set the environment variable
`FIFTYONE_DELEGATED_OPERATION_LOG_PATH` in either:

1. The template (applies to all instances):

    ```yaml
    delegatedOperatorJobTemplates:
      template:
        env:
          FIFTYONE_DELEGATED_OPERATION_LOG_PATH: /your/path/
    ```

1. Or, per instance:

    ```yaml
    delegatedOperatorJobTemplates:
      jobs:
        teamsDoCpuDefaultK8s:
          env:
            FIFTYONE_DELEGATED_OPERATION_LOG_PATH: /your/path
    ```

To use plugins with custom dependencies, build and use
[Custom Plugins Images](../../docs/custom-plugins.md).

## Using `delegatedOperatorDeployments` for always-on executors

The values in `delegatedOperatorDeployments.template` will be applied to
every deployment instance under `delegatedOperatorDeployments.deployments`.
Per deployment values set in
`delegatedOperatorDeployments.deployments.<DEPLOYMENT>`
will override the template.

For overlapping settings, the following rules apply:

1. When the setting is a `map` object, the
   `delegatedOperatorDeployments.template.<SETTING>` will be key-wise merged
   with `delegatedOperatorDeployments.deployments.<DEPLOYMENT>.<SETTING>`.
   For duplicate setting definitions, the
   `delegatedOperatorDeployments.deployments.<DEPLOYMENT>.<SETTING>` will
   take precedence.

1. When the setting is a `list` object, the
   `delegatedOperatorDeployments.template.<SETTING>` will be ignored and the
   `delegatedOperatorDeployments.deployments.<DEPLOYMENT>.<SETTING>` will
   take precedence.

See
[examples](#examples)
for more information.

As of v2.14.0, delegated operators are enabled by default, with a deployment
called `teamsDoCpuDefault` ("teams-do-cpu-default"). Please see
[values.yaml](../fiftyone-teams-app/values.yaml)
for the details of this always-on executor, including the default resource
`requests`.

To disable this deployment, you must create an entry `teamsDoCpuDefault` key and
set `enabled: false`. Example:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDoCpuDefault:
      enabled: false
```

> [!NOTE] By default, the `teamsDoCpuDefault` delegated operator deployment will
> be enabled alongside other delegated operator deployments defined here.

To enable non-default delegated operators, add an object to
`delegatedOperatorDeployments.deployments`:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}
    teamsDoCpuDefault:  # Optional. These lines can be omitted if keeping the
      enabled: false    # default deployment is desired.
```

The Kubernetes deployment's name will be generated from `deployments` key-name
converted to kebab-case.
In the above example (key named `teamsDo`),
the resulting deployment name would be `teams-do`.

Delegated operators can be added to any of the three existing
[plugin modes](./configuring-plugins.md).

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Built-in Plugins

For built-in plugins, no additional configuration
is needed.

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}
```

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Shared/Dedicated Plugins

For shared/dedicated plugins, mount the plugins'
`PersistentVolumeClaim` with `Read` access
at `FIFTYONE_PLUGINS_DIR`.

This can be done by modifying `values.yaml` in one
of these two ways:

1. The template (applies to all instances):

    ```yaml
    delegatedOperatorDeployments:
      deployments:
        teamsDo: {}
      template:
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

> [!NOTE] As noted in
> [List Merges](#list-merges),
> lists (incl. `volumes` and `volumeMounts`) will *not* be inherited by
> deployment instances that define their own values. This includes the implicit
> `teamsDoCpuDefault` delegated operator deployment, which has its own required
> `volumes` and `volumeMounts`. As a result, Shared/Dedicated Plugins will not
> be available in these deployments using the above technique without additional
> configuration.

1. Or, per instance:

    ```yaml
    delegatedOperatorDeployments:
      deployments:
        teamsDo:
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

See
[Adding Shared Storage for FiftyOne Enterprise Plugins](./plugins-storage.md)
for configuring persistent volumes and claims.

Optionally, the logs generated during running of a delegated operation may be
uploaded to a network-mounted file system or cloud storage path
available to this deployment.
Logs are uploaded in the format
`<configured_path>/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log`

In `values.yaml`, set the environment variable
`FIFTYONE_DELEGATED_OPERATION_LOG_PATH` in either:

1. The template (applies to all instances):

    ```yaml
    delegatedOperatorDeployments:
      template:
        env:
          FIFTYONE_DELEGATED_OPERATION_LOG_PATH: /your/path/
    ```

1. Or, per instance:

    ```yaml
    delegatedOperatorDeployments:
      deployments:
        teamsDo:
          env:
            FIFTYONE_DELEGATED_OPERATION_LOG_PATH: /your/path
    ```

To use plugins with custom dependencies, build and use
[Custom Plugins Images](../../docs/custom-plugins.md).

## Examples

See below for examples on how merging templates are applied.

### Map Merges

The following values:

```yaml
delegatedOperatorDeployments:
  template:
    resources:
      limits:
        cpu: 2
        memory: 2Gi
      requests:
        cpu: 1
        memory: 1Gi
  deployments:
    teamsDo:
      resources:
        limits:
          cpu: 6
          memory: 6Gi
```

Would result in a Kubernetes deployment resource with:

```yaml
resources:
  requests:
    cpu: 1
    memory: 1Gi
  limits:
    cpu: 6
    memory: 6Gi
```

Note that `requests` was merged key-wise.
Therefore, settings from both the template and instance are included,
with the instance values taking precedent.

### List Merges

The following values:

```yaml
delegatedOperatorDeployments:
  template:
    tolerations:
      - key: "template-example-key"
        operator: "Exists"
        effect: "NoSchedule"
  deployments:
    teamsDo:
      tolerations:
        - key: "instance-example-key"
          operator: "Exists"
          effect: "PreferNoSchedule"
```

Would result in a Kubernetes deployment resource with:

```yaml
tolerations:
  - key: "instance-example-key"
    operator: "Exists"
    effect: "PreferNoSchedule"
```

Note that `delegatedOperatorDeployments.template.tolerations`
was overridden by `delegatedOperatorDeployments.deployments.teamsDo.tolerations`.

## Migrating from `delegatedOperatorExecutorSettings` to `delegatedOperatorDeployments`

To migrate from `delegatedOperatorExecutorSettings` to `delegatedOperatorDeployments`:

In your `values.yaml`

1. Set `delegatedOperatorExecutorSettings.enabled` to `false`
1. Add `delegatedOperatorDeployments.deployments.teamsDo`
1. Copy any configuration details from `delegatedOperatorExecutorSettings` to
   either `delegatedOperatorDeployments.template` or
   `delegatedOperatorDeployments.deployments.teamsDo`
      1. Note the spacing as
         `delegatedOperatorDeployments.deployments.teamsDo` is more
         indented than `delegatedOperatorExecutorSettings`

### Example

A version 2.6.0 values file contained:

```yaml
delegatedOperatorExecutorSettings:
  enabled: true
  env:
    FIFTYONE_PLUGINS_DIR: /opt/plugins
  image:
    repository: my-internal-repo/fiftyone-teams-cv-full
    tag: v2.6.0
  replicaCount: 3
  resources:
    limits:
      cpu: 2
      ephemeral-storage: 1Gi
      memory: 6Gi
    requests:
      cpu: 2
      ephemeral-storage: 1Gi
      memory: 6Gi
  securityContext:
    readOnlyRootFilesystem: true
  volumeMounts:
    - name: opt-fiftyone
      mountPath: /opt/fiftyone
    - name: tmpdir
      mountPath: /tmp
  volumes:
    - name: nfs-plugins-ro-vol
      persistentVolumeClaim:
        claimName: release-rc-plugins-pvc
        readOnly: true
    - name: tmpdir
      emptyDir: {}
```

A migrated `values.yaml` file would contain:

```yaml
delegatedOperatorExecutorSettings:
  enabled: false

delegatedOperatorDeployments:
  deployments:
    teamsDo:
      env:
        FIFTYONE_PLUGINS_DIR: /opt/plugins
      image:
        repository: my-internal-repo/fiftyone-teams-cv-full
        tag: v2.6.0
      replicaCount: 3
      resources:
        limits:
          cpu: 2
          ephemeral-storage: 1Gi
          memory: 6Gi
        requests:
          cpu: 2
          ephemeral-storage: 1Gi
          memory: 6Gi
      securityContext:
        readOnlyRootFilesystem: true
      volumeMounts:
        - name: opt-fiftyone
          mountPath: /opt/fiftyone
        - name: tmpdir
          mountPath: /tmp
      volumes:
        - name: nfs-plugins-ro-vol
          persistentVolumeClaim:
            claimName: release-rc-plugins-pvc
            readOnly: true
        - name: tmpdir
          emptyDir: {}
```

After this migration, you will have two delegated operators on FiftyOne:
`builtin` and `teams-do`.
`builtin` currently has no resources allocated to it and can be safely
dropped via:

```python
# Import the FiftyOne Operators Orchestrator
import fiftyone.operators.orchestrator as foo
orc_svc = foo.OrchestratorService()

# List the current operators
for orc in orc_svc.list():
    print("{} \"{}\" {}".format(orc.instance_id, orc.description, orc.id))

# Delete the builtin operator
orc_svc.delete(id='builtin')

# Verify there is no longer a `builtin`
for orc in orc_svc.list():
    print("{} \"{}\" {}".format(orc.instance_id, orc.description, orc.id))
```

## Prior to v2.7.0

This option can be added to any of the three existing
[plugin modes](./configuring-plugins.md).
If you're using the builtin-operator
only option, the Persistent Volume Claim should be omitted.

To enable this mode

- In `values.yaml`, set
  - `delegatedOperatorExecutorSettings.enabled: true`
  - The path for a Persistent Volume Claim mounted to the
    `teams-do` deployment in
    - `delegatedOperatorExecutorSettings.env.FIFTYONE_PLUGINS_DIR`
- See
  [Adding Shared Storage for FiftyOne Enterprise Plugins](./plugins-storage.md)
  - Mount a Persistent Volume Claim (PVC) that provides
    - `ReadWrite` permissions to the `teams-do` deployment
      at the `FIFTYONE_PLUGINS_DIR` path

Optionally, the logs generated during running of a delegated operation can be
uploaded to a network-mounted file system or cloud storage path that is
available to this deployment. Logs are uploaded in the format
`<configured_path>/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log`
In `values.yaml`, set `configured_path`

- `delegatedOperatorExecutorSettings.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH`

To use plugins with custom dependencies, build and use
[Custom Plugins Images](../../docs/custom-plugins.md).
