SHELL := $(SHELL) -e
ASDF := $(shell asdf where golang)
VERSION ?= 1.7.0

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

auth:
	gcloud auth application-default login --project computer-vision-team
	gcloud auth configure-docker us-central1-docker.pkg.dev

hooks:  ## Install git hooks (pre-commit)
	@pre-commit install
	# disabled until we adopt conventional-commits
	# @pre-commit install --hook-type commit-msg

	# Install environments for all available hooks now (rather than when they are first executed)
	@pre-commit install --install-hooks

pre-commit:  ## Run pre-commit against all files
	@pre-commit run -a

start:  ## Run minikube with ingress and gcp-auth
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

dev: helm-repos  ## run skaffold dev
	skaffold dev

dev-keep: helm-repos  ## run skaffold dev with keep-runining-on-failure
	skaffold dev --keep-running-on-failure

port-forward-app:  ## port forward the service `teams-app` on the host port 3000
	kubectl port-forward --namespace fiftyone-teams svc/teams-app 3000:80 --context minikube

port-forward-api:  ## port forward to service `teams-api` on the host port 8000
	kubectl port-forward --namespace fiftyone-teams svc/teams-api 8000:80 --context minikube

port-forward-mongo:  ## port forward to service `mongodb` on the host port 27017
	kubectl port-forward --namespace fiftyone-teams svc/mongodb 27017:27017 --context minikube

run: helm-repos  ## run skaffold run
	skaffold run

run-cert-manager: helm-repos  ## run skaffold run
	skaffold run \
	  --filename skaffold-cert-manager.yaml

run-mongodb: helm-repos  ## run skaffold run
	skaffold run \
	  --filename skaffold-mongodb.yaml

run-profile-only-fiftyone: helm-repos  ## run skaffold run -p only-fiftyone
	skaffold run -p only-fiftyone

tunnel:  ## run minikube tunnel to access the k8s ingress via localhost ()
	minikube tunnel

helm-repos:  ## add helm repos for the project
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add jetstack https://charts.jetstack.io

clean: clean-unit-compose clean-unit-helm clean-integration-compose clean-integration-helm  ## delete all test output and reports

clean-integration-compose:  ## delete docker compose integration test output and reports
	rm -rf tests/integration/compose/test_output_internal || true
	rm -rf tests/integration/compose/test_output_legacy || true
	rm tests/integration/compose/test_output.log || true

clean-integration-helm:  ## delete helm integration test output and reports
	rm -rf tests/integration/helm/test_output_internal || true
	rm -rf tests/integration/helm/test_output_legacy || true
	rm tests/integration/helm/test_output.log || true

clean-unit-compose:  ## delete docker compose unit test output and reports
	rm -rf tests/unit/compose/test_output || true
	rm tests/unit/compose/test_output.log || true

clean-unit-helm:  ## delete helm unit test output and reports
	rm -rf tests/unit/helm/test_output || true
	rm tests/unit/helm/test_output.log || true

dependencies-integration-compose:  ## create a (temporary) directory for mongodb container
	mkdir -p /tmp/mongodb

login:  ## Docker login to Google Artifact Registry (for accessing internal gcr.io container images)
	gcloud auth print-access-token | \
	  docker login -u oauth2accesstoken \
	    --password-stdin https://us-central1-docker.pkg.dev

test-unit-compose:  ## run go test on the tests/unit/compose directory
	@cd tests/unit/compose; \
	go test -count=1 -timeout=10m -v -tags unit

test-unit-helm:  ## run go test on the tests/unit/helm directory
	@cd tests/unit/helm; \
	go test -count=1 -timeout=10m -v -tags unit

