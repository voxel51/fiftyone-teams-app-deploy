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

# Upgrading FiftyOne Enterprise

<!-- toc -->

- [Upgrading From Previous Versions](#upgrading-from-previous-versions)
  - [A Note On Database Migrations](#a-note-on-database-migrations)
  - [From FiftyOne Enterprise Version 2.0.0 and Later](#from-fiftyone-enterprise-version-200-and-later)
    - [FiftyOne Enterprise v2.7+ Delegated Operator Changes](#fiftyone-enterprise-v27-delegated-operator-changes)
    - [FiftyOne Enterprise v2.5+ Delegated Operator Changes](#fiftyone-enterprise-v25-delegated-operator-changes)
    - [FiftyOne Enterprise v2.2+ Delegated Operator Changes](#fiftyone-enterprise-v22-delegated-operator-changes)
    - [Delegated Operation Capacity](#delegated-operation-capacity)
    - [Existing Orchestrators](#existing-orchestrators)
  - [From FiftyOne Enterprise Versions 1.6.0 to 1.7.1](#from-fiftyone-enterprise-versions-160-to-171)
  - [From FiftyOne Enterprise Version 1.1.0 and Before Version 1.6.0](#from-fiftyone-enterprise-version-110-and-before-version-160)
  - [From Before FiftyOne Enterprise Version 1.1.0](#from-before-fiftyone-enterprise-version-110)
  - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)

<!-- tocstop -->

## Upgrading From Previous Versions

Voxel51 assumes you use the published Docker compose files to deploy your
FiftyOne Enterprise environment.
If you use custom deployment mechanisms, carefully review the changes in the
[Docker Compose Files](../)
and update your deployment accordingly.

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

### From FiftyOne Enterprise Version 2.0.0 and Later

1. [Upgrade to FiftyOne Enterprise version 2.11.2](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Enterprise SDK users to FiftyOne Enterprise
   version 2.11.2
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Voxel51 recommends that you upgrade all your datasets.

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at version 1.8.0, run

   ```shell
   fiftyone migrate --info
   ```

#### FiftyOne Enterprise v2.7+ Delegated Operator Changes

FiftyOne Enterprise v2.7.0 changes the `FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH`
environment variable to `FIFTYONE_DELEGATED_OPERATION_LOG_PATH`.
Please note that this change is backwards compatible, but should
be changed in your manifests moving forward.

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

To utilize the prior image, update your `common-services.yaml` similar to the below:

```yaml
teams-do-common:
  image: voxel51/fiftyone-app:v2.11.2
```

#### FiftyOne Enterprise v2.2+ Delegated Operator Changes

FiftyOne Enterprise v2.2 introduces some changes to delegated operators, detailed
below.

#### Delegated Operation Capacity

By default, all deployments are provisioned with capacity to support up to three
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

### From FiftyOne Enterprise Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Enterprise v2.11.2 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
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
   Enterprise docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. [Upgrade to FiftyOne Enterprise version 2.11.2](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.11.2
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE**: Any FiftyOne SDK less than 2.11.2
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.11.2` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 1.8.0, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Enterprise Version 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Enterprise v2.11.2 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Enterprise Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Enterprise v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.11.2 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
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
1. In your `.env` file, set the `LOCAL_LICENSE_FILE_DIR` variable value. Copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Enterprise docker compose host.

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

1. [Upgrade to FiftyOne Enterprise version 2.11.2](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.11.2
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE**: Any FiftyOne SDK less than 2.11.2
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.11.2` is required.

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

> **NOTE**: FiftyOne Enterprise v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](../README.md#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.11.2 _requires_ your users to
> log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Enterprise Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.11.2 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
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
1. In your `.env` file, set the `LOCAL_LICENSE_FILE_DIR` variable value. Copy
   the license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Enterprise docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. Update your web server routes to include routing
   `/cas/*` traffic to the `teams-cas` service.
   Please see our [example nginx configurations](../) for more information.
1. [Upgrade to FiftyOne Enterprise v2.11.2](#upgrading-from-previous-versions)
   with `FIFTYONE_DATABASE_ADMIN=true`
   (this is not the default for this release).
    > **NOTE**: FiftyOne SDK users will lose access to the FiftyOne
    > Enterprise Database at this step until they upgrade to `fiftyone==2.11.2`

1. Upgrade your FiftyOne SDKs to version 2.11.2
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Enterprise version, navigate to
      `Account > Install FiftyOne`
1. Confirm that datasets have been migrated to version 1.8.0

    ```shell
    fiftyone migrate --info
    ```

   - If not all datasets have been upgraded, have an admin run

      ```shell
      FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
      ```

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this
upgrade.
You will need to either create a new Identity Provider (IdP) or modify your
existing configuration to migrate to a new Auth0 Tenant.
