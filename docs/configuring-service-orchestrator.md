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

# Configuring Service Orchestrators

Some builtin operators run as long-lived services (always-on model servers)
rather than on-demand delegated jobs. A service orchestrator is the
delegated-operator worker that hosts these services and keeps them reachable
from `teams-api`.

Services are declared in the deployment's builtin services list, which
`teams-api` mounts and reconciles at startup. Each ships created stopped. Start
one from the FiftyOne Enterprise UI under `Settings -> Services`.

The builtin services are:

- `annotation-ai` (SAM2) powers AI-assisted segmentation in the annotation
  editor. Needs a GPU.
- `agentic-labeler` powers few-shot VLM labeling. Needs a GPU.

Configure the service orchestrator for your deployment:

- [Docker Compose](../docker/docs/configuring-agentic-labeler.md)
- [Kubernetes](../helm/docs/configuring-delegated-operators.md)

## Broker settings

The service broker runs inside `teams-api`. Set these variables on the
`teams-api` deployment (Kubernetes) or the `teams-api` service (Docker Compose)
to tune it.

| Variable | Default | Description |
| --- | --- | --- |
| `FIFTYONE_SERVICE_RECONCILE_DELAY_S` | `300` | Seconds after `teams-api` starts before it reconciles builtin services. Auto-start can queue heavy GPU work, so it is held off until the deployment settles. Set to `0` to reconcile as soon as the server starts. |
| `FIFTYONE_SERVICE_POD_READY_TIMEOUT_S` | `900` | Seconds the broker waits for a service pod to become ready before it times out and tears the pod down. A cold GPU start that provisions a node from zero can take several minutes, so raise this if pods are killed before they finish starting. Kubernetes service broker only. |
| `FIFTYONE_SERVICE_POD_READY_POLL_INTERVAL_S` | `2` | Seconds between pod readiness checks while the broker waits. Kubernetes service broker only. |
| `FIFTYONE_SERVICE_POD_SERVICE_ACCOUNT` | auto | ServiceAccount for service pods. Defaults to the `teams-api` ServiceAccount so pods inherit its Workload Identity and cloud access. Set to override. Kubernetes service broker only. |
