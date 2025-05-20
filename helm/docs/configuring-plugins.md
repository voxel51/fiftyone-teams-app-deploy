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

Enabled by default.
No additional configurations are required.

## Shared Plugins

Plugins run in the `fiftyone-app` deployment.
To enable this mode

- [Add Shared Storage for FiftyOne Enterprise Plugins][./plugins-storage]
  - Mount a Persistent Volume Claim (PVC) that provides
    - `ReadWrite` permissions to the `teams-api` deployment
      at the `FIFTYONE_PLUGINS_DIR` path
    - `ReadOnly` permission to the `fiftyone-app` deployment
      at the `FIFTYONE_PLUGINS_DIR` path
- In `values.yaml`, set the path for a PVC
  mounted to the `teams-api` and `fiftyone-app` deployments, and enable
  plugins caching
  - `apiSettings.env.FIFTYONE_PLUGINS_DIR: /path/to/pvc`
  - `appSettings.env.FIFTYONE_PLUGINS_DIR: /path/to/pvc`

## Dedicated Plugins

To enable this mode

- [Add Shared Storage for FiftyOne Enterprise Plugins][./plugins-storage]
  - Mount a Persistent Volume Claim (PVC) that provides
    - `ReadWrite` permissions to the `teams-api` deployment
      at the `FIFTYONE_PLUGINS_DIR` path
    - `ReadOnly` permission to the `teams-plugins` deployment
      at the `FIFTYONE_PLUGINS_DIR` path
- In `values.yaml`, set
  - `pluginsSettings.enabled: true`
  - the path for a PVC mounted to the `teams-api` and `teams-plugins`
    deployments, and enable plugins caching
    - `apiSettings.env.FIFTYONE_PLUGINS_DIR: /path/to/pvc`
    - `pluginsSettings.env.FIFTYONE_PLUGINS_DIR: /path/to/pvc`
- If you are
  [using a proxy](./configuring-proxies.md),
  add the `teams-plugins` service name to your `no_proxy` and
  `NO_PROXY` environment variables.

[plugins-storage]: ,/plugins-storage.md
