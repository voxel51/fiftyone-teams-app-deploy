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

# Configuring Proxies

FiftyOne Teams supports routing traffic through proxy servers.
To configure this, set following environment variables in your
`compose.override.yaml`

1. All services

    ```yaml
    http_proxy: ${HTTP_PROXY_URL}
    https_proxy: ${HTTPS_PROXY_URL}
    no_proxy: ${NO_PROXY_LIST}
    HTTP_PROXY: ${HTTP_PROXY_URL}
    HTTPS_PROXY: ${HTTPS_PROXY_URL}
    NO_PROXY: ${NO_PROXY_LIST}
    ```

1. All services based on the `fiftyone-teams-app` and `fiftyone-teams-cas`
   images

    ```yaml
    GLOBAL_AGENT_HTTP_PROXY: ${HTTP_PROXY_URL}
    GLOBAL_AGENT_HTTPS_PROXY: ${HTTPS_PROXY_URL}
    GLOBAL_AGENT_NO_PROXY: ${NO_PROXY_LIST}
    ```

The environment variable `NO_PROXY_LIST` value should be a comma-separated list
of Docker Compose services that may communicate without going through a proxy
server. By default these service names are:

- `fiftyone-app`
- `teams-api`
- `teams-app`
- `teams-cas`
- `teams-plugins`

Examples of these settings are included in the FiftyOne Teams configuration files

- [common-services.yaml](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/common-services.yaml)
- [legacy-auth/env.template](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/legacy-auth/env.template)

By default, the Global Agent Proxy will log all outbound connections and
identify which connections are routed through the proxy.
To reduce the logging verbosity, add this environment variable to your
`teams-app` and `teams-cas` services.

```yaml
services:
  teams-app:
    environment:
      ROARR_LOG: false
  teams-cas:
    environment:
      ROARR_LOG: false
```
