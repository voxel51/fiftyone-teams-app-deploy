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

# Adding Shared Storage for FiftyOne Teams Plugins

This document will provide guidance for adding shared
storage for FiftyOne Teams Plugins using

* Persistent Volumes (PVs)
* Persistent Volume Claims (PVCs)
* NFS

Alternate storage solutions vary based on cloud providers and infrastructure services.
PVCs may be configured using provider specific services

* Google Cloud
  [Filestore](https://cloud.google.com/filestore/docs)
* Amazon
  [EFS](https://aws.amazon.com/efs/)
* Azure
  [Files](https://learn.microsoft.com/en-us/azure/storage/files/).

## NFS Share

Configured NFS server with an exported share that
grants permission to the kubernetes cluster.
One such configuration would be to share the `/exports/deployment_name/plugins`
directory using a configuration like

```shell
$ cat /etc/exports
/exports 10.202.15.0/24(rw,insecure,fsid=0,root_squash,all_squash,no_subtree_check)
/exports/fiftyone_teams_app 10.202.15.0/24(rw,insecure,no_root_squash,anonuid=1000,anongid=1000,no_subtree_check)
/exports/fiftyone_teams_app/plugins 10.202.15.0/24(rw,insecure,no_root_squash,anonuid=1000,anongid=1000,no_subtree_check)
```

We recommend that you test this export to make sure
the NFS configuration is accurate before proceeding.
Testing now will save frustration later.

## PV and PVC Creation

The following yaml configuration will create a PV and
PVC designed to access the NFS share established above

```yaml
# nfs-pv-pvc.yaml
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: teams-plugins-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
    - ReadOnlyMany
  nfs:
    server: nfs-server
    path: "/fiftyone_teams_app/plugins"

---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: teams-plugins-pvc
spec:
  accessModes:
    - ReadWriteMany
    - ReadOnlyMany
  storageClassName: ""
  resources:
    requests:
      storage: 10Gi
```

Apply the configuration to create the PV and PVC for plugins storage

```shell
kubectl apply -f nfs-pv-pvc.yaml
```

## Add PVs to `values.yaml`

To run plugins in the `fiftyone-app` deployment, provide

* `ReadOnly` access to the `fiftyone-app` deployment
* `ReadWrite` access to the `teams-api` deployment

To run plugins in a dedicated `teams-plugins` deployment, provide

* `ReadOnly` access to the `teams-plugins` deployment
* `ReadWrite` access to the `teams-api` deployment

To run delegated operators in the builtin orchestrator, additionally provide

* `ReadOnly` access to the `teams-do` deployment

Add the appropriate `volumes` and `volumeMounts` configurations
to the `apiSettings` section of your `values.yaml`

```yaml
# values.yaml
apiSettings:
  [...existing config...]
  volumes:
    - name: nfs-plugins-vol
      persistentVolumeClaim:
        claimName: teams-plugins-pvc
  volumeMounts:
    - name: nfs-plugins-vol
      mountPath: /opt/plugins
  [...existing config...]
```

Add the `volumes` and `volumeMounts` configurations to either
the `pluginsSettings` or `appSettings` section of your `values.yaml`. And
optionally the `delegatedOperatorExecutorSettings` section for the builtin
delegated operator orchestrator.

```yaml
# values.yaml
[plugins|app]Settings:
  [...existing config...]
  volumes:
    - name: nfs-plugins-ro-vol
      persistentVolumeClaim:
        claimName: teams-plugins-pvc
        readOnly: true
  volumeMounts:
    - name: nfs-plugins-ro-vol
      mountPath: /opt/plugins
  [...existing config...]
```

## Apply the Changes to the Existing fiftyone-teams-app Deployment

Apply the changes to the existing fiftyone-teams-app deployment using Helm

```shell
helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
  -f values.yaml
```
