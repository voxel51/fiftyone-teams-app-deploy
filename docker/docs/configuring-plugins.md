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

# Configuring FiftyOne Enterprise Plugins

## Builtin Plugins Only

Enabled by default. No additional configurations are required.

## Shared Plugins

Plugins run in the `fiftyone-app` service.
To enable this mode

1. Use the file
[legacy-auth/compose.plugins.yaml](../legacy-auth/compose.plugins.yaml)
instead of
[legacy-auth/compose.yaml](../legacy-auth/compose.yaml)
as part of your `docker compose` command to create a new Docker Volume shared
between FiftyOne Enterprise services.

Example `docker compose` command for this mode from the `legacy-auth` directory

```shell
docker compose \
  -f compose.plugins.yaml \
  -f compose.override.yaml \
  up -d
```

## Dedicated Plugins

Plugins run in the `teams-plugins` service.
To enable this mode

1. make sure `FIFTYONE_TEAMS_PLUGIN_URL` is set in your `.env`
file

    - `FIFTYONE_TEAMS_PLUGIN_URL=http://teams-plugins:5151`

1. Use the file
[legacy-auth/compose.dedicated-plugins.yaml](../legacy-auth/compose.dedicated-plugins.yaml)
instead of
[legacy-auth/compose.yaml](../legacy-auth/compose.yaml).
as part of your `docker comppse` command to create a new Docker Volume shared
between FiftyOne Enterprise services.

1. If you are
  [using a proxy](./configuring-proxies.md),
  add the `teams-plugins` service name to your environment variables

    - `no_proxy`
    - `NO_PROXY`

Example `docker compose` command for this mode from the `legacy-auth` directory

```shell
docker compose \
  -f compose.dedicated-plugins.yaml \
  -f compose.override.yaml \
  up -d
```
