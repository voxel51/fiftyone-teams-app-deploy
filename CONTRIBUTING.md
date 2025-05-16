# Contributing

<!-- toc -->

- [General](#general)
- [Quickstart to Skaffold in Minikube](#quickstart-to-skaffold-in-minikube)
- [pre-commit Hooks](#pre-commit-hooks)
- [Localized Deployments (for internal-testing)](#localized-deployments-for-internal-testing)
  - [minikube](#minikube)
  - [skaffold](#skaffold)
    - [profiles](#profiles)
    - [Container Images Stored in Private Repositories](#container-images-stored-in-private-repositories)
      - [Google Artifact Repository](#google-artifact-repository)
      - [Docker Hub](#docker-hub)
  - [Accessing the k8s Resources](#accessing-the-k8s-resources)
    - [Ingress](#ingress)
    - [Ingress - Login](#ingress---login)
    - [Port Forward to the `teams-app` Service](#port-forward-to-the-teams-app-service)
    - [Port Forward to the `teams-api` Service](#port-forward-to-the-teams-api-service)
    - [Port Forward - Login](#port-forward---login)

<!-- tocstop -->

## General

1. Install tool dependencies
    1. Install
       [asdf](https://asdf-vm.com/)
    1. Install tools managed by `asdf`

        ```shell
        make asdf
        ```

## Quickstart to Skaffold in Minikube

1. Auth with gcloud for the project `computer-vision-team`

    ```shell
    make auth
    ```

1. Install the asdf tools

    ```shell
    make asdf
    ```

1. In one terminal, start minikube

    ```shell
    make start
    ```

1. Download dev license file (legacy or internal)

    ```shell
    make license-secret-legacy
    ```

    or

    ```shell
    make license-secret-internal
    ```

1. Run skaffold

    ```shell
    make dev-keep
    ```

    Cancelling this process will destroy the cert-manager and mongodb deployments.
    Alternatively, they can be started with

    ```shell
    make run-cert-manager run-mongodb
    ```

    and then run this to manage the fiftyone-teams deployment resources

    ```shell
    skaffold dev \
      --profile only-fiftyone \
      --keep-running-on-failure \
      --kube-context minikube
    ```

    Cancelling this process will destroy only the fiftyone-teams deployment
    resources (leaving the cert-manager and mongodb resources).

1. In another terminal, run minikube tunnel (and provide your password when prompted)

    ```shell
    sudo minikube tunnel
    ```

    > **NOTE**: This command will prompt for sudo permission
    > on systems where 80 and 443 are privileged ports

1. Navigate to
   [https://local.fiftyone.ai](https://local.fiftyone.ai)
   and login

## pre-commit Hooks

Our Helm Chart's README.md is automatically
generated using the pre-commit hooks for

- [https://github.com/norwoodj/helm-docs](https://github.com/norwoodj/helm-docs)
- [https://github.com/Lucas-C/pre-commit-hooks-nodejs](https://github.com/Lucas-C/pre-commit-hooks-nodejs)

1. Install the pre-commit hooks

    ```shell
    make hooks
    ```

1. Update the Go Template
  [helm/fiftyone-teams-app/README.md.gotmpl](./helm/fiftyone-teams-app/README.md.gotmpl).
1. To render
  [helm/fiftyone-teams-app/README.md](./helm/fiftyone-teams-app/README.md)
    - Add the changed file `helm/fiftyone-teams-app/README.md.gotmpl`
    - Either
      - Commit the changes and let the hooks render from the template

          ```shell
          [fiftyone-teams-app-deploy]$ git add helm/fiftyone-teams-app/README.md.gotmpl
          [fiftyone-teams-app-deploy]$ git commit -m 'adding new section'
          check for added large files...........................................Passed
          check for case conflicts..............................................Passed
          check that scripts with shebangs are executable.......................Passed
          check yaml........................................(no files to check)Skipped
          detect aws credentials................................................Passed
          fix end of files......................................................Passed
          mixed line ending.....................................................Passed
          pretty format json................................(no files to check)Skipped
          trim trailing whitespace..............................................Passed
          No-tabs checker.......................................................Passed
          markdownlint......................................(no files to check)Skipped
          markdownlint-fix..................................(no files to check)Skipped
          codespell.............................................................Passed
          yamllint..........................................(no files to check)Skipped
          Helm Docs.............................................................Failed
          - hook id: helm-docs
          - files were modified by this hook

          INFO[2023-11-09T16:11:14-07:00] Found Chart directories [.]
          INFO[2023-11-09T16:11:14-07:00] Generating README Documentation for chart helm/fiftyone-teams-app

          Insert a table of contents in Markdown files, like a README.md........Passed
          [fiftyone-teams-app-deploy]$ git add helm/fiftyone-teams-app/README.md
          [fiftyone-teams-app-deploy]$ git commit -m 'adding new section'
          check for added large files...........................................Passed
          check for case conflicts..............................................Passed
          check that scripts with shebangs are executable.......................Passed
          check yaml........................................(no files to check)Skipped
          detect aws credentials................................................Passed
          fix end of files......................................................Passed
          mixed line ending.....................................................Passed
          pretty format json................................(no files to check)Skipped
          trim trailing whitespace..............................................Passed
          No-tabs checker.......................................................Passed
          markdownlint..........................................................Passed
          markdownlint-fix......................................................Passed
          codespell.............................................................Passed
          yamllint..........................................(no files to check)Skipped
          Helm Docs.............................................................Passed
          Insert a table of contents in Markdown files, like a README.md.......................Passed
          [AS-22-helm-docs a81c21b] adding new section
          2 files changed, 10 insertions(+)
          ```

      - Manually run the pre-commit hooks

          ```shell
          git add helm/fiftyone-teams-app/README.md.gotmpl
          pre-commit run helm-docs
          pre-commit run markdown-toc
          git add helm/fiftyone-teams-app/README.md
          git commit -m '<COMMIT_MESSAGE>'
          ```

## Localized Deployments (for internal-testing)

1. Install additional dependencies

    - Install
      [Docker](https://www.docker.com/products/docker-desktop/)

1. Add the helm repos

    ```shell
    make helm-repos
    ```

### minikube

[minikube](https://minikube.sigs.k8s.io/docs/)
provides a local kubernetes cluster
in VMs (or docker containers) on macOS, Linux and Windows.

```shell
minikube start
```

### skaffold

We use
[Skaffold](https://skaffold.dev/)
to deploy our application to the local minikube cluster with
Helm and overrides (`values.yaml`).

The license file contains the secrets.
Copy the license file for our local dev organization.

For legacy CAS mode

```shell
make license-secret-legacy
```

For internal CAS mode

```shell
make license-secret-internal
```

When debugging, it may be helpful to start minikube with the flag
`--keep-running-on-failure` so that the k8s resources are not deleted
if the helm installation(s) fail.

```shell
skaffold dev --keep-running-on-failure
```

It takes a few minutes for the deployments to stabilize as
we wait for Helm to install MongoDB and cert-managed (for self-signed certificates).
The fiftyone-teams app installation also takes a few minutes.
The fiftyone-app will start and upgrade the database
and the teams-api will connect to and configure MongoDB.

We use Skaffold "profiles" to control "modules".
By default, Skaffold will Helm install

- MongoDB
- cert-manager
  - CRDs
  - self-singed ClusterIssuer
  - cert-manager from chart defaults
- FiftyOne Enterprise License
- FiftyOne Enterprise

#### profiles

To skip installing MongoDB, run

```shell
skaffold dev --profile no-mongodb
```

To skip installing cert-manager, run

```shell
skaffold dev ---profile no-cert-manager
```

To skip installing both MongoDB and cert-manager, run

```shell
skaffold dev --profile only-fiftyone
```

#### Container Images Stored in Private Repositories

Our FiftyOne Enterprise container images are stored in the private repositories

- [Google Artifact Repository (Docker)](#google-artifact-repository)
  - Contains private development images created by our private repository
    [Google Cloud Build](https://github.com/voxel51/cloud-build-and-deploy/)
    CI/CD runs
    - Development
    - Release Candidates
- [Docker Hub](#docker-hub)
  - Contains released versions

Accessing images in a private repository requires setting
up authentication to that container registry.

##### Google Artifact Repository

To run released images from Google Artifact repository in the
GCP project `computer-vision-team`, configure minikube and skaffold

1. Configure GCP Credentials
   [gcloud auth](https://cloud.google.com/sdk/gcloud/reference/auth)
1. Configure
   [gcloud auth application-default](https://cloud.google.com/sdk/gcloud/reference/auth/application-default)

1. Start minikube and enable the addon `gcp-auth`

    ```shell
    minikube start
    minikube addons enable gcp-auth
    ```

1. In
   [skaffold.yaml](./skaffold.yaml)
   comment `imagePullSecrets` for the helm release named `fiftyone-teams-app`
   in `setValueTemplates.imagePullSecrets[0].name=regcred`

    ```yaml
    deploy:
      helm:
        releases:
          - name: fiftyone-teams-app
            setValueTemplates:
              # imagePullSecrets:
              #   - name: regcred
    ```

1. To use an image different than the Helm Chart Version,
   update the corresponding `image.tag` value.
   For each service

    - `apiSettings`
    - `appSettings`
    - `casSettings`
    - `pluginsSettings`
    - `teamsAppSettings`

   For example for the version `2.0.0` at the latest `rc`s.

    ```yaml
    apiSettings:
      image:
        repository: us-central1-docker.pkg.dev/computer-vision-team/dev-docker/fiftyone-teams-api
        tag: v2.0.0rc17
    appSettings:
      image:
        repository: us-central1-docker.pkg.dev/computer-vision-team/dev-docker/fiftyone-app
        tag: v2.0.0rc17
    casSettings:
      image:
        repository: us-central1-docker.pkg.dev/computer-vision-team/dev-docker/fiftyone-teams-cas
        tag: v2.0.0-rc.16
    pluginsSettings:
      image:
        repository: us-central1-docker.pkg.dev/computer-vision-team/dev-docker/fiftyone-app
        tag: v2.0.0rc17
    teamsAppSettings:
      image:
        repository: us-central1-docker.pkg.dev/computer-vision-team/dev-docker/fiftyone-teams-app
        # Note: the naming convention for the image `fiftyone-teams-app` differs from
        # the other images (`fiftyone-app`, `fiftyone-app` and `fiftyone-teams-api`).
        # The others are `vW.X.Y.devZ` (note `.devZ` vs `-dev.Z`).
        # This is a byproduct of `npm` versioning versus Python PEP 440.
        tag: v2.0.0-rc.16
    ```

    > _Note:_ To see the available tags for each image, see
    > [https://console.cloud.google.com/artifacts/docker/computer-vision-team/us-central1/dev-docker?project=computer-vision-team](https://console.cloud.google.com/artifacts/docker/computer-vision-team/us-central1/dev-docker?project=computer-vision-team)

1. Run skaffold

    ```shell
    skaffold dev

    # Or with the optional flag
    # skaffold dev --keep-running-on-failure
    ```

##### Docker Hub

> _Note:_ Release Artifacts are available in the Google Artifact Registry.
> To obtain a Docker Hub Private Access Token,
> contact your friendly neighborhood Aloha Shirt.

To run released images from Docker hub, configure minikube and Skaffold

1. Start minikube and enable the addon `registry-creds`

    ```shell
    minikube start
    minikube addons configure registry-creds
    ```

1. Create the file `voxel51-docker.json` file
    1. Get base64 encoded string of docker username and
       Docker Personal Access Token (PAT)

        ```shell
        echo -n 'voxeldocker:<YOUR_DOCKER_PERSONAL_ACCESS_TOKEN>' | base64
        ```

    1. Using this template, add replace the `<BASE64_ENCODED_STRING_OF_DOCKER_USERNAME_COLON_PAT>`
       with the output from the previous step

        ```json
        {
          "auths": {
            "https://index.docker.io/v1/": {
              "auth": "<BASE64_ENCODED_STRING_OF_DOCKER_USERNAME_COLON_PAT>",
              "email": "docker@voxel51.com"
            }
          }
        }
        ```

1. Create the Kubernetes namespace configured in
   [skaffold.yaml](./skaffold.yaml)

    ```shell
    export NAMESPACE=fiftyone-teams
    kubectl create namespace "${NAMESPACE}"
    ```

1. Create the imagePullSecret named `regcred`

    ```shell
    kubectl create secret generic regcred \
      --from-file=.dockerconfigjson=/var/tmp/voxel51-docker.json \
      --type kubernetes.io/dockerconfigjson \
      --namespace "${NAMESPACE}" \
    ```

1. In
   [skaffold.yaml](./skaffold.yaml)
   set `imagePullSecrets` for the helm release named `fiftyone-teams-app`
   in `setValueTemplates.imagePullSecrets[0].name=regcred`

    ```yaml
    deploy:
      helm:
        releases:
          - name: fiftyone-teams-app
            setValueTemplates:
              imagePullSecrets:
                - name: regcred
    ```

1. Run skaffold

    ```shell
    skaffold dev
    ```

For more information, see the Kubernetes documentation
[Pull an Image from a Private Registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

> _Note:_ After running `minikube delete`, the secret `regcred` must be recreated.

### Accessing the k8s Resources

There are two ways to access resources within the minikube cluster:

- Ingress (recommended)
- Port Forward

#### Ingress

See
[Using `minikube tunnel`](https://minikube.sigs.k8s.io/docs/handbook/accessing/#using-minikube-tunnel).

1. Enable the minikube addon `ingress`

    ```shell
    minikube addons enable ingress
    ```

1. Install the app via skaffold (see above)
1. Start the minikube tunnel (and provide sudo password when prompted)

    ```shell
    $ minikube tunnel
    ✅  Tunnel successfully started

    📌  NOTE: Please do not close this terminal as this process must stay alive for the tunnel to be accessible ...

    ❗  The service/ingress fiftyone-teams-fiftyone-teams-app requires privileged ports to be exposed: [80 443]
    🔑  sudo permission will be asked for it.
    🏃  Starting tunnel for service fiftyone-teams-fiftyone-teams-app.
    Password:
    ```

#### Ingress - Login

This section assumes the use of TLS certificates and the `https` protocol.

1. In a web browser and navigate to

    1. Select "Continue with Voxel51 Internal"

1. In a web browser, navigate to
   [https://local.fiftyone.ai](https://local.fiftyone.ai)
1. Login with `Continue with Voxel51 Internal`
1. After authentication, you will be redirected to
   [https://local.fiftyone.ai/datasets](https://local.fiftyone.ai/datasets)

> _Note:_ For local development with, we use the
> Auth0 Tenant `dev-fiftyone` and the Auth0 Application `local-dev`.
> The `local-dev` app contains the setting Allowed Callback URLs
> (aka Redirect URLs) with
> [https://local.fiftyone.ai](https://local.fiftyone.ai)
> .
> In `skaffold.yaml`, in both `appSettings.env` and `teamsAppSettings.env`,
> either omit `APP_USE_HTTPS=false` or set `APP_USE_HTTPS=true`
> for the app to set the Redirect URL's protocol to `https`.

#### Port Forward to the `teams-app` Service

To access the teams-app webpage, run a kubernetes port forward
(to forward traffic from the host's port) to the kubernetes service `teams-app`.
Afterwards, access the FiftyOne Enterprise app via
[http://localhost:3000](http://localhost:3000).

1. Initiate the port forward to the service `teams-app` on port 3000

    ```shell
    kubectl port-forward \
      --namespace fiftyone-teams \
      svc/teams-app 3000:80
    ```

1. Validate port forwarding is working

    ```shell
    $ curl http://localhost:3000/api/hello
    {"name":"John Doe"}
    ```

#### Port Forward to the `teams-api` Service

1. Initiate the port forward to the service `team-api` on port 8000

    ```shell
    kubectl port-forward --namespace fiftyone-teams svc/teams-api 8000:80
    ```

1. Validate port forwarding is working

    ```shell
    $ curl http://localhost:8000/health
    {"status":"available"}
    ```

#### Port Forward - Login

With the port forward running,

1. In a web browser, navigate to
   [http://localhost:3000](http://localhost:3000)
1. Login with `Continue with Voxel51 Internal`
1. After authentication, you will be redirected to
   [http://localhost:3000/datasets](http://localhost:3000/datasets)

> _Note:_ For local development, we use the Auth0 Tenant `dev-fiftyone` and
> the Auth0 Application `local-dev` contains the setting Allowed Callback URLs
> (aka Redirect URLs) with
> [http://localhost:3000](http://localhost:3000).
> In `skaffold.yaml` we set `APP_USE_HTTPS=false`
> to prohibit the app from setting the Redirect URL protocol to `https`.
> Must be set in both `appSettings.env` and `teamsAppSettings.env`.
> Without this setting, the app code makes the callback URL
> [https://localhost:3000](https://localhost:3000)
> and Auth0 throws a Callback URL mismatch error.
