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

# Configuring the Content Security Policy

As of FiftyOne Enterprise v2.28.0, the FiftyOne Enterprise App
(`teams-app`) sends an enforcing
[Content-Security-Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP)
header with every response, along with cross-origin headers.
The policy restricts which hosts the browser may load scripts, styles, fonts,
images, media, frames and network requests from.

The defaults work for standard deployments.
You only need the settings in this guide when something your deployment uses
is blocked.

The policy for the authentication service (`teams-cas`) is fixed in its image
and has no settings.

## When You Need These Settings

- A custom plugin loads scripts, styles or fonts from a CDN,
  or embeds an iframe.
  See [Custom Plugins](#custom-plugins).
- Another site embeds the App in an iframe.
  See [Embedding the App in Another Site](#embedding-the-app-in-another-site).
- You turned off any-host media and your media lives outside the default hosts.
  See [Restricting Media Hosts](#restricting-media-hosts).
- Something is blocked and you do not know what.
  See [Diagnosing a Blocked Resource](#diagnosing-a-blocked-resource).

## Settings

All settings are environment variables on the `teams-app` container.
Changes take effect when the container restarts.
No new image is needed.

| Variable | Default | Purpose |
| --- | --- | --- |
| `FIFTYONE_CSP_MEDIA_ANY_HOST` | `true` | Allow media and fetches from any host. |
| `FIFTYONE_CSP_EXTRA_SOURCES` | empty | Add hosts to images, media and fetches. |
| `FIFTYONE_CSP_EXTRA_SCRIPT_SOURCES` | empty | Add script hosts. |
| `FIFTYONE_CSP_EXTRA_STYLE_SOURCES` | empty | Add style hosts. |
| `FIFTYONE_CSP_EXTRA_FONT_SOURCES` | empty | Add font hosts. |
| `FIFTYONE_CSP_EXTRA_IMG_SOURCES` | empty | Add image hosts. |
| `FIFTYONE_CSP_EXTRA_MEDIA_SOURCES` | empty | Add audio and video hosts. |
| `FIFTYONE_CSP_EXTRA_CONNECT_SOURCES` | empty | Add hosts for fetches and WebSockets. |
| `FIFTYONE_CSP_EXTRA_FRAME_SOURCES` | empty | Add hosts the App may embed in iframes. |
| `FIFTYONE_CSP_EXTRA_FRAME_ANCESTORS` | empty | Add sites that may embed the App. |
| `FIFTYONE_CSP_REPORT_ONLY` | `false` | Report violations without blocking. |

### Source Format

The `FIFTYONE_CSP_EXTRA_*` variables take hosts separated by spaces or commas.
Each host may include a scheme, a leading `*.` wildcard, a port and a path:

```text
https://cdn.example.com
cdn.example.com
https://*.example.com
https://cdn.example.com:8443
https://cdn.example.com/assets/
```

- A path matches as a prefix:
  `https://cdn.example.com/assets` allows everything under `/assets/`.
- `FIFTYONE_CSP_EXTRA_SCRIPT_SOURCES` and `FIFTYONE_CSP_EXTRA_FRAME_ANCESTORS`
  accept hosts only.
  A bare scheme such as `https:` is ignored.
- The other variables also accept a bare scheme such as `https:`.
- Anything that is not a host or scheme, such as a keyword like `'unsafe-inline'`
  or a second directive, is ignored.
  The App logs a warning with the number of ignored values.

## Custom Plugins

A plugin that loads a script from a CDN needs that host in
`FIFTYONE_CSP_EXTRA_SCRIPT_SOURCES`.
A plugin that shows an iframe needs the iframe's host in
`FIFTYONE_CSP_EXTRA_FRAME_SOURCES`.
Stylesheets and fonts from any `https` host are already allowed, and most
plugins that only call their own backend need nothing.

In Helm deployments, add the variables to your `values.yaml` like

```yaml
teamsAppSettings:
  env:
    FIFTYONE_CSP_EXTRA_SCRIPT_SOURCES: https://cdn.example.com/plugin/
    FIFTYONE_CSP_EXTRA_FRAME_SOURCES: https://dashboards.example.com
```

In Docker Compose deployments, add them to your `compose.override.yaml` like

```yaml
services:
  teams-app:
    environment:
      FIFTYONE_CSP_EXTRA_SCRIPT_SOURCES: https://cdn.example.com/plugin/
      FIFTYONE_CSP_EXTRA_FRAME_SOURCES: https://dashboards.example.com
```

## Embedding the App in Another Site

By default, only the App's own origin may embed the App in an iframe.
To allow another site, add its host to `FIFTYONE_CSP_EXTRA_FRAME_ANCESTORS`.
List every site that embeds the App, separated by spaces or commas.

```yaml
teamsAppSettings:
  env:
    FIFTYONE_CSP_EXTRA_FRAME_ANCESTORS: https://portal.example.com
```

```yaml
services:
  teams-app:
    environment:
      FIFTYONE_CSP_EXTRA_FRAME_ANCESTORS: https://portal.example.com
```

## Restricting Media Hosts

By default, the App loads media and makes fetches to any `https` host
(and any `http` host when `APP_USE_HTTPS` is not `true`),
because each deployment's media hosts are its own.

To limit these to Google Cloud Storage, Amazon S3 and Azure Blob Storage,
set `FIFTYONE_CSP_MEDIA_ANY_HOST` to `false`.
Then add any other media host, such as a MinIO endpoint or a CDN,
with `FIFTYONE_CSP_EXTRA_SOURCES`.

```yaml
teamsAppSettings:
  env:
    FIFTYONE_CSP_MEDIA_ANY_HOST: "false"
    FIFTYONE_CSP_EXTRA_SOURCES: https://minio.example.com https://media.example.com/assets/
```

```yaml
services:
  teams-app:
    environment:
      FIFTYONE_CSP_MEDIA_ANY_HOST: "false"
      FIFTYONE_CSP_EXTRA_SOURCES: https://minio.example.com https://media.example.com/assets/
```

Use `FIFTYONE_CSP_EXTRA_IMG_SOURCES`, `FIFTYONE_CSP_EXTRA_MEDIA_SOURCES` or
`FIFTYONE_CSP_EXTRA_CONNECT_SOURCES` to add a host to one kind of request only.

## Diagnosing a Blocked Resource

When a page misbehaves after an upgrade, open your browser's developer
console.
The browser logs each blocked request as a Content Security Policy violation
and names the blocked host and the directive that refused it.
The App also logs each violation as a warning in the `teams-app` container log.

To find everything a deployment needs without breaking users in the meantime,
set `FIFTYONE_CSP_REPORT_ONLY` to `true`.
The App then sends the same policy as `Content-Security-Policy-Report-Only`:
the browser reports violations but blocks nothing.

```yaml
teamsAppSettings:
  env:
    FIFTYONE_CSP_REPORT_ONLY: "true"
```

```yaml
services:
  teams-app:
    environment:
      FIFTYONE_CSP_REPORT_ONLY: "true"
```

Add each reported host to the matching `FIFTYONE_CSP_EXTRA_*` variable,
then remove `FIFTYONE_CSP_REPORT_ONLY` to enforce the policy again.
Report-only mode is for diagnosis; it leaves the App without the protection of
an enforced policy.
