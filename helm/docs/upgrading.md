# Upgrading FiftyOne Teams

<!-- toc -->

- [Upgrading From Previous Versions](#upgrading-from-previous-versions)
  - [From FiftyOne Teams Version 2.0.0 or Higher](#from-fiftyone-teams-version-200-or-higher)
    - [FiftyOne Teams v2.7+ Delegated Operator Changes](#fiftyone-teams-v27-delegated-operator-changes)
    - [FiftyOne Teams v2.5+ Delegated Operator Changes](#fiftyone-teams-v25-delegated-operator-changes)
    - [FiftyOne Teams v2.2+ Delegated Operator Changes](#fiftyone-teams-v22-delegated-operator-changes)
      - [Delegated Operation Capacity](#delegated-operation-capacity)
      - [Existing Orchestrators](#existing-orchestrators)
    - [Version 2.2+ InitContainers Additions](#version-22-initcontainers-additions)
  - [From FiftyOne Teams Versions 1.6.0 to 1.7.1](#from-fiftyone-teams-versions-160-to-171)
  - [From FiftyOne Teams Versions After 1.1.0 and Before Version 1.6.0](#from-fiftyone-teams-versions-after-110-and-before-version-160)
  - [From Before FiftyOne Teams Version 1.1.0](#from-before-fiftyone-teams-version-110)
  - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)

<!-- tocstop -->

## Upgrading From Previous Versions

Voxel51 assumes you use the published Helm Chart to deploy your FiftyOne Teams
environment.
If you are using a custom deployment mechanism, carefully review the changes in
the
[Helm Chart](https://github.com/voxel51/fiftyone-teams-app-deploy)
and update your deployment accordingly.

A minimal example `values.yaml` may be found
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/values.yaml).

1. Edit the `values.yaml` file
1. To upgrade an existing helm installation

    1. Make sure you have followed the appropriate directions for
       [Upgrading From Previous Versions](#upgrading-from-previous-versions)

    1. Update your kubectl configuration to set your current namespace for
       your kubectl context

        ```shell
        kubectl config set-context --current --namespace your-namespace-here
        ```

    1. Update your Voxel51 Helm repository and upgrade your FiftyOne Teams
       deployment

        ```shell
        helm repo update voxel51
        helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
          -f ./values.yaml
        ```

    > **NOTE** To view the changes Helm would apply during installations
    > and upgrades, consider using
    > [helm diff](https://github.com/databus23/helm-diff).
    > Voxel51 is not affiliated with the author of this plugin.
    >
    > For example:
    >
    > ```shell
    > helm diff -C1 upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
    > ```

### From FiftyOne Teams Version 2.0.0 or Higher

1. [Upgrade to FiftyOne Teams version 2.6.1](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Teams SDK users to FiftyOne Teams
   version 2.6.1
    1. Login to the FiftyOne Teams UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Voxel51 recommends that you upgrade all your datasets, but it is not
   required.

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at version 1.3.1, run

   ```shell
   fiftyone migrate --info
   ```

#### FiftyOne Teams v2.7+ Delegated Operator Changes

FiftyOne Teams v2.7.0 changes the `FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH`
environment variable to `FIFTYONE_DELEGATED_OPERATION_LOG_PATH`.
Please note that this change is backwards compatible, but should
be changed in your manifests moving forward.

#### FiftyOne Teams v2.5+ Delegated Operator Changes

FiftyOne Teams v2.6 changes the base image of the built-in delegated
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

#### FiftyOne Teams v2.2+ Delegated Operator Changes

FiftyOne Teams v2.2 introduces some changes to delegated operators, detailed
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
Any image supporting `nslookup`, such as `docker.io/busybox`, are applicable
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

### From FiftyOne Teams Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Teams v2.6.1 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
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
   a new kubernetes secret:

    ```shell
    kubectl --namespace your-namespace-here create secret generic \
      fiftyone-license --from-file=license=./your-license-file
    ```

1. [Upgrade to FiftyOne Teams version 2.6.1](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.6.1
    1. Login to the FiftyOne Teams UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    > **NOTE** Any FiftyOne SDK less than 2.6.1 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.6.1` is recommended before migrating
    > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 1.3.1, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Versions After 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Teams v2.6.1 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted
> Web App. You should coordinate this upgrade carefully with your
> end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.6.1 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
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
1. [Upgrade to FiftyOne Teams version 2.6.1](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.6.1
    1. Login to the FiftyOne Teams UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    > **NOTE** Any FiftyOne SDK less than 2.6.1 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.6.1` is recommended before migrating
    > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 1.3.1, run

    ```shell
    fiftyone migrate --info
    ```

### From Before FiftyOne Teams Version 1.1.0

> **NOTE**: Upgrading from versions of FiftyOne Teams prior to v1.1.0
> requires upgrading the database and will interrupt all SDK connections.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.6.1 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.6.1 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
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

1. [Upgrade to FiftyOne Teams v2.6.1](#upgrading-from-previous-versions)
    > **NOTE**: At this step, FiftyOne SDK users will lose access to the
    > FiftyOne Teams Database until they upgrade to `fiftyone==2.6.1`
1. Upgrade your FiftyOne SDKs to version 2.6.1
    1. Login to the FiftyOne Teams UI
    1. To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 1.3.1, run

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
