<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

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

As an example, you might use the following Dockerfile to build a
custom `your-internal-registry/fiftyone-app-internal` image:

```dockerfile
ARG TEAMS_IMAGE_NAME

FROM python:3.10 as wheelhouse

RUN pip wheel --wheel-dir=/tmp/wheels pandas

FROM ${TEAMS_IMAGE_NAME} as pandarelease

RUN --mount=type=cache,from=wheelhouse,target=/wheelhouse,ro \
    pip --no-cache-dir install -q --no-index \
    --find-links=/wheelhouse/tmp/wheels pandas
```

With a Dockerfile like this, you could use the following commands to
build, and publish, your image to your internal registry

```shell
TEAMS_VERSION=v1.7.0
docker buildx build --push \
  --build-arg TEAMS_IMAGE_NAME='voxel51/fiftyone-app:${TEAMS_VERSION}' \
  -t your-internal-registry/fiftyone-app-internal:${TEAMS_VERSION} .
```

You should upgrade your custom plugins image using the `TEAMS_VERSION`
you plan to use in your FiftyOne Teams Deployment.

## Using Your Custom Plugins Image in Docker Compose

After your custom plugins image is built, you can add it to your
`compose.override.yaml` file like

```yaml
services:
  teams-plugins:
    image: your-internal-registry/fiftyone-app-internal:v1.7.0
```

Please see
[Enabling FiftyOne Teams Plugins](../docker/README.md#enabling-fiftyone-teams-plugins)
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

Assuming you tagged your custom container with the same version
number as the FiftyOne Teams release, the Helm chart will
automatically use the chart version to pull your image.

Please see
[FiftyOne Teams Plugins](../helm/fiftyone-teams-app/README.md#fiftyone-teams-plugins)
for additional information regarding `teams-plugins` configuration.
