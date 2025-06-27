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

# Known Issues

## Known Issues for FiftyOne Enterprise v1.6.0 and Above

### Invitations Disabled for Internal Authentication Mode

FiftyOne Enterprise v1.6 introduces the Central Authentication Service (CAS),
which includes both [`legacy` authentication mode][legacy-auth-mode] and
[`internal` authentication mode][internal-auth-mode].

Prior to v2.2.0, inviting users to join your FiftyOne Enterprise instance was
not supported when `FIFTYONE_AUTH_MODE` is set to `internal`. After v2.2.0+, you
can enable invitations for your organization through the CAS SuperAdmin UI. To
enable sending invitations as emails, you must also configure an SMTP
connection.

[internal-auth-mode]:
  https://docs.voxel51.com/enterprise/pluggable_auth.html#internal-mode
[legacy-auth-mode]:
  https://docs.voxel51.com/enterprise/pluggable_auth.html#legacy-mode

## Delegated Operators: Troubleshooting

### :brain: Handling `DataLoader worker exited unexpectedly` Errors

This error often occurs when using Torch-based models that rely on
`torch.multiprocessing`, which utilizes **shared memory (`/dev/shm`)** to
exchange data between processes.

If the available shared memory is insufficient, plugins running these models may
fail with:

```txt
DataLoader worker exited unexpectedly
```

This can happen in:

- The **plugin container** (e.g., `teams-plugins`)
- The **delegated operator container** (e.g., `teams-do`)

### :hammer_and_wrench: Solution: Increase Shared Memory Allocation

To resolve this, increase the shared memory (`shm_size`) available to the
affected service.

For example, in `compose.delegated-operators.yaml`:

```yaml
services:
  teams-do:
    extends:
      file: ../common-services.yaml
      service: teams-do-common
    shm_size: "512m"
```

> :repeat: You can adjust the value based on your workload (e.g., `'1g'` for
> larger models).

If you’re using shared plugins (i.e., plugins running inside `fiftyone-app`),
you may need to apply the same `shm_size` setting to that service instead.

### :mag_right: Tip

To verify shared memory usage and limits inside a container, run:

```bash
docker exec -it <container_name> df -h /dev/shm
```
