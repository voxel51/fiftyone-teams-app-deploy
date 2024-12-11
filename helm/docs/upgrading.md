# Upgrading FiftyOne Teams

Voxel51 assumes you use the published
Helm Chart to deploy your FiftyOne Teams environment.
If you are using a custom deployment
mechanism, carefully review the changes in the
[Helm Chart](https://github.com/voxel51/fiftyone-teams-app-deploy)
and update your deployment accordingly.

## Upgrading From Previous Versions

A minimal example `values.yaml` may be found
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/values.yaml).

1. Edit the `values.yaml` file
1. Deploy FiftyOne Teams with `helm install`
    1. For a new installation
        1. Create a new namespace and set the current namespace for your kubectl
           context

           ```shell
           kubectl create namespace your-namespace-here
           kubectl config set-context --current --namespace your-namespace-here
           ```

        1. If you are using the Voxel51 DockerHub registry to install your
           container images, use the Voxel51-provided DockerHub credentials to
           create an Image Pull Secret, and uncomment the `imagePullSecrets`
           section of your `values.yaml`

           ```shell
           kubectl --namespace your-namespace-here create secret generic \
           regcred --from-file=.dockerconfigjson=./voxel51-docker.json \
           --type kubernetes.io/dockerconfigjson
           ```

        1. Use your Voxel51-provided License file to create a FiftyOne License
           Secret

           ```shell
           kubectl --namespace your-namepace-here create secret generic \
           fiftyone-license --from-file=license=./your-license-file
           ```

        1. Add the Voxel51 Helm repository and install FiftyOne Teams

           ```shell
           helm repo add voxel51 https://helm.fiftyone.ai
           helm repo update voxel51
           helm install fiftyone-teams-app voxel51/fiftyone-teams-app \
           -f ./values.yaml
           ```

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

        > **NOTE**  To view the changes Helm would apply during installations
        > and upgrades, consider using
        > [helm diff](https://github.com/databus23/helm-diff).
        > Voxel51 is not affiliated with the author of this plugin.
        >
        > For example:
        >
        > ```shell
        > helm diff -C1 upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
        > ```

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success
team member to coordinate this upgrade.
You will need to either create a new Identity Provider (IdP)
or modify your existing configuration to migrate to a new Auth0 Tenant.

### From Before FiftyOne Teams Version 1.1.0

> **NOTE**: Upgrading from versions of FiftyOne Teams prior to v1.1.0
> requires upgrading the database and will interrupt all SDK connections.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](../README.new.md)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.1.3 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.1.3 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
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
       (add it before the `path: /` rule)

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
    kubectl --namespace your-namepace-here create secret generic \
        fiftyone-license --from-file=license=./your-license-file
    ```

1. [Upgrade to FiftyOne Teams v2.1.3](#upgrading-from-previous-versions)
    > **NOTE**: At this step, FiftyOne SDK users will lose access to the
    > FiftyOne Teams Database until they upgrade to `fiftyone==2.1.3`
1. Upgrade your FiftyOne SDKs to version 2.1.3
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. Validate that all datasets are now at version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Versions After 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Teams v2.1.3 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted
> Web App. You should coordinate this upgrade carefully with your
> end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](../README.new.md)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.1.3 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
> from your `values.yaml` and from any secrets created outside of the Voxel51
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
    kubectl --namespace your-namepace-here create secret generic \
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
1. [Upgrade to FiftyOne Teams version 2.1.3](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.1.3
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    > **NOTE** Any FiftyOne SDK less than 2.1.3 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.1.3` is recommended before migrating
        > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. Validate that all datasets are now at version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Teams v2.1.3 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
> from your `values.yaml` and from any secrets created outside of the Voxel51
> install process.

---

> **NOTE**: If you had previously set
> `teamsAppSettings.env.FIFTYONE_APP_INSTALL_FIFTYONE_OVERRIDE` to include your
> Voxel51 private PyPI token, you can remove it from your configuration. The
> Voxel51 private PyPI token is now loaded correctly from your license file.

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
    kubectl --namespace your-namepace-here create secret generic \
        fiftyone-license --from-file=license=./your-license-file
    ```

1. [Upgrade to FiftyOne Teams version 2.1.3](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.1.3
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    > **NOTE** Any FiftyOne SDK less than 2.1.3 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.1.3` is recommended before migrating
        > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. Validate that all datasets are now at version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Version 2.0.0

1. [Upgrade to FiftyOne Teams version 2.1.3](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Teams SDK users to FiftyOne Teams
   version 2.1.3, but it is not required
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Voxel51 recommends that you upgrade all your datasets, but it is not
   required.  Users using the FiftyOne Teams 2.0.0 SDK will continue to operate
   uninterrupted during, and after, this migration

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at version 0.25.1, run

   ```shell
   fiftyone migrate --info
   ```