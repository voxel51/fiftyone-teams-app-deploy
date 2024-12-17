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

The `NO_PROXY`, `no_proxy`, and `GLOBAL_AGENT_NO_PROXY` values must include the
Kubernetes service names that may communicate without going through a proxy
server.
By default, these service names are

- `fiftyone-app`
- `teams-app`
- `teams-api`
- `teams-cas`

This list may also include `teams-plugins` if you have enabled a dedicated
plugins service.

If the service names were overridden in `*.service.name`, use the override
values instead.

To configure this, set the following environment variables on

1. All pods, in the environment (`*.env`):

    ```yaml
    http_proxy: http://proxy.yourcompany.tld:3128
    https_proxy: https://proxy.yourcompany.tld:3128
    no_proxy: fiftyone-app, teams-app, teams-api, teams-cas, <your_other_exclusions>
    HTTP_PROXY: http://proxy.yourcompany.tld:3128
    HTTPS_PROXY: https://proxy.yourcompany.tld:3128
    NO_PROXY: fiftyone-app, teams-app, teams-api, teams-cas, <your_other_exclusions>
    ```

    > **NOTE**: If you have enabled a
    > [dedicated `teams-plugins`](../docs/plugin-configuration.md)
    > deployment you will need to include `teams-plugins` in your `NO_PROXY` and
    > `no_proxy` configurations

    ---

    > **NOTE**: If you have overridden your service names with `*.service.name`
    > you will need to include the override service names in your `NO_PROXY` and
    > `no_proxy` configurations instead

1. The deployments based on the `fiftyone-teams-app` (`teamsAppSettings.env`) or
   `fiftyone-teams-cas` (`casSettings.env`) images

    ```yaml
    GLOBAL_AGENT_HTTP_PROXY: http://proxy.yourcompany.tld:3128
    GLOBAL_AGENT_HTTPS_PROXY: https://proxy.yourconpay.tld:3128
    GLOBAL_AGENT_NO_PROXY: fiftyone-app, teams-app, teams-api, teams-cas, <your_other_exclusions>
    ```

    > **NOTE**: If you have enabled a
    > [dedicated `teams-plugins`](../docs/plugin-configuration.md)
    > deployment you will need to include `teams-plugins` in your
    > `GLOBAL_AGENT_NO_PROXY` configuration

    ---

    > **NOTE**: If you have overridden your service names with `*.service.name`
    > you will need to include the override service names in your
    > `GLOBAL_AGENT_NO_PROXY` configuration instead

By default, the Global Agent Proxy will log all outbound connections
and identify which connections are routed through the proxy.
To reduce the logging verbosity, add this environment variable to your `teamsAppSettings.env`

```yaml
teamsAppSettings:
  env:
    ROARR_LOG: false
```
