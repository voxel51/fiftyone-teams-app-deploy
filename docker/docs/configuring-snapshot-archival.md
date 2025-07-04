<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img alt="Voxel51 Logo" src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px">
&nbsp;
<img alt="Voxel51 FiftyOne" src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# Configuring Snapshot Archival

Since version v1.5, FiftyOne Enterprise supports
[archiving snapshots](https://docs.voxel51.com/enterprise/dataset_versioning.html#snapshot-archival)
to cold storage locations to prevent filling up the MongoDB database. To enable
this feature, set the `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH` environment variable to
the path of a chosen storage location.

Supported locations are network-mounted filesystems and cloud storage folders.

- Network-mounted filesystem
  - Set the environment variable `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH` to the
    mounted filesystem path in these containers
    - `teams-api`
    - `teams-app`
  - Mount the filesystem to the `teams-api` container (`teams-app` does not need
    this despite the variable set above). For an example, see
    [legacy-auth/compose.plugins.yaml](../legacy-auth/compose.plugins.yaml).
- Cloud storage folder
  - Set the environment variable `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH` to a cloud
    storage path (for example
    `gs://my-voxel51-bucket/dev-deployment-snapshot-archives/`) in these
    containers
    - `teams-api`
    - `teams-app`
  - Ensure the
    [cloud credentials](https://docs.voxel51.com/enterprise/installation.html#cloud-credentials)
    loaded in the `teams-api` container have full edit capabilities to this
    bucket

See the
[configuration documentation](https://docs.voxel51.com/enterprise/dataset_versioning.html#dataset-versioning-configuration)
for other configuration values that control the behavior of automatic snapshot
archival.
