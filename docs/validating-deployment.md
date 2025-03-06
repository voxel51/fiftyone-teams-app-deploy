<!-- markdownlint-disable no-inline-html line-length no-alt-text -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length no-alt-text -->

# Validating Your Deployment

FiftyOne enterprise comes with a `test_api_connection()` method which will
attempt to validate the connection between your SDK and your deployment.

<!-- toc -->

- [Pre-requisites](#pre-requisites)
- [Running Checks](#running-checks)

<!-- tocstop -->

## Pre-requisites

The following validation method assumes you:

1. Deployed FiftyOne enterprise in either kubernetes or
   docker-compose
1. Configured a DNS record(s) for your application
1. Configured TLS termination for your application
1. Configured
   [your authentication provider](https://docs.voxel51.com/teams/pluggable_auth.html)
1. Installed the FiftyOne Enterprise SDK
1. Generated an API Key via the Enterprise UI

## Running Checks

To run the checks, set the `FIFTYONE_API_KEY` and `FIFTYONE_API_URL`
environment variables:

```shell
export FIFTYONE_API_URL=https://your-api-url
export FIFTYONE_API_KEY=you4ap1k3y
```

Test the API connection:

```shell
python -c 'import fiftyone.management as fom; fom.test_api_connection()'
```

When successful, it will return:

```shell
$ python -c 'import fiftyone.management as fom; fom.test_api_connection()'

API connection succeeded
```

If you have issues during setup, please contact your customer success representative.