test-unit-compose-interleaved: install-terratest-log-parser  ## run go test on the tests/unit/compose directory and run the terratest_log_parser for reports
	@cd tests/unit/compose; \
	rm -rf test_output/*; \
	go test -count=1 -timeout=10m -v -tags unit | tee test_output.log; \
	${ASDF}/packages/bin/terratest_log_parser -testlog test_output.log -outputdir test_output

test-unit-helm-interleaved: install-terratest-log-parser  ## run go test on the tests/unit/helm directory and run the terratest_log_parser for reports
	@cd tests/unit/helm; \
	rm -rf test_output/*; \
	go test -count=1 -timeout=10m -v -tags unit | tee test_output.log; \
	${ASDF}/packages/bin/terratest_log_parser -testlog test_output.log -outputdir test_output

test-integration-compose: test-integration-compose-internal test-integration-compose-legacy ## run go test on the tests/integration/compose directory for both internal and legacy auth modes

test-integration-compose-internal: dependencies-integration-compose ## run go test on the tests/integration/compose directory for internal auth mode
	@cd tests/integration/compose; \
	go test -count=1 -timeout=10m -v -tags integrationComposeInternalAuth

test-integration-compose-legacy: dependencies-integration-compose ## run go test on the tests/integration/compose directory for legacy auth mode
	@cd tests/integration/compose; \
	go test -count=1 -timeout=10m -v -tags integrationComposeLegacyAuth

test-integration-compose-interleaved:  test-integration-compose-interleaved-internal test-integration-compose-interleaved-legacy  ## run go test on the tests/integration/compose directory and run the terratest_log_parser for reports

test-integration-compose-interleaved-internal: install-terratest-log-parser dependencies-integration-compose clean-integration-compose ## run go test on the tests/integration/compose directory for internal auth mode and run the terratest_log_parser for reports
	@cd tests/integration/compose; \
	rm -rf test_output_internal/*; \
	go test -count=1 -timeout=10m -v -tags integrationComposeInternalAuth | tee test_output_internal.log; \
	${ASDF}/packages/bin/terratest_log_parser -testlog test_output_internal.log -outputdir test_output_internal

test-integration-compose-interleaved-legacy: install-terratest-log-parser dependencies-integration-compose clean-integration-compose ## run go test on the tests/integration/compose directory for legacy auth mode and run the terratest_log_parser for reports
	@cd tests/integration/compose; \
	rm -rf test_output_legacy/*; \
	go test -count=1 -timeout=10m -v -tags integrationComposeLegacyAuth | tee test_output_legacy.log; \
	${ASDF}/packages/bin/terratest_log_parser -testlog test_output_legacy.log -outputdir test_output_legacy

test-integration-helm: test-integration-helm-internal test-integration-helm-legacy ## run go test on the tests/integration/helm directory for both internal and legacy auth modes

test-integration-helm-internal:  ## run go test on the tests/integration/helm directory for internal auth mode
	@cd tests/integration/helm; \
	go test -count=1 -timeout=10m -v -tags integrationHelmInternalAuth

test-integration-helm-legacy:  ## run go test on the tests/integration/helm directory for legacy auth mode
	@cd tests/integration/helm; \
	go test -count=1 -timeout=10m -v -tags integrationHelmLegacyAuth

test-integration-helm-interleaved-internal:  ## run go test on the tests/integration/helm directory for internal auth mode
	@cd tests/integration/helm; \
	rm -rf test_output_internal/*; \
	go test -count=1 -timeout=10m -v -tags integrationHelmInternalAuth | tee test_output_internal.log; \
	${ASDF}/packages/bin/terratest_log_parser -testlog test_output_internal.log -outputdir test_output_internal

test-integration-helm-interleaved-legacy:  ## run go test on the tests/integration/helm directory for legacy auth mode
	@cd tests/integration/helm; \
	rm -rf test_output_legacy/*; \
	go test -count=1 -timeout=10m -v -tags integrationHelmLegacyAuth | tee test_output_legacy.log; \
	${ASDF}/packages/bin/terratest_log_parser -testlog test_output_legacy.log -outputdir test_output_legacy

install-terratest-log-parser:  ## install terratest_log_parser
	go install github.com/gruntwork-io/terratest/cmd/terratest_log_parser@latest

get-image-versions:  ## display the latest internal image matching version string
	./utils/get-image-versions.sh "${VERSION}"

get-image-versions-dev:  ## display the latest internal image matching version string
	./utils/get-image-versions.sh "${VERSION}" dev

get-image-versions-rc:  ## display the latest internal image matching version string
	./utils/get-image-versions.sh "${VERSION}" rc
