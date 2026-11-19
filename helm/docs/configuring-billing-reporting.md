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

Deployments with a usage-based billing license get a **Settings > Billing**
page in the FiftyOne Enterprise App. Admins see the reporting status, the
current quarter's invoice estimate, and the reporting actions: submit a usage
report now, or, on air-gapped deployments, download the usage report and
upload the enablement token Customer Success returns.

`FIFTYONE_BILLING_REPORTING` on `teams-api` sets how much of that page the
deployment serves:

| Value            | Settings > Billing                                    | `teams-api`                                                                   |
| ---------------- | ----------------------------------------------------- | ----------------------------------------------------------------------------- |
| `full` (default) | Status, invoice estimate and reporting actions        | Serves everything                                                             |
| `status`         | Status and reporting actions, without the estimate    | Never returns the invoice estimate                                            |
| `off`            | Hidden from the settings navigation, with its queries | Refuses the billing details query and the submit, download and upload actions |

`teams-api` enforces the value, so the API refuses the same reads and actions
the page hides. An unrecognized value is treated as `off`.

The setting does not change usage reporting itself: `teams-cas` keeps
reporting on its normal schedule, and admins still see the banner when the
enablement token is due for renewal or has expired.

## Setting it

Set the value in `apiSettings.env` in your `values.yaml`, then upgrade the
release:

```yaml
apiSettings:
  env:
    FIFTYONE_BILLING_REPORTING: status
```
