DEV_VERSION ?= dev # Could be other string or version number
DEV_REGISTRY ?= quay.io/bedrockinstallerfid


ifneq ($(shell echo "$(DEV_VERSION)" | grep -E '^[^0-9]'),)
	TAG := $(DEV_VERSION)
else
	TAG := v$(DEV_VERSION)
endif

DEV_IMG ?= $(DEV_REGISTRY)/ibm-account-iam-operator-bundle:$(TAG)

DEV_BUNDLE_IMG ?= $(DEV_REGISTRY)/ibm-account-iam-operator-bundle:$(TAG)

DEV_CATALOG_IMG ?= $(DEV_REGISTRY)/ibm-account-iam-operator-catalog:$(TAG)

# Change the image to dev when applying deployment manifests
deploy: configure-dev

# Configure the varaiable for the dev build
.PHONY: configure-dev
configure-dev:
	$(eval VERSION := $(DEV_VERSION))
	$(eval IMG := $(DEV_IMG))
	$(eval BUNDLE_IMG := $(DEV_BUNDLE_IMG))
	$(eval CATALOG_IMG := $(DEV_CATALOG_IMG))
	
##@ Development Build
.PHONY: docker-build-dev
docker-build-dev: configure-dev docker-build 

.PHONY: docker-build-push-dev
docker-build-push-dev: docker-build-dev docker-push

.PHONY: bundle-build-dev
bundle-build-dev: configure-dev bundle-build

.PHONY: bundle-build-push-dev
bundle-build-push-dev: bundle-build-dev bundle-push

.PHONY: catalog-build-dev
catalog-build-dev: configure-dev catalog-build

.PHONY: catalog-build-push-dev
catalog-build-push-dev: catalog-build-dev catalog-push

clean-before-commit:
	cd config/manager && $(KUSTOMIZE) edit set image controller=controller:latest
	cp ./config/manager/manager.yaml ./config/manager/tmp.yaml
	sed -e 's/Always/IfNotPresent/g' ./config/manager/tmp.yaml > ./config/manager/manager.yaml
	rm ./config/manager/tmp.yaml


# Test
.PHONY: check
check: ## @code Run the code check
	@echo "Running check for the code."
	@echo "Run make install-docker-buildx"
	$(MAKE) install-docker-buildx
