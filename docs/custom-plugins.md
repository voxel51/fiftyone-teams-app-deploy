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

# Custom Plugins Images

Some plugins have custom python dependencies,
which requires the creation of a new plugins image.
This document outlines the steps Voxel51 recommends
for creating those custom plugins containers.

## Create a New Image From an Existing Voxel51 Image

By default, dedicated plugins use the `voxel51/fiftyone-app` image.
Basing your custom image on the existing base image will ensure that
the required `fiftyone` packages and configurations are available.

The builtin delegated operation orchestrator (`teams-do`) uses a different
image, `voxel51/fiftyone-teams-cv-full`, which adds computer vision and
machine learning libraries.
If your plugins run delegated operations on the builtin orchestrator,
build a second custom image from `voxel51/fiftyone-teams-cv-full`.
Basing it on `voxel51/fiftyone-app` instead would remove those libraries.

As an example, you might use the following Dockerfile to build a
custom `your-internal-registry/fiftyone-app-internal` image:

```dockerfile
ARG FIFTYONE_ENTERPRISE_IMAGE_NAME

FROM python:3.10 as wheelhouse

RUN pip wheel --wheel-dir=/tmp/wheels pandas

FROM ${FIFTYONE_ENTERPRISE_IMAGE_NAME} as pandarelease

RUN --mount=type=cache,from=wheelhouse,target=/wheelhouse,ro \
    pip --no-cache-dir install -q --no-index \
    --find-links=/wheelhouse/tmp/wheels pandas
```

With a Dockerfile like this, you could use the following commands to
build, and publish, your image to your internal registry

```shell
FIFTYONE_ENTERPRISE_VERSION=v2.26.0
docker buildx build --push \
  --build-arg FIFTYONE_ENTERPRISE_IMAGE_NAME="voxel51/fiftyone-app:${FIFTYONE_ENTERPRISE_VERSION}" \
  -t your-internal-registry/fiftyone-app-internal:${FIFTYONE_ENTERPRISE_VERSION} .
```

The same Dockerfile builds the custom delegated operation orchestrator
image, with `voxel51/fiftyone-teams-cv-full` as the base:

```shell
FIFTYONE_ENTERPRISE_VERSION=v2.26.0
docker buildx build --push \
  --build-arg FIFTYONE_ENTERPRISE_IMAGE_NAME="voxel51/fiftyone-teams-cv-full:${FIFTYONE_ENTERPRISE_VERSION}" \
  -t your-internal-registry/fiftyone-teams-cv-full-internal:${FIFTYONE_ENTERPRISE_VERSION} .
```

You should upgrade your custom plugins image using the `FIFTYONE_ENTERPRISE_VERSION`
you plan to use in your FiftyOne Enterprise Deployment.

## Using Your Custom Plugins Image in Docker Compose

After your custom plugins image is built, you can add it to your
`compose.override.yaml` file like

```yaml
services:
  teams-plugins:
    image: your-internal-registry/fiftyone-app-internal:v2.26.0
```

If you are using the builtin delegated operation orchestrator, also set
your custom orchestrator image on `teams-do`
(and on `teams-do-2` and `teams-do-3`, if you enabled them):

```yaml
services:
  teams-do:
    image: your-internal-registry/fiftyone-teams-cv-full-internal:v2.26.0
```

Please see
[Enabling FiftyOne Enterprise Plugins](../docker/README.md#fiftyone-enterprise-plugins)
for example `docker compose` commands for starting and upgrading your
deployment.

## Using Your Custom Plugins Image in Helm Deployments

After your custom plugins image is built, you can add it to your
`values.yaml` file like

```yaml
pluginsSettings:
  image:
    repository: your-internal-registry/fiftyone-app-internal
```

If you are using the builtin delegated operation orchestrator, set your
custom orchestrator image in the `template` for always-on workers
(`delegatedOperatorDeployments`) and, if you use them, for on-demand
Kubernetes jobs (`delegatedOperatorJobTemplates`):

```yaml
delegatedOperatorDeployments:
  template:
    image:
      repository: your-internal-registry/fiftyone-teams-cv-full-internal

delegatedOperatorJobTemplates:
  template:
    image:
      repository: your-internal-registry/fiftyone-teams-cv-full-internal
```

Each `template` applies to every deployment or job.
An image set on a single deployment or job overrides it.

> [!NOTE]
> Earlier versions of this guide used `delegatedOperatorExecutorSettings`.
> The chart no longer reads that key, so an image set there is ignored
> and the orchestrator runs the default image.

Assuming you tagged your custom containers with the same version
number as the FiftyOne Enterprise release, the Helm chart will
automatically use the chart version to pull your images.

Please see
[FiftyOne Enterprise Plugins](../helm/fiftyone-teams-app/README.md#plugins)
for additional information regarding `teams-plugins` configuration, and
[Configuring Delegated Operators](../helm/docs/configuring-delegated-operators.md)
for orchestrator configuration.
