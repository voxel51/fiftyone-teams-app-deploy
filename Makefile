SHELL := $(SHELL) -e

# Help
.PHONY: $(shell sed -n -e '/^$$/ { n ; /^[^ .\#][^ ]*:/ { s/:.*$$// ; p ; } ; }' $(MAKEFILE_LIST))

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

dependencies: asdf

asdf:  ## Update plugins, add plugins, install plugins, set local, reshim
	@echo "Updating asdf plugins..."
	@asdf plugin update --all >/dev/null 2>&1 || true

	@echo "Adding asdf plugins..."
	@cut -d" " -f1 .tool-versions | xargs -I{} asdf plugin add {} >/dev/null 2>&1 || true

	@echo "Installing asdf tools..."
	@cat .tool-versions | xargs -I{} bash -c 'asdf install {}'

	@echo "Setting local package versions..."
	@cat .tool-versions | xargs -I{} bash -c 'asdf local {}'

	@echo "Reshimming.."
	@asdf reshim


hooks:  ## Install git hooks (pre-commit)
	@pre-commit install
	# disabled until we adopt conventional-commits
	# @pre-commit install --hook-type commit-msg

	# Install environments for all available hooks now (rather than when they are first executed)
	@pre-commit install --install-hooks

pre-commit:  ## Run pre-commit against all files
	@pre-commit run -a

start:   ## Run minikube with ingress and gcp-auth
	# to persist mongodb data, we may want to start minikube with a volume mount
	# minikube start --mount=true \
	#   --mount-string=/var/tmp/mongodb_data:/tmp/hostpath-provisioner/fiftyone-teams/mongodb
	minikube start
	minikube addons enable ingress

	# Requires setting up GCP credentials (application default credentials)
	# for the GCP project `computer-vision-team`.
	# Then run
	#
	# ```shell
	# gcloud auth application-default login
	# ```
	#
	minikube addons enable gcp-auth

	# registery-creds is an alternative methods for accessing private repositories.
	# If used, needs to be reconfired every time minikube is deleted.
	# minikube addons configure registry-creds

	# create the regcred secret to allow pulling images from dockerhub
	# kubectl create namespace fiftyone-teams --context minikube
	# kubectl --namespace fiftyone-teams \
	#   --context minikube \
	#   create secret generic regcred \
	#   --from-file=.dockerconfigjson=/var/tmp/voxel51-docker.json \
	#   --type kubernetes.io/dockerconfigjson

stop:  ## Stop minikube
	minikube stop

delete:  ## Delete minikube
	minikube delete

dev:  ## run skaffold dev
	skaffold dev

dev-keep:  ## run skaffold dev with keep-runining-on-failure
	skaffold dev --keep-running-on-failure

port-forward-app:  ## port forward the service `teams-app` on the host port 3000
	kubectl port-forward --namespace fiftyone-teams svc/teams-app 3000:80 --context minikube

port-forward-api:  ## port forward to service `teams-api` on the host port 8000
	kubectl port-forward --namespace fiftyone-teams svc/teams-api 8000:80 --context minikube

port-forward-mongo:  ## port forward to service `mongodb` on the host port 27017
	kubectl port-forward --namespace fiftyone-teams svc/mongodb 27017:27017 --context minikube
