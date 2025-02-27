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

# Configuring FiftyOne Enterprise Plugins

## Builtin Plugins Only

Enabled by default. No additional configurations are required.

## Shared Plugins

Plugins run in the `fiftyone-app` service.
To enable this mode, use the file
[legacy-auth/compose.plugins.yaml](legacy-auth/compose.plugins.yaml)
instead of
[legacy-auth/compose.yaml](legacy-auth/compose.yaml)
to create a new Docker Volume shared between FiftyOne Enterprise
services.

1. Configure the services to access to the plugin volume

- `fiftyone-app` requires `read`
- `fiftyone-api` requires `read-write`

1. Example `docker compose` command for this mode from the `legacy-auth`
directory

    ```shell
    docker compose \
      -f compose.plugins.yaml \
      -f compose.override.yaml \
      up -d
    ```

## Dedicated Plugins

Plugins run in the `teams-plugins` service.
To enable this mode, use the file
[legacy-auth/compose.dedicated-plugins.yaml](legacy-auth/compose.dedicated-plugins.yaml)
instead of
[legacy-auth/compose.yaml](legacy-auth/compose.yaml).
to create a new Docker Volume shared between FiftyOne Enterprise
services.

1. Configure the services to access to the plugin volume

- `teams-plugins` requires `read`
- `fiftyone-api` requires `read-write`

1. If you are
  [using a proxy](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/docs/configuring-proxies.md),
  add the `teams-plugins` service name to your environment variables

- `no_proxy`
- `NO_PROXY`

1. Example `docker compose` command for this mode from the `legacy-auth`
  directory

    ```shell
    docker compose \
      -f compose.dedicated-plugins.yaml \
      -f compose.override.yaml \
      up --d
    ```
