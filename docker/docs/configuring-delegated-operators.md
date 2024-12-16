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

This option can be added to any of the 3 existing
[plugin modes](./configuring-plugins.md).

To enable this mode and launch worker containers, use
[legacy-auth/compose.delegated-operators.yaml](legacy-auth/compose.delegated-operators.yaml)
in conjunction with one of the 3 plugin configurations.

- Example `docker compose` command for enabling this mode on top of dedicated
  plugins mode, from the `legacy-auth` directory

    ```shell
    docker compose \
      -f compose.dedicated-plugins.yaml \
      -f compose.delegated-operators.yaml \
      -f compose.override.yaml \
      up --d
    ```

Optionally, the logs generated during running of a delegated operation can be
uploaded to a network-mounted file system or cloud storage path that is
available to this deployment. Logs are uploaded in this format:
`<configured_path>/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log`
Set `FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH` to `configured_path`

To use plugins with custom dependencies, build and use
[Custom Plugins Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).
