<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>

---

# Exposing the `teams-api` service

You may wish to expose your FiftyOne Teams API for SDK access.

You can expose your `teams-api` service in any manner that suits your deployment strategy; the following are two possible solutions but do not represent the entirety of possible solutions.  Essentially any solution that allows the FiftyOne Teams SDK to use websockets to access port 8000 on the `teams-api` container should work.

**NOTE**: The `teams-api` service uses websockets to maintain connections and allow for long-running processes to complete.  Please ensure your Infrastructure supports websockets before attempting
to expose the `teams-api` service. (e.g. You will have to migrate from AWS Classic Load Balancers to AWS Application Load Balancers to provide websockets support.)

**NOTE**: If you are using file-based storage credentials, or setting environment variables, you will need to make sure the same credentials are shared with the `fiftyone-app` and `teams-api` containers.  Voxel51 recommends the use of Database Cloud Storage Credentials, which can be configured at `/settings/cloud_storage_credentials`.


## Expose `teams-api` directly

1. Edit your `.env` file and set `API_BIND_ADDRESS` to `0.0.0.0`
1. Recreate your environment using [the appropriate](./README.md#enabling-fiftyone-teams-plugins) `docker compose` command
1. Access your FiftyOne Teams API using the same hostname as your FiftyOne Teams App using port 8000

**NOTE**: This method does not protect your API endpoint with TLS and will send API Keys in clear text.  While it is the simplest mechanism, there are security considerations that you should assess before using this method.

## Expose `teams-api` using Nginx and a unique hostname

1. Create a second hostname for your API (e.g. demo-api.fiftyone.ai) and create SSL certificates for that hostname
2. Using the [example-nginx-api.conf](../example-nginx-api.conf) as a template, create a second service for your nginx reverse proxy
3. Access your FiftyOne Teams API using the new hostname
