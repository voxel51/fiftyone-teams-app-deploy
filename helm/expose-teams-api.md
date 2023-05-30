<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>

---

# Exposing the `teams-api` service

There are some limited instances in which you may wish to expose your FiftyOne Teams API for SDK access.

You can expose your `teams-api` service in any manner that suits your deployment strategy; the following is one solution, but does not represent the entirety of possible solutions.  Essentially any solution that allows the FiftyOne Teams SDK to access port 80 on the `teams-api` service should work.


## Adding a second host to the Ingress Controller

1. set `apiSettings.dnsName` to the hostname to route API requests to (e.g. demo-api.fiftyone.ai)
2. upgrade your deployment using the v1.3.0 Helm chart:
    ```
	helm repo update voxel51
    helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
	```
