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

# Upgrading FiftyOne Teams

<!-- toc -->

- [Upgrading From Previous Versions](#upgrading-from-previous-versions)
  - [A Note On Database Migrations](#a-note-on-database-migrations)
  - [From FiftyOne Enterprise Version 2.0.0 or Higher](#from-fiftyone-enterprise-version-200-or-higher)
    - [FiftyOne Enterprise v2.9+ Startup Probe Changes](#fiftyone-enterprise-v29-startup-probe-changes)
    - [FiftyOne Enterprise v2.9+ Delegated Operator Changes](#fiftyone-enterprise-v29-delegated-operator-changes)
    - [FiftyOne Enterprise v2.8+ `initContainer` Changes](#fiftyone-enterprise-v28-initcontainer-changes)
    - [FiftyOne Enterprise v2.7+ Delegated Operator Changes](#fiftyone-enterprise-v27-delegated-operator-changes)
    - [FiftyOne Enterprise v2.5+ Delegated Operator Changes](#fiftyone-enterprise-v25-delegated-operator-changes)
    - [FiftyOne Enterprise v2.2+ Delegated Operator Changes](#fiftyone-enterprise-v22-delegated-operator-changes)
      - [Delegated Operation Capacity](#delegated-operation-capacity)
      - [Existing Orchestrators](#existing-orchestrators)
    - [Version 2.2+ InitContainers Additions](#version-22-initcontainers-additions)
  - [From FiftyOne Enterprise Versions 1.6.0 to 1.7.1](#from-fiftyone-enterprise-versions-160-to-171)
  - [From FiftyOne Enterprise Versions After 1.1.0 and Before Version 1.6.0](#from-fiftyone-enterprise-versions-after-110-and-before-version-160)
  - [From Before FiftyOne Enterprise Version 1.1.0](#from-before-fiftyone-enterprise-version-110)
  - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)

<!-- tocstop -->

## Upgrading From Previous Versions

Voxel51 assumes you use the published Helm Chart to deploy your FiftyOne Enterprise
environment.
If you are using a custom deployment mechanism, carefully review the changes in
the
[Helm Chart](https://github.com/voxel51/fiftyone-teams-app-deploy)
and update your deployment accordingly.

Voxel51 provides a
[minimum example `values.yaml`](../values.yaml).

1. Edit the `values.yaml` file
1. To upgrade an existing helm installation

    1. Make sure you have followed the appropriate directions for
       [Upgrading From Previous Versions](#upgrading-from-previous-versions)

    1. Update your kubectl configuration to set your current namespace for
       your kubectl context

        ```shell
        kubectl config set-context --current --namespace your-namespace-here
        ```

    1. Update your Voxel51 Helm repository and upgrade your FiftyOne Enterprise
       deployment

        ```shell
        helm repo update voxel51
        helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
          -f ./values.yaml
        ```

    > **NOTE**: To view the changes Helm would apply during installations
    > and upgrades, consider using
    > [helm diff](https://github.com/databus23/helm-diff).
    > Voxel51 is not affiliated with the author of this plugin.
    >
    > For example:
    >
    > ```shell
    > helm diff -C1 upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
    > ```

### A Note On Database Migrations

The environment variable `FIFTYONE_DATABASE_ADMIN`
controls whether the database may be migrated.
This is a safety check to prevent automatic database
upgrades that will break other users' SDK connections.
When false (or unset), either an error will occur

```shell
$ fiftyone migrate --all
Traceback (most recent call last):
...
OSError: Cannot migrate database from v0.22.0 to v0.22.3 when database_admin=False.
```

or no action will be taken:

```shell
$ fiftyone migrate --info
FiftyOne Enterprise version: 0.14.4
FiftyOne compatibility version: 0.22.3
Other compatible versions: >=0.19,<0.23
Database version: 0.21.2
dataset     version
----------  ---------
quickstart  0.22.0
$ fiftyone migrate --all
$ fiftyone migrate --info
FiftyOne Enterprise version: 0.14.4
FiftyOne compatibility version: 0.23.0
Other compatible versions: >=0.19,<0.23
Database version: 0.21.2
dataset     version
----------  ---------
quickstart  0.21.2
```

### From FiftyOne Enterprise Version 2.0.0 or Higher

1. [Upgrade to FiftyOne Enterprise version 2.14.0](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Enterprise SDK users to FiftyOne Enterprise
   version 2.14.0
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`

1. Voxel51 recommends that you upgrade all your datasets, but it is not
   required.

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at version 1.8.0, run

   ```shell
   fiftyone migrate --info
   ```

#### FiftyOne Enterprise v2.9+ Startup Probe Changes

<!-- Differs from docker-compose docs -->

In `v2.9.0`, the `values.yaml` location of startup probe settings
moved from `<settings>.service.startup` to `<settings>.startup` section.

If your `values.yaml` contains

```yaml
appSettings:
  service:
    startup:
      failureThreshold: 10
      periodSeconds: 15
```

move `appSettings.service.startup` to `appSettings.startup`

```yaml
appSettings:
  startup:
    failureThreshold: 10
    periodSeconds: 15
```

Repeat for each occurrence of `<settings>.service.startup` to `<settings>.startup`.

#### FiftyOne Enterprise v2.9+ Delegated Operator Changes

<!-- Differs from docker-compose docs -->

FiftyOne Enterprise v2.7.0 deprecated the
`delegatedOperatorExecutorSettings` setting in `values.yaml`.
This has been removed in v2.9.0.

Please refer to
[the delegated operator documentation](./configuring-delegated-operators.md#migrating-from-delegatedoperatorexecutorsettings-to-delegatedoperatordeployments)
for migrating to the new setting.

#### FiftyOne Enterprise v2.8+ `initContainer` Changes

FiftyOne Enterprise v2.8.2 introduces numerous changes to the default settings
for each component's `initContainers`.

1. `initContainers` default to the container security context
   shown below in order to comply with kubernetes security best practices.
   This configuration prevents privilege escalation and running any
   initialization processes as root.

    ```yaml
      containerSecurityContext:
        allowPrivilegeEscalation: false  # Disables privilege escalation
        runAsNonRoot: true  # Disables running as the `root` user
        runAsUser: 1000  # Runs the init processes as UID 1000
    ```

1. `initContainers` default to the resources shown below.
   These initialization processes are lightweight
   and can therefore set small resource requests and limits instead of using
   a cluster's defaults.

   ```yaml
      resources:
        limits:
          cpu: 10m
          memory: 128Mi
        requests:
          cpu: 10m
          memory: 128Mi
    ```

#### FiftyOne Enterprise v2.7+ Delegated Operator Changes

FiftyOne Enterprise v2.7.0 introduces numerous changes to delegated operators.

1. The `FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH`
   environment variable has been changed to to
   `FIFTYONE_DELEGATED_OPERATION_LOG_PATH`.
   Please note that this change is backwards compatible, but should
   be changed in your manifests moving forward.

1. The `delegatedOperatorExecutorSettings` setting in `values.yaml` has
   been deprecated in favor of `delegatedOperatorDeployments`.
   Please refer to
   [the delegated operator documentation](./configuring-delegated-operators.md#v270)
   for migrating to the new setting.

#### FiftyOne Enterprise v2.5+ Delegated Operator Changes

FiftyOne Enterprise v2.5.0 changes the base image of the built-in delegated
operators (`teams-do`) from `voxel51/fiftyone-app` to `voxel51/fiftyone-teams-cv-full`.
The `voxel51/fiftyone-teams-cv-full` image includes all of the dependencies
required to run complex workflows out of the box.

If you built your own image with custom dependencies,
you will likely want to remake those images based off
of this new `voxel51/fiftyone-teams-cv-full` image.

Please note: this image is approximately 2GB larger than its predecessor
and, as such, might take longer to pull and start.

To utilize the prior image, update your `values.yaml` similar to the below:

```yaml
delegatedOperatorExecutorSettings:
  image:
    repository: voxel51/fiftyone-app
```

#### FiftyOne Enterprise v2.2+ Delegated Operator Changes

FiftyOne Enterprise v2.2 introduces some changes to delegated operators, detailed
below.

##### Delegated Operation Capacity

By default, all deployments are provisioned with capacity to support up to three
delegated operations simultaneously. You will need to configure the
[builtin orchestrator](https://helm.fiftyone.ai/#builtin-delegated-operator-orchestrator)
or an external orchestrator, with enough workers, to be able to utilize this
full capacity.
If your team finds the usage is greater than this, please reach out to your
Voxel51 support team for guidance and to increase this limit!

##### Existing Orchestrators

> [!NOTE]
> If you are currently utilizing an external orchestrator for delegated
> operations, such as Airflow or Flyte, you may have an outdated execution
> definition that could negatively affect the experience. Please reach out to
> Voxel51 support team for guidance on updating this code.

Additionally,

> [!WARNING]
> If you cannot update the orchestrator DAG/workflow code, you must set
> `delegatedOperatorExecutorSettings.env.FIFTYONE_ALLOW_LEGACY_ORCHESTRATORS: true`
> in `values.yaml` in order for the delegated operation system to function
> properly.

#### Version 2.2+ InitContainers Additions

Kubernetes [`initContainers`][init-containers]
were added in Version 2.2.0 to enforce the order of pod startup.
The image and tag are customizable.
Any image supporting `wget`, such as `docker.io/busybox`, are applicable
replacements.

For a full list of settings, please refer to the
[values list](https://helm.fiftyone.ai/#values).

> [!NOTE]
> It is recommended to add a `podSecurityContext` to avoid running
> init containers as root.
> An example policy is shown below:

```yaml
  podSecurityContext:
    runAsUser: 1000
```

> [!NOTE]
> Init containers can be disabled.
> Voxel51 does not recommend disabling init containers to enforce
> inter-pod dependencies are satisfied before proceeding.

### From FiftyOne Enterprise Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Enterprise v2.14.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets. You may remove those secrets
> from your `values.yaml` and from any secrets created outside of the Voxel51
> install process.

---

> **NOTE**: If you had previously set
> `teamsAppSettings.env.FIFTYONE_APP_INSTALL_FIFTYONE_OVERRIDE` to include your
> Voxel51 private PyPI token, remove it from your configuration. The
> Voxel51 private PyPI token is loaded from your license file.

---

1. Ensure all FiftyOne SDK users either
    - Set the `FIFTYONE_DATABASE_ADMIN` to `false`

      ```shell
      FIFTYONE_DATABASE_ADMIN=false
      ```

    - Unset the environment variable `FIFTYONE_DATABASE_ADMIN`
      (this should generally be your default)

        ```shell
        unset FIFTYONE_DATABASE_ADMIN
        ```

1. Use the license file provided by the Voxel51 Customer Success Team to create
   a new kubernetes secret

    ```shell
    kubectl --namespace your-namespace-here create secret generic \
      fiftyone-license --from-file=license=./your-license-file
    ```

1. [Upgrade to FiftyOne Enterprise version 2.14.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.14.0
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`

1. Upgrade all the datasets

    > **NOTE**: Any FiftyOne SDK less than 2.14.0 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.14.0` is recommended before migrating
    > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 1.8.0, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Enterprise Versions After 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Enterprise v2.14.0 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Enterprise Hosted
> Web App. You should coordinate this upgrade carefully with your
> end-users.

---

> **NOTE**: FiftyOne Enterprise v1.6 introduces the Central Authentication
> Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.14.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all the Auth0 configuration that was
> previously provided through kubernetes secrets. You may remove those secrets
> from your `values.yaml` and from any secrets created outside the Voxel51
> install process.
---

1. Ensure all FiftyOne SDK users either
    - Set the `FIFTYONE_DATABASE_ADMIN` to `false`

      ```shell
      FIFTYONE_DATABASE_ADMIN=false
      ```

    - Unset the environment variable `FIFTYONE_DATABASE_ADMIN`
      (this should generally be your default)

        ```shell
        unset FIFTYONE_DATABASE_ADMIN
        ```

1. Use the license file provided by the Voxel51 Customer Success Team to create
   a new kubernetes secret:

    ```shell
    kubectl --namespace your-namespace-here create secret generic \
      fiftyone-license --from-file=license=./your-license-file
    ```

1. In your `values.yaml`, set the required values
    1. `secret.fiftyone.encryptionKey` (or your deployment's
       equivalent)
        1. This sets the `FIFTYONE_ENCRYPTION_KEY` environment variable
           in the appropriate service pods
    1. `secret.fiftyone.fiftyoneAuthSecret` (or your deployment's equivalent)
        1. This sets the `FIFTYONE_AUTH_SECRET` environment variable
           in the appropriate service pods
1. [Upgrade to FiftyOne Enterprise version 2.14.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.14.0
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`

1. Upgrade all the datasets

    > **NOTE**: Any FiftyOne SDK less than 2.14.0 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.14.0` is recommended before migrating
    > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 1.8.0, run

    ```shell
    fiftyone migrate --info
    ```

### From Before FiftyOne Enterprise Version 1.1.0

> **NOTE**: Upgrading from versions of FiftyOne Enterprise prior to v1.1.0
> requires upgrading the database and will interrupt all SDK connections.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Enterprise v1.6 introduces the Central Authentication
> Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.14.0 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Enterprise Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.14.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets. You may remove those secrets
> from your `values.yaml` and from any secrets created outside of the Voxel51
> install process.

---

1. In your `values.yaml`, set the required values
    1. `secret.fiftyone.encryptionKey` (or your deployment's equivalent)
        1. This sets the `FIFTYONE_ENCRYPTION_KEY` environment variable
           in the appropriate service pods
    1. `secret.fiftyone.fiftyoneAuthSecret` (or your deployment's equivalent)
        1. This sets the `FIFTYONE_AUTH_SECRET` environment variable
           in the appropriate service pods
    1. `appSettings.env.FIFTYONE_DATABASE_ADMIN: true`
        1. This is not the default value in the Helm Chart and must be overridden
    1. When using path-based routing, update your ingress with the rule

        ```yaml
        ingress:
          paths:
            - path: /cas
              pathType: Prefix
              serviceName: teams-cas
              servicePort: 80
        ```

1. Use the license file provided by the Voxel51 Customer Success Team to create
   a new kubernetes secret:

    ```shell
    kubectl --namespace your-namespace-here create secret generic \
      fiftyone-license --from-file=license=./your-license-file
    ```

1. [Upgrade to FiftyOne Enterprise v2.14.0](#upgrading-from-previous-versions)
    > **NOTE**: At this step, FiftyOne SDK users will lose access to the
    > FiftyOne Enterprise Database until they upgrade to `fiftyone==2.14.0`
1. Upgrade your FiftyOne SDKs to version 2.14.0
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Enterprise version, navigate to
      `Account > Install FiftyOne`

1. Upgrade all the datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 1.8.0, run

    ```shell
    fiftyone migrate --info
    ```

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this
upgrade.
You will need to either create a new Identity Provider (IdP) or modify your
existing configuration to migrate to a new Auth0 Tenant.

<!-- Reference Links -->
[init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
