# Validating Your Deployment

FiftyOne enterprise comes with a `test_api_connection()` method which will
attempt to validate the connection between your SDK and your deployment.

## Pre-requisites

The following validation method assumes:

1. You have deployed FiftyOne enterprise in either kubernetes or
   docker-compose.
1. You have configured a DNS record(s) for your application
1. You have configured TLS termination for your application
1. You have configured
   [your authentication provider](https://docs.voxel51.com/teams/pluggable_auth.html).
1. You have installed the FiftyOne Enterprise SDK

## Running Checks

To run the checks, ensure your `FIFTYONE_API_KEY` and `FIFTYONE_API_URL`
are set in your environment:

```shell
export FIFTYONE_API_URL=https://your-api-url
export FIFTYONE_API_KEY=you4ap1k3y
```

Then run `fiftyone.management.test_api_connection()`:

```shell
python -c 'import fiftyone.management as fom; fom.test_api_connection()'
```

If all goes well, you will see the following log:

```shell
% python -c 'import fiftyone.management as fom; fom.test_api_connection()'

API connection succeeded
```

If you have issues during setup, please contact your customer success representative.
