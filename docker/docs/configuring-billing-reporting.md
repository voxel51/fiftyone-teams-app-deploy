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

# Configuring the Billing Page

Deployments with a usage-based billing license
get a **Settings > Billing** page in the FiftyOne Enterprise App.
On this page, admins see:

- The reporting status
- The current quarter's invoice estimate
- The reporting actions:
  - Submit a usage report now
  - On air-gapped deployments,
    download the usage report
    and upload the enablement token that Customer Success sends back

The `teams-api` service environment variable `FIFTYONE_BILLING_REPORTING`
sets how much the billing page shows:

| Value            | Settings > Billing                                       | `teams-api`                                                                    |
| ---------------- | -------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `full` (default) | Shows status, invoice estimate, and reporting actions    | Serves everything                                                              |
| `status`         | Shows status and reporting actions, but no estimate      | Never returns the invoice estimate                                             |
| `off`            | Hidden from the settings menu, and its queries don't run | Refuses the billing details query and the submit, download, and upload actions |

`teams-api` enforces the value,
so the API refuses the same reads and actions that the page hides.
`teams-api` treats an unrecognized value as `off`.

The setting doesn't change usage reporting itself.
`teams-cas` keeps reporting on its normal schedule.
Admins still see the banner
when the enablement token is due for renewal or has expired.

## Set the value

Set the value in your `.env` file:

```shell
FIFTYONE_BILLING_REPORTING=<BILLING_VALUE>
```

Then run your usual `docker compose ... up -d` command
with the same `-f` files you deploy with.
Docker Compose recreates `teams-api` with the new value.
For example, with dedicated plugins:

```shell
docker compose \
  -f compose.yaml \
  -f compose.dedicated-plugins.yaml \
  -f compose.override.yaml \
  up -d
```

If you leave out the `-f` files,
Docker Compose loads only `compose.yaml` and `compose.override.yaml`.
On a deployment with plugins,
that recreates `teams-api` without its plugin settings.
