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

## Known Issues for FiftyOne Enterprise v2.26.0 and Above

### Invitations After Upgrading From Legacy Authentication Mode

FiftyOne Enterprise v2.26.0 removes `legacy` authentication mode and moves
existing `legacy` deployments to `internal` authentication mode on startup.
In `legacy` mode, Auth0 sent invitations to join your organization.
After the upgrade, FiftyOne Enterprise sends them instead, and only if
invitations are enabled for your organization in the CAS SuperAdmin UI.
Check that setting after upgrading. To send invitations as emails, you must
also configure an SMTP connection in the CAS SuperAdmin UI.

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
