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

# Databricks On-Demand Orchestrator Setup

This document provides a step-by-step guide to configuring FiftyOne Enterprise
to use [Databricks](https://www.databricks.com/) as an orchestrator for running
delegated operations on-demand.

[Databricks SDK documentation](https://databricks-sdk-py.readthedocs.io/en/latest/)

## Introduction

This document outlines the steps necessary to configure your FiftyOne
Enterprise system to send Delegated Operations to your Databricks environment
for execution, on-demand.

## Create requirements.txt

[Databricks DBFS documentation](https://docs.databricks.com/files/)

Databricks executors need to define the dependencies necessary for executing a
delegated operation. The below script will create a requirements.txt file
with the minimum required dependencies for running builtin operations.

> **NOTE**: If you experience [dependency conflicts](#dependency-conflicts)
> between FiftyOne and the Databricks base image,  please contact your customer
> success representative for assistance in resolving them.

- If you have custom operators that require additional dependencies you will
add them here.
- Some zoo models require additional packages. You can check the requirements
for any zoo model in the [FiftyOne documentation](https://docs.voxel51.com/model_zoo/models.html):
find the model, then look under `Requirements` > `Packages`.

Save your `DBFS_PATH` for later as it will be used when creating your
job configuration. The script will also create the file in your Databricks
account using the Databricks SDK. Alternatively you can create this manually in
the UI.

If you’d prefer to build [your own image](https://docs.databricks.com/aws/en/compute/custom-containers),
Databricks offers that as well.

```python
from databricks.sdk import WorkspaceClient
from databricks.sdk.service.workspace import ImportFormat

w = WorkspaceClient()

# TODO: replace with path of your choice in DBFS
DBFS_PATH = "/FileStore/my_project/requirements.txt"

PYTHON_DEPENDENCIES = [
   "fiftyone==2.13.1",  # use your FiftyOne version here
   "ultralytics",
   "torch",
   "transformers",
   "timm",
   "umap-learn"
]
print("CREATING requirements.txt")
file_content_str = "\n".join(PYTHON_DEPENDENCIES)

content_as_bytes = file_content_str.encode('utf-8')

print(f"Uploading to Workspace path: {DBFS_PATH}")
w.workspace.upload(
  path=DBFS_PATH,
  content=content_as_bytes,
  overwrite=True,
  format=ImportFormat.RAW
)

print("SUCCESS")
```

## Create Databricks secrets

[Databricks secrets documentation](https://docs.databricks.com/aws/en/security/secrets/)

Your executor requires environment variables containing certain secrets that
correspond to your FiftyOne deployment. These secrets are: Mongo database URI,
FiftyOne encryption key, and FiftyOne pypi url. To follow security best
practices, the code below will create secrets in Databricks. Keep the path of
these secrets with the scope you create, which should look something like:
`secrets/your-scope/FIFTYONE_DATABASE_URI`. We will use these secrets when
creating the job config environment variables.

```python
from databricks.sdk import WorkspaceClient

w = WorkspaceClient()

# TODO: replace with your actual scope name
SCOPE_NAME = "your-scope"

# TODO: replace with your actual secrets
SECRETS_TO_CREATE = {
   "FIFTYONE_DATABASE_URI": "YOUR_ACTUAL_DATABASE_URI_HERE",
   "FIFTYONE_ENCRYPTION_KEY": "YOUR_ACTUAL_ENCRYPTION_KEY_HERE",
   "FIFTYONE_PYPI_URL": "https://your.company.pypi/simple"
}

print("CREATING SECRETS")

for key, value in SECRETS_TO_CREATE.items():
   print(f"  - Creating secret '{key}'...")
   w.secrets.put_secret(
       scope=SCOPE_NAME,
       key=key,
       string_value=value
   )

print("SUCCESS")
```

## Create Job Entrypoint

Below is the entry point for any FiftyOne Enterprise job that should exist in
your Databricks file system (DBFS). This is a simple script that allows
the FiftyOne API to send arbitrary
[FiftyOne CLI commands](https://docs.voxel51.com/cli/index.html)
to be executed for running Delegated Operators and orchestrator registration.
Make sure to keep
the path where you’ve uploaded the script; we will be using that when creating
the job config. This can be uploaded directly to your Databricks account, or you
can use the script in the next section to do that using the Databricks SDK.

```python
import subprocess
import nest_asyncio
import argparse
import shlex


parser = argparse.ArgumentParser(description="Run a command via subprocess.")
parser.add_argument(
    "--command",
    type=str,
    help="The full command string to execute."
)
args = parser.parse_args()


cmd = shlex.split(args.command)


print(f"Executing command: {' '.join(cmd)}")
nest_asyncio.apply()


result = subprocess.run(
    cmd,
    check=True,
    text=True,
)
print("\nProcess completed successfully!")
```

If you created this locally and want to upload it using the Databricks SDK use
this:

```python
import os
from databricks.sdk import WorkspaceClient
from databricks.sdk.service.workspace import ImportFormat

w = WorkspaceClient()

# This must be a path on your local computer.
# Assuming 'entrypoint.py' is in the same directory as this script.
LOCAL_FILE_PATH = "entrypoint.py"

# This is the destination path in your Databricks Workspace.
WORKSPACE_DESTINATION_PATH = f"/Workspace/some_path/{os.path.basename(LOCAL_FILE_PATH)}"

print("UPLOADING JOB SCRIPT")
with open(LOCAL_FILE_PATH, "rb") as f:
  w.workspace.upload(
      path=WORKSPACE_DESTINATION_PATH,
      content=f.read(),
      overwrite=True,
      format=ImportFormat.RAW,
  )
print("SUCCESS")
```

You can read more about the
[FiftyOne CLI](https://docs.voxel51.com/cli/index.html)
in our docs.

## Create Instance Pool

Databricks [instance pools](https://docs.databricks.com/aws/en/compute/pool-best-practices)
documentation

An instance pool is used to specify worker scaling, compute, and lifetime. You
should configure these to match your compute needs, but below is an example of
how to create a simple CPU backed pool. Make sure to keep the generated
instance pool id as we will use that when creating the job config.

```python
from databricks.sdk import WorkspaceClient

# TODO: configure as you want
POOL_CONFIG = {
   "instance_pool_name": "your-actual-pool-name",
   "node_type_id": "c3-standard-4-lssd",
   "min_idle_instances": 0,
   "max_capacity": 20,
   "enable_elastic_disk": True,
   "idle_instance_autotermination_minutes": 15,
   "preloaded_spark_versions": ["16.4.x-scala2.13"]
}

w = WorkspaceClient()

print("CREATING INSTANCE POOL")

new_pool_info = w.instance_pools.create(**POOL_CONFIG)

print("SUCCESS")
print(f"Save your pool id for later:   {new_pool_info.instance_pool_id}")
```

### Optional Registration Instance Pool

Part of the FiftyOne plugin workflow is that workers report their available
operators to FiftyOne to avoid users attempting to run custom code that might
not be available in the given environment. This process is triggered on demand
by an Admin in FiftyOne and will run a small registration script that informs
it of what plugins are available in your Databricks environment. To avoid
wasting expensive compute you can optionally create a second worker pool to be
used by just that script. An example might look like:

- Main compute: expensive GPU to run auto labeling jobs
- Registration compute: cheap CPU to inform us what operators are available in
  your environment

If your main compute is cost effective to be used in both cases, then feel free
to just create the one worker pool above.

## Setup Plugin Volume

Plugins allow for custom functionality to be run in FiftyOne or delegated to
your Databricks orchestrator. Built-in plugins are available out of the box
with FiftyOne, but Databricks will need access to a plugin directory to execute
custom plugins. There are many ways to set this up, but here are some examples:

- Upload to DBFS
- Download the directory from cloud storage in your `init.sh` or your startup
  script
- Give shared volume access to your Databricks

Regardless of your chosen solution, save the absolute file path to be used in
the `FIFTYONE_PLUGINS_DIR` environment variable when setting up your job
config. Read more about configuring plugins for
[helm](../helm/docs/configuring-plugins.md) and
[docker](../docker/docs/configuring-plugins.md).

## Create Job

[Databricks job documentation](https://docs.databricks.com/aws/en/jobs/automate)

All of the previous steps were to provide the necessary configurations for your
job. Jobs have many options so feel free to edit this to your liking. Below is
just a basic example using the values you should have saved in the previous
steps. The minimum result is a job with one task responsible for executing the
[entrypoint we uploaded previously](#create-job-entrypoint). This task should
use workers from your instance pool, dependencies from your
[requirements.txt](#create-requirementstxt), and have the required environment
variables including the [secrets you created](#create-databricks-secrets).

Note: you can change `max_concurrent_runs` to limit how many jobs can run at
once; this should likely match the deployment's delegated operations capacity.

Once you’ve created your job, note the Job ID, Execution Task ID and
[Optional Registration Task ID](#optional-registration-instance-pool) (not
necessary if you’ve removed it), we will use these when registering your
endpoint in FiftyOne. Note: You can remove the optional registration task and
registration task cluster below if you are okay with on-demand registration
happening in your execution cluster.

```python
from databricks.sdk.service.jobs import JobSettings as Job

# Replace these
ENTRYPOINT_PATH = "dbfs:/your-path/entrypoint.py"
REQUIREMENTS_PATH = "dbfs:/your-path/requirements.txt"
EXECUTION_CLUSTER_NAME = "gpu_cluster"
REGISTRATION_CLUSTER_NAME = "cpu_cluster"
EXECUTION_POOL_ID = ""
REGISTRATION_POOL_ID = ""

# Replace these
ENV_VARS = {
   "FIFTYONE_DATABASE_NAME": "\"fiftyone\"",
   "FIFTYONE_INTERNAL_SERVICE": "1",
   "FIFTYONE_DATABASE_URI": "{{secrets/your-scope/FIFTYONE_DATABASE_URI}}",
   "FIFTYONE_ENCRYPTION_KEY": "{{secrets/your-scope/FIFTYONE_ENCRYPTION_KEY}}",
   "API_URL": "",
   "PIP_EXTRA_INDEX_URL": "{{secrets/your-scope/FIFTYONE_PYPI_URL}}",
   "FIFTYONE_PLUGINS_DIR": "\"/Workspace/your-plugin-dir/plugins\"",
   "FIFTYONE_PLUGINS_CACHE_ENABLED": "true",
   "FIFTYONE_MAX_PROCESS_POOL_WORKERS": "4",
}

demo_job = Job.from_dict(
   {
       "name": "demonstration-task-processor",
       "max_concurrent_runs": 5,
       "tasks": [
           {
               "task_key": "execute_task",
               "spark_python_task": {
                   "python_file": ENTRYPOINT_PATH,
                   "parameters": [
                       "--command",
                       "{{job.parameters.command}}",
                   ],
               },
               "job_cluster_key": EXECUTION_CLUSTER_NAME,
               "libraries": [
                   {
                       "requirements": REQUIREMENTS_PATH,
                   },
               ],
           },
           {
               "task_key": "register_task",
               "spark_python_task": {
                   "python_file": ENTRYPOINT_PATH,
                   "parameters": [
                       "--command",
                       "{{job.parameters.command}}",
                   ],
               },
               "job_cluster_key": REGISTRATION_CLUSTER_NAME,
               "libraries": [
                   {
                       "requirements": REQUIREMENTS_PATH,
                   },
               ],
           },
       ],
       "job_clusters": [
           {
               "job_cluster_key": EXECUTION_CLUSTER_NAME,
               "new_cluster": {
      "use_ml_runtime": True,
                   "spark_version": "16.4.x-scala2.13",
                   "spark_env_vars": ENV_VARS,
                   "instance_pool_id": EXECUTION_POOL_ID,
                   "data_security_mode": "DATA_SECURITY_MODE_DEDICATED",
                   "runtime_engine": "STANDARD",
                   "kind": "CLASSIC_PREVIEW",
                   "is_single_node": False,
                   "num_workers": 1,
               },
           },
           {
               "job_cluster_key": REGISTRATION_CLUSTER_NAME,
               "new_cluster": {
                   "spark_version": "16.4.x-scala2.13",
                   "spark_env_vars": ENV_VARS,
                   "instance_pool_id": REGISTRATION_POOL_ID,
                   "data_security_mode": "DATA_SECURITY_MODE_DEDICATED",
                   "runtime_engine": "STANDARD",
                   "kind": "CLASSIC_PREVIEW",
                   "is_single_node": False,
                   "num_workers": 1,
               },
           },
       ],
       "queue": {
           "enabled": True,
       },
       "parameters": [
           {
               "name": "command",
               "default": "fiftyone --version",
           },
       ],
   }
)

from databricks.sdk import WorkspaceClient

w = WorkspaceClient()
w.jobs.create(**demo_job.as_shallow_dict())

print("SUCCESS")
```

## Create Service Creds

Create service creds that FiftyOne will use, give the service access to run
the jobs you created above and keep the following fields for providing to
FiftyOne:

- Host
  - The URL you view your account at eg: `https://1290481.3.gcp.databricks.com/`
- Account Id
  - Account drop down (top right)
  - Manage account
  - Account id can be found in the URL
- Client Id
  - Settings
  - Identity and Access
  - Service Principals
  - Create your own
  - Secrets > Generate secrets
  - It will display the client id as well as the client secret
- Client secret
  - See client ID steps

## Register Orchestrator in FiftyOne

To register your orchestrator with FiftyOne, you can use the
[FiftyOne Management SDK](https://docs.voxel51.com/enterprise/management_sdk.html#module-fiftyone.management.orchestrator).
You will need to supply the environment you want to run your
orchestrator (`fom.OrchestratorEnvironment.DATABRICKS`), and then the
configuration and credential information needed to access that runner. To use
the FiftyOne Management SDK, you will also need an `API_URI` set in the
environment or FiftyOne configuration.

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

Example snippet using the Management SDK to register a Databricks orchestrator:

```python
import fiftyone.management as fom
fom.register_orchestrator(
    instance_id="your-orchestrator-name",
    description="Your orchestrator description",
    environment=fom.OrchestratorEnvironment.DATABRICKS,
    config={
        fom.OrchestratorEnvironment.DATABRICKS: {  # config
            "jobId": "your-job-id",
            "executionTaskId": "your-execution-task-id",
            "registrationTaskId": "your-registration-task-id",  # optional
        }
    },
    secrets={
        fom.OrchestratorEnvironment.DATABRICKS: {  # secrets
            "host": "your-databricks-host",
            "accountId": "your-databricks-account-id",
            "clientId": "your-databricks-client-id",
            "clientSecret": "your-databricks-client-secret",  # pragma: allowlist secret
        },
    },
)
```

This will register a new orchestrator with the identifier
`your-orchestrator-name`.

Additionally, it will save four new Secrets, one each for
`host, accountId, clientId, clientSecret`. Those new secrets will have the
following names, respectively:

`HOST_YOUR_ORCHESTRATOR_NAME`
`ACCOUNT_ID_YOUR_ORCHESTRATOR_NAME`
`CLIENT_ID_YOUR_ORCHESTRATOR_NAME`
`CLIENT_SECRET_YOUR_ORCHESTRATOR_NAME`

As noted above, if you already had Secrets saved with values you would like to
use, these names could be supplied in place of the values in the `secrets`
parameter. Here is an example:

```python
import fiftyone.management as fom
fom.register_orchestrator(
   instance_id="your-orchestrator-name",
   description="Your orchestrator description",
   environment=fom.OrchestratorEnvironment.DATABRICKS,
   config={
       fom.OrchestratorEnvironment.DATABRICKS: {  # config
           "jobId": "your-job-id",
           "executionTaskId": "your-execution-task-id",
           "registrationTaskId": "your-registration-task-id"  # optional
       }
   },
   secrets={
       fom.OrchestratorEnvironment.DATABRICKS: {  # secrets
           "host": "EXISTING_HOST_SECRET",
           "accountId": "EXISTING_ACCOUNT_ID_SECRET",
           "clientId": "EXISTING_CLIENT_ID_SECRET",
           "clientSecret": "EXISTING_CLIENT_SECRET_SECRET"  # pragma: allowlist secret
       },
   },
)
```

In this case, new Secrets will not be created since valid names for existing
secrets have been provided. Those existing Secrets will be associated with the
orchestrator.

## Refresh Orchestrator Operators

Before doing this step make sure your FiftyOne API deployment has the optional
dependency “databricks-sdk”. It is not built into our deployments by default so
you’ll need to add it by following the
[Custom Plugins images docs](../custom-plugins.md#custom-plugins-images).

This step is only required if you’ve added a plugin directory with custom
plugins to your Databricks environment.

Once your orchestrator is registered in FiftyOne you can now refresh the
available operators for that environment. To do so, go to any dataset/runs page
and select your orchestrator on the right hand side.

Select the “refresh” button and click “confirm” when prompted. This will
kick off a job in your Databricks that will tell FiftyOne what operators are
available in that environment. Once you see the job is complete, reload the
page and verify your “available operators” show the ones that you have
configured.

In the future, anytime you add new operators to your Databricks environment,
you will go through this same workflow or you can run that same task again
directly in Databricks.

## Additional Considerations

Your Databricks service account will need the following permissions for your
cloud storage platform of choice:

- Storage Bucket Viewer
- Storage Object Viewer
- Write permissions, If you setup
  [cloud storage logging](https://docs.voxel51.com/enterprise/plugins.html#logs)
- Blob sign permission, if the plugin uses signed URLs and your cloud platform
  requires additional permissions.

Additionally:

- `databricks-sdk` is not automatically built into the API image so you’ll need
  to add it as an extra dependency.
  See the
  [Custom Plugins images docs](../custom-plugins.md#custom-plugins-images).
- Due to a limitation discovered in the connection between Databricks and
  MongoDB Atlas, using more than 4 parallel processes can lead to connection
  issues. We recommend setting the environment variable
  `FIFTYONE_MAX_PROCESS_POOL_WORKERS` to `4` in your job config to avoid
  this issue, if you are using MongoDB Atlas.
- If you still experience connection issues or database-stored cloud
  credentials are not being found, you should set
  `FIFTYONE_MAX_PROCESS_POOL_WORKERS` to `0` to disable multiprocessing.

### Credential Expiration and Rotation

The Databricks credentials that FiftyOne use can expire and so will need to be
rotated regularly.

In order to rotate your Databricks credentials in FiftyOne:

1. Regenerate credentials through Databricks UI or SDK
1. Update the credentials in FiftyOne using the following FOM commands:

```python
import fiftyone.management as fom

orc = fom.get_orchestrator("<your-orc-instance-id>")
fom.update_secret(
   key=orc.secrets['client_secret'],
   value="<new_credentials>",
)
```

## Common Issues

### Dependency Conflicts

Databricks surfaces dependency conflicts in multiple ways typically
during the image build or image execution steps of a job. Some
errors we have seen before as a result of conflicts are:

- `Could not reach driver of cluster`
- `Cannot read the python file`
- `Library installation error`
- `The requested operation requires that "some-dependency==X" is
    installed on your machine, but found "some-dependency==Y"`

Conflicts of this nature are often unique to your dependency versions,
but if you are unable to resolve them please reach out to customer
success.
