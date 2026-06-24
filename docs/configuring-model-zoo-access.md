# Controlling Access to Models or Datasets in the Enterprise App

Administrators may control which public Models and Datasets are made
available to users in the Enterprise App via an **allowlist of licenses**
or an **allowlist of names**.

Set any of the following Enterprise
[Secrets](https://docs.voxel51.com/enterprise/secrets.html#fiftyone-enterprise-secrets)
to configure allowlists. Each accepts a comma-separated list.

| Secret                                  | Effect                                                        |
|-----------------------------------------|---------------------------------------------------------------|
| `FIFTYONE_ZOO_ALLOWED_MODEL_LICENSES`   | Include only models distributed under one of these licenses   |
| `FIFTYONE_ZOO_ALLOWED_DATASET_LICENSES` | Include only datasets distributed under one of these licenses |
| `FIFTYONE_ZOO_ALLOWED_MODEL_NAMES`      | Include only these specific models                            |
| `FIFTYONE_ZOO_ALLOWED_DATASET_NAMES`    | Include only these specific datasets                          |

## Setting the secrets

An admin adds these in the Enterprise App under **Settings > Secrets**:

1. Click **Add secret**.
2. Enter the **Key** (upper snake case, e.g. `FIFTYONE_ZOO_ALLOWED_MODEL_LICENSES`).
3. Enter the **Value** (the comma-separated list).

## Examples

Restrict to permissively licensed models and datasets only:

| Key                                     | Value                    |
|-----------------------------------------|--------------------------|
| `FIFTYONE_ZOO_ALLOWED_MODEL_LICENSES`   | `MIT,Apache 2.0`         |
| `FIFTYONE_ZOO_ALLOWED_DATASET_LICENSES` | `CC-BY-SA-3.0,CC-BY-4.0` |

Expose only an explicit, vetted set of entries:

| Key                                  | Value                                                              |
|--------------------------------------|--------------------------------------------------------------------|
| `FIFTYONE_ZOO_ALLOWED_MODEL_NAMES`   | `clip-vit-base32-torch,zero-shot-classification-transformer-torch` |
| `FIFTYONE_ZOO_ALLOWED_DATASET_NAMES` | `quickstart,coco-2017`                                             |

## Discovering available models and datasets

The following FiftyOne CLI commands can be used to discover Models and
Datasets available in the FiftyOne Zoo.

```shell
fiftyone zoo models list
fiftyone zoo datasets list
```

The output includes each entry's license, which is the value you match
against with the `_LICENSES` variables above.

## See also

- Model Zoo documentation: <https://docs.voxel51.com/model_zoo/index.html>
- FiftyOne Enterprise Secrets: <https://docs.voxel51.com/enterprise/secrets.html>
