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

# Configuring FiftyOne Teams Delegated Operators

This option can be added to any of the three existing
[plugin modes](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/confuring-plugins.md).
If you're using the builtin-operator
only option, the Persistent Volume Claim should be omitted.

To enable this mode

- In `values.yaml`, set
  - `delegatedOperatorExecutorSettings.enabled: true`
  - The path for a Persistent Volume Claim mounted to the
    `teams-do` deployment in
    - `delegatedOperatorExecutorSettings.env.FIFTYONE_PLUGINS_DIR`
- See
  [Adding Shared Storage for FiftyOne Teams Plugins](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md)
  - Mount a Persistent Volume Claim (PVC) that provides
    - `ReadWrite` permissions to the `teams-do` deployment
      at the `FIFTYONE_PLUGINS_DIR` path

Optionally, the logs generated during running of a delegated operation can be
uploaded to a network-mounted file system or cloud storage path that is
available to this deployment. Logs are uploaded in the format
`<configured_path>/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log`
In `values.yaml`, set `configured_path`

- `delegatedOperatorExecutorSettings.env.FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH`

To use plugins with custom dependencies, build and use
[Custom Plugins Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).
