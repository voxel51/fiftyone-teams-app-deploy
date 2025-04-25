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

- [v2.7.0+](#v270)
- [Using `delegatedOperatorDeployments`](#using-delegatedoperatordeployments)
  - [Built-in Plugins](#built-in-plugins)
  - [Shared/Dedicated Plugins](#shareddedicated-plugins)
  - [Examples](#examples)
    - [Map Merges](#map-merges)
    - [List Merges](#list-merges)
- [Migrating from `delegatedOperatorExecutorSettings` to `delegatedOperatorDeployments`](#migrating-from-delegatedoperatorexecutorsettings-to-delegatedoperatordeployments)
  - [Example](#example)
- [Prior to v2.7.0](#prior-to-v270)

<!-- tocstop -->

## v2.7.0+

As of version 2.7.0, `delegatedOperatorExecutorSettings`
has been deprecated in favor of `delegatedOperatorDeployments`.
`delegatedOperatorExecutorSettings` has been marked for deletion
for versions released after May 31st, 2025.

`delegatedOperatorDeployments` enables you to deploy multiple instances
of delegated operators targeting different hardware or use-cases.

## Using `delegatedOperatorDeployments`

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

To enable delegated operators, add an object to `delegatedOperatorDeployments.deployments`:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}
```

The Kubernetes deployment's name will be generated from `deployments` key-name
converted to kebab-case.
In the above example (key named `teamsDo`),
the resulting deployment name would be `teams-do`.

Delegated operators can be added to any of the three existing
[plugin modes](./confuring-plugins.md).

### Built-in Plugins

For built-in plugins, no additional configuration
is needed.

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}
```

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

### Examples

See below for examples on how merging templates are applied.

#### Map Merges

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

#### List Merges

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
[plugin modes](./confuring-plugins.md).
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
