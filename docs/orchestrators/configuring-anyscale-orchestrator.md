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

# Anyscale On-Demand Orchestrator Setup

This document provides a step-by-step guide to configuring FiftyOne Enterprise
to use [Anyscale](https://www.anyscale.com/) as an orchestrator for running
delegated operations on-demand.

## Compute Config

Define and manage the machines, scaling, and other configurations for computer
resources [directly in Anyscale](https://docs.anyscale.com/configuration/compute/overview/).

## Dependency Management

[Dependency management](https://docs.anyscale.com/configuration/dependency-management/dependency-overview)
can be done using Anyscale Container Images. The images are written with
Docker-like specification and must be
[customized](https://docs.anyscale.com/configuration/dependency-management/dependency-container-images#customizing-a-container-image)
to include FiftyOne and, optionally, any custom operators.

Note: Some zoo models require additional packages. You can check the
requirements for any zoo model in the [fiftyone documentation](https://docs.voxel51.com/model_zoo/models.html):
find the model, then look under `Requirements` > `Packages`.

1. Determine base for Container Image using one of the following options:
   - An [Anyscale base image](https://docs.anyscale.com/reference/anyscale-base-images)
   - A custom image meeting the [requirements](https://docs.anyscale.com/configuration/dependency-management/image-requirement)
1. Add the FiftyOne Python package to the Container Image
   (_see following example Dockerfile_)

    ```dockerfile
    FROM anyscale/ray:2.46.0-slim-py312

    ARG FIFTYONE_ENTERPRISE_PYPI_TOKEN
    ARG FIFTYONE_ENTERPRISE_VERSION

    # Install system level packages here

    # Install fiftyone
    RUN pip install fiftyone==${FIFTYONE_ENTERPRISE_VERSION} \
      --index-url https://${FIFTYONE_ENTERPRISE_PYPI_TOKEN}@pypi.dev.fiftyone.ai/simple/ \
      --extra-index-url "https://pypi.org/simple"

    # Install extra python packages here.

    # Install custom operators here.

    # If MongoDB Atlas, set max process pool workers to avoid connection issues
    ENV FIFTYONE_MAX_PROCESS_POOL_WORKERS=4
    ```

1. Build and tag the Container Image (_see following example command_)

    ```commandline
    docker build  . \
          -t fiftyone-anyscale-example \
          --build-arg FIFTYONE_ENTERPRISE_VERSION=2.11 \
          --build-arg FIFTYONE_ENTERPRISE_PYPI_TOKEN=abc123
    ```

1. Push Container Image to an
  [Anyscale supported Docker registry](https://docs.anyscale.com/configuration/dependency-management/dependency-byod#step-2-push-your-image)

    ```commandline
    docker push fiftyone-anyscale-example
    ```

## Create Service Creds

Create service creds that FiftyOne will use, give the service access to run the
jobs you created above, and keep the following fields for providing to
FiftyOne:

- auth_token

## Register Orchestrator in FiftyOne

To register your orchestrator with FiftyOne, you can use the
[FiftyOne Management SDK](https://docs.voxel51.com/enterprise/management_sdk.html#module-fiftyone.management.orchestrator).
You will need to supply the environment you want to run your orchestrator
(`fom.OrchestratorEnvironment.ANYSCALE`), and then the configuration and
credential information needed to access that runner. To use the FiftyOne
Management SDK, you will also need an `API_URI` set in the environment or
FiftyOne configuration.

When registering your orchestrator with FiftyOne, you will need to supply
credential information, which is stored as a
[FiftyOne Secret](https://docs.voxel51.com/enterprise/secrets.html). The
`secrets` parameter to
[`fom.register_orchestrator()`](https://docs.voxel51.com/enterprise/management_sdk.html#fiftyone.management.orchestrator.register_orchestrator)
takes a top level key that must match your orchestrator environment. The
object that follows has key and value pairs that are specific to the
credentials needed to access your orchestrator.

When supplying one of the values, a new Secret will be created for you that
securely stores the information provided. These can be managed via the
Secrets manager.

Optionally, if you have an existing Secret that already has the credentials
you’d like to use, you can provide the name of that Secret and it will be used
instead of creating a new one. Examples of both options are included below.

Example snippet using the Management SDK to register an Anyscale orchestrator:

```python
import fiftyone.management as fom
fom.register_orchestrator(
    instance_id="your-orchestrator-name",
    description="Your orchestrator description",
    environment=fom.OrchestratorEnvironment.ANYSCALE,
    config={
        fom.OrchestratorEnvironment.ANYSCALE: {  # config
            "jobQueueName": "your-job-queue-name",
            "imageUri": "your-image-uri",
            "executionComputeConfig": "your-execution-compute-config",
            "registrationComputeConfig": "your-registration-compute-config",  # optional
            "idleTimeoutS": 300,  # optional, defaults to 300
            "pluginsDir": "your-plugins-dir",  # optional
        }
    },
    secrets={
        fom.OrchestratorEnvironment.ANYSCALE: {  # secrets
            "authToken": "your-anyscale-auth-token"
        }
    }
)
```

This will register a new orchestrator with the identifier
`your-orchestrator-name`.

Additionally, it will save a new Secret for the value supplied in authToken.
That new secret will have the following name:

`AUTH_TOKEN_YOUR_ORCHESTRATOR_NAME`

As noted above, if you already had Secrets saved with values you would like to
use, these names could be supplied in place of the values in the `secrets`
parameter. Here is an example:

```python
import fiftyone.management as fom
fom.register_orchestrator(
    instance_id="your-orchestrator-name",
    description="Your orchestrator description",
    environment=fom.OrchestratorEnvironment.ANYSCALE,
    config={
        fom.OrchestratorEnvironment.ANYSCALE: {  # config
            "jobQueueName": "your-job-queue-name",
            "imageUri": "your-image-uri",
            "executionComputeConfig": "your-execution-compute-config",
            "registrationComputeConfig": "your-registration-compute-config",  # optional
            "idleTimeoutS": 300,  # optional, defaults to 300
            "pluginsDir": "your-plugins-dir",  # optional
        }
    },
    secrets={
        fom.OrchestratorEnvironment.ANYSCALE: {  # secrets
            "authToken": "EXISTING_AUTH_TOKEN_SECRET"
        }
    }
)
```

In this case, new Secrets will not be created since valid names for existing
secrets have been provided. Those existing Secrets will be associated with the
orchestrator.

## Refresh Orchestrator Operators

Before you can do this step make sure you’ve added the optional dependency
`anyscale` into your FiftyOne API deployment, and set the environment
variable `API_EXTERNAL_URL` (the external Teams API base URL) used by Anyscale
workers to talk back to FiftyOne during registration/refresh. The external URL
can be found under `/settings/api_keys` in the UI. The API must be exposed
outside of the internal network
([helm](../../helm/docs/expose-teams-api.md) / [docker](../../docker/docs/expose-teams-api.md)).

This step is only required if you’ve added a plugin directory with custom
plugins to your Anyscale environment.

Once your orchestrator is registered in FiftyOne you can now refresh the
available operators for that environment. To do so, go to any dataset/runs page
and select your orchestrator on the right hand side.

Select the “refresh” button and click “confirm” when prompted. This will kick
off a job in your Anyscale that will tell FiftyOne what operators are available
in that environment. Once you see the job is complete, reload the page and
verify your “available operators” show the ones that you have configured.

In the future, anytime you add new operators to your Anyscale environment, you
will go through this same workflow or you can run that same task again directly
in Anyscale.

## Additional Considerations

Your Anyscale service account will need the following permissions for your
cloud storage platform of choice:

- Storage Bucket Viewer
- Storage Object Viewer
- Write permissions, If you setup
  [cloud storage logging](https://docs.voxel51.com/enterprise/plugins.html#logs)
- Blob sign permission, if the plugin uses signed URLs and your cloud platform
  requires additional permissions.

Additionally:

- `anyscale` SDK is not automatically built into the API image so you’ll need
  to add it as an extra dependency
- Make sure your API service has the environment variable `API_EXTERNAL_URL`
  set to your API_URI since this will be used to set the API endpoint in your
  Anyscale workers. Note: the provided deployment resources in this repo
  already include this environment variable.
- Due to a limitation discovered in the connection between Anyscale and
  MongoDB Atlas, using more than 4 parallel processes can lead to connection
  issues. We recommend setting the environment variable
  `FIFTYONE_MAX_PROCESS_POOL_WORKERS` to `4` in your created docker image
  to avoid this issue, if you are using MongoDB Atlas.
- If you still experience connection issues or database-stored cloud
  credentials are not being found, you should set
  `FIFTYONE_MAX_PROCESS_POOL_WORKERS` to `0` to disable multiprocessing.

## Credential Expiration and Rotation

In order to rotate your Anyscale credentials in FiftyOne:

1. Regenerate credentials through the Anyscale UI or SDK
1. Update the credentials in FiftyOne using the following FOM commands:

```python
import fiftyone.management as fom

orc = fom.get_orchestrator("<your-orc-instance-id>")
fom.update_secret(
   key=orc.secrets['auth_token'],
   value="<new_credentials>",
)
```
