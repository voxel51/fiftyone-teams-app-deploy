<!-- markdownlint-disable no-inline-html line-length no-alt-text -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length no-alt-text -->

---

# Configuring FiftyOne Teams Delegated Operators

<!-- toc -->

- [v2.7.0+](#v270)
- [Using `delegatedOperatorDeployments`](#using-delegatedoperatordeployments)
  - [Built-in Plugins](#built-in-plugins)
  - [Shared/Dedicated Plugins](#shareddedicated-plugins)
  - [Examples](#examples)
    - [Map Merges](#map-merges)
    - [List Merges](#list-merges)
- [Migrating from `delegatedOperatorExecutorSettings` to `delegatedOperatorDeployments`](#migrating-from-delegatedoperatorexecutorsettings-to-delegatedoperatordeployments)
- [Prior to v2.7.0](#prior-to-v270)

<!-- tocstop -->

## v2.7.0+

As of version 2.7.0, `delegatedOperatorExecutorSettings`
has been deprecated in favor of `delegatedOperatorDeployments`.
This additional value allows users to deploy multiple instances
of delegated operators, targeting different hardware or different
use-cases.

## Using `delegatedOperatorDeployments`

`delegatedOperatorDeployments` allows you to apply a basic
template to one or multiple delegated operator deployments.
Values in `delegatedOperatorDeployments.template` will be applied to
each deployment instance under `delegatedOperatorDeployments.deployments`.

For settings applied in both places, the following rules apply:

1. If the setting is a `map` object, the
   `delegatedOperatorDeployments.template.SETTING` will be merged, key-wise,
   with `delegatedOperatorDeployments.deployments.DEPLOYMENT.SETTING`.
   For settings that are multiply defined, the
   `delegatedOperatorDeployments.deployments.DEPLOYMENT.SETTING` will
   take precedence.

1. If the setting is a `list` object, the
   `delegatedOperatorDeployments.template.SETTING` will be ignored and the
   `delegatedOperatorDeployments.deployments.DEPLOYMENT.SETTING` will
   take precedence.

See [examples](#examples) for more information.

To enable delegated operators, add an object to `delegatedOperatorDeployments.deployments`:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}
```

The Kubernetes name will be generated from the key-name in the `deployments`
map by applying kebab-case to it.
In the above example, the resulting kubernetes object would be named `teams-do`.

Delegated operators can be added to any of the three existing
[plugin modes](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/confuring-plugins.md).

### Built-in Plugins

For built-in plugins, no additional configuration
is needed.

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}
```

### Shared/Dedicated Plugins

For shared/dedicated plugins, you'll need to mount
the plugins `PersistentVolumeClaim` with `Read` access.
It should be mounted at `FIFTYONE_PLUGINS_DIR`.

This can be done by modifying `values.yaml` in one
of the two ways below:

1. The template (applies to all instances):

```yaml
delegatedOperatorDeployments:
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
  deployments:
    teamsDo: {}
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
[Adding Shared Storage for FiftyOne Teams Plugins](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md)
for configuring persistent volumes and claims.

Optionally, the logs generated during running of a delegated operation can be
uploaded to a network-mounted file system or cloud storage path that is
available to this deployment.
Logs are uploaded in the format
`<configured_path>/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log`

In `values.yaml`, set `FIFTYONE_DELEGATED_OPERATION_LOG_PATH` in either:

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
[Custom Plugins Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).

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

Would result in the following:

```yaml
resources:
  requests:
    cpu: 1
    memory: 1Gi
  limits:
    cpu: 6
    memory: 6Gi
```

Note that `requests` was merged key-wise and, therefore,
settings from both the template and instance are included.

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

Would result in the following:

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

1. Add a `teamsDo` object to `delegatedOperatorDeployments.deployments`.
1. Copy any configuration details from `delegatedOperatorExecutorSettings` to
   either `delegatedOperatorDeployments.template` or
   `delegatedOperatorDeployments.deployments.teamsDo`.
      1. Be mindful of spacing as
         `delegatedOperatorDeployments.deployments.teamsDo` is more
         indented than its predecessor.
1. Set `delegatedOperatorExecutorSettings.enabled` to `false`

## Prior to v2.7.0

This option can be added to any of the three existing
[plugin modes](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/confuring-plugins.md).
If you're using the builtin-operator
only option, the Persistent Volume Claim should be omitted.

To enable this mode

- In `values.yaml`, set
  - `delegatedOperatorExecutorSettings.enabled: true`
  - The path for a Persistent Volume Claim mounted to the
    `teams-do` deployment in
    - `delegatedOperatorExecutorSettings.env.FIFTYONE_PLUGINS_DIR`
- See
  [Adding Shared Storage for FiftyOne Teams Plugins](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md)
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
[Custom Plugins Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).
