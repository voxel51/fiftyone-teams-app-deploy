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

# Configuring On-Demand Orchestrators

As of FiftyOne Enterprise v2.11.0, delegated operations can be run on-demand
in select external compute platforms (orchestrators).

This is the list of supported orchestrators and their configuration guides:

- [Anyscale](./orchestrators/configuring-anyscale-orchestrator.md)
- [Databricks](./orchestrators/configuring-databricks-orchestrator.md)
- [Kubernetes](./orchestrators/configuring-kubernetes-orchestrator.md)

If your deployment uses multimodal datasets, the orchestrator must install
the `multimodal` extra, `fiftyone[multimodal]`, so it can run projection
ingestion and compaction. See the multimodal configuration guide for
[docker](../docker/docs/configuring-multimodal.md#external-orchestrators)
or
[helm](../helm/docs/configuring-multimodal.md#external-orchestrators).
