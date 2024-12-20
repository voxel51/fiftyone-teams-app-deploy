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

# Configuring Snapshot Archival

Since version v1.5, FiftyOne Teams supports
[archiving snapshots](https://docs.voxel51.com/teams/dataset_versioning.html#snapshot-archival)
to cold storage locations to prevent filling up the MongoDB database.
To enable this feature, set the `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
environment variable to the path of a chosen storage location.

Supported locations are network-mounted filesystems and cloud storage folders.

## Network-mounted filesystem

- In `values.yaml`, set the path for a Persistent Volume Claim mounted to the
    `teams-api` deployment (not necessary to mount to other deployments) in both
  - `appSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
  - `teamsAppSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
- Mount a Persistent Volume Claim with `ReadWrite` permissions to
    the `teams-api` deployment at the `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH` path.
    For an example, see
    [Plugins Storage][plugins-storage].

## Cloud storage folder

- In `values.yaml`, set the cloud storage path (for example
    `gs://my-voxel51-bucket/dev-deployment-snapshot-archives/`)
    in
  - `appSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
  - `apiSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
- Ensure the
    [cloud credentials](https://docs.voxel51.com/teams/installation.html#cloud-credentials)
    loaded in the `teams-api` deployment have full edit capabilities to this bucket

See the
[configuration documentation](https://docs.voxel51.com/teams/dataset_versioning.html#dataset-versioning-configuration)
for other configuration values that control the behavior of automatic snapshot archival.

[plugins-storage]: https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md
