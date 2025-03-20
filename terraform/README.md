# Deploying FiftyOne Enterprise Infrastructure using Terraform

Module for deploying FiftyOne Enterprise Infrastructure

## Google Compute

### Simple Two-Node Deploy

**NOTE**: This deploys a simple two-node infrastructure
that does not provide High Availability.
If High Availability is a requirement in your environment, consider modifying
this deployment to include multiple MongoDB nodes and multiple app nodes.

This terraform does not deploy the application - it only deploys two systems;
you will still need to use another mechanism to deploy the application itself.

Edit `simple-two-node-deploy/google/terraform.tfvars` setting
your Google Compute project name as the `google_project`.

Edit `simple-two-node-deploy/google/public-ssh-keys` setting any
SSH keys that will be provisioned for the nodes that are created.
(format `username:contents of public ssh keyfile`)

```shell
gcloud auth login
cd simple-two-node-deploy/google
terraform init
terraform apply
```

The terraform apply will output the IP address of each node, the users
defined in the `public-ssh-keys` file will have ssh access to the system.

## TODO

- Add a dedicated data drive for MongoDB
- Make the MongoDB Data Drive XFS formatted
- Disable Transparent Huge Pages for `fiftyone-dbnode`
- Find a elegant way to have Terraform deploy MongoDB and FiftyOne Enterprise
