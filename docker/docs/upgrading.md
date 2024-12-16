# Upgrading FiftyOne Teams

## Upgrading From Previous Versions

Voxel51 assumes you use the published Docker compose files to deploy your
FiftyOne Teams environment.
If you use custom deployment mechanisms, carefully review the changes in the
[Docker Compose Files](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/main/docker)
and update your deployment accordingly.

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this
upgrade.
You will need to either create a new Identity Provider (IdP) or modify your
existing configuration to migrate to a new Auth0 Tenant.

### From Before FiftyOne Teams Version 1.1.0

> **NOTE**: Upgrading from versions of FiftyOne Teams prior to v1.1.0
> requires upgrading the database and will interrupt all SDK connections.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](../README.md#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.3.0 _requires_ your users to log in
> after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.3.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
> install process.

---

1. Copy your `compose.override.yaml` and `.env` files into the `legacy-auth`
   directory
1. `cd` into the `legacy-auth` directory
1. In your `.env` file, set the required environment variables:
    - `FIFTYONE_ENCRYPTION_KEY`
    - `FIFTYONE_API_URI`
    - `FIFTYONE_AUTH_SECRET`
1. Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Teams docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. Ensure your web server routes are updated to include routing
   `/cas/*` traffic to the `teams-cas` service.
   Example nginx configurations can be found
   [here](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/main/docker)
1. [Upgrade to FiftyOne Teams v2.3.0](#upgrading-from-previous-versions)
   with `FIFTYONE_DATABASE_ADMIN=true`
   (this is not the default for this release).
    > **NOTE**: FiftyOne SDK users will lose access to the FiftyOne
    > Teams Database at this step until they upgrade to `fiftyone==2.2.0`

1. Upgrade your FiftyOne SDKs to version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Teams version, navigate to
      `Account > Install FiftyOne`
1. Confirm that datasets have been migrated to version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

   - If not all datasets have been upgraded, have an admin run

      ```shell
      FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
      ```

### From FiftyOne Teams Version 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Teams v2.3.0 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.3.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through  environment variables; you may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
> install process.

---

1. Copy your `compose.override.yaml` and `.env` files into the `legacy-auth`
   directory
1. `cd` into the `legacy-auth` directory
1. In the `.env` file, set the required environment variables
    - `FIFTYONE_API_URI`
    - `FIFTYONE_AUTH_SECRET`
    - `CAS_BASE_URL`
    - `CAS_BIND_ADDRESS`
    - `CAS_BIND_PORT`
    - `CAS_DATABASE_NAME`
    - `CAS_DEBUG`
    - `CAS_DEFAULT_USER_ROLE`

    > **NOTE**: For the `CAS_*` variables, consider using
    > the seed values from the `.env.template` file.
    > See
    > [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
1. Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Teams docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

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

1. [Upgrade to FiftyOne Teams version 2.2.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE** Any FiftyOne SDK less than 2.2.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.2.0` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 0.25.1, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Teams v2.3.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
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

1. Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Teams docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. [Upgrade to FiftyOne Teams version 2.2.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE** Any FiftyOne SDK less than 2.2.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.2.0` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 0.25.1, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Version 2.0.0

1. [Upgrade to FiftyOne Teams version 2.2.0](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Teams SDK users to FiftyOne Teams
   version 2.2.0, but it is not required
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

### From FiftyOne Teams Version 2.1.3

1. [Upgrade to FiftyOne Teams version 2.2.0](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Teams SDK users to FiftyOne Teams
   version 2.2.0, but it is not required
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Voxel51 recommends that you upgrade all your datasets, but it is not
   required.  Users using the FiftyOne Teams 2.0.0 SDK will continue to operate
   uninterrupted during, and after, this migration

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at version 1.1.0, run

   ```shell
   fiftyone migrate --info
   ```

#### FiftyOne Teams v2.2+ Delegated Operator Changes

FiftyOne Teams v2.2 introduces some changes to delegated operators, detailed
below.

#### Delegated Operation Capacity

By default, all deployments are provisioned with capacity to support up to 3
delegated operations simultaneously. You will need to configure the
[builtin orchestrator](../README.md#builtin-delegated-operator-orchestrator)
or an external
orchestrator, with enough workers, to be able to utilize this full capacity.
If your team finds the usage is greater than this, please reach out to your
Voxel51 support team for guidance and to increase this limit!

#### Existing Orchestrators

> [!NOTE]
> If you are currently utilizing an external orchestrator for delegated
> operations, such as Airflow or Flyte, you may have an outdated execution
> definition that could negatively affect the experience. Please reach out to
> Voxel51 support team for guidance on updating this code.

Additionally,

> [!WARNING]
> If you cannot update the orchestrator DAG/workflow code, you must set
> `FIFTYONE_ALLOW_LEGACY_ORCHESTRATORS=true` in order for the delegated
> operation system to function properly.