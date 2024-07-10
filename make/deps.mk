DOCKER_BUILDX ?= buildx
DOCKER_BUILDX_VERSION ?= v0.12.1
DOCKER_CLI_PLUGINS ?= ~/.docker/cli-plugins

YQ_VERSION ?= v4.44.1

.PHONY: install-docker-buildx
install-docker-buildx:
	@echo "List folder of DOCKER_CLI_PLUGINS: $(DOCKER_CLI_PLUGINS)"
	ls -la $(DOCKER_CLI_PLUGINS)
	curl -s -L -f "https://github.com/docker/buildx/releases/download/$(DOCKER_BUILDX_VERSION)/buildx-$(DOCKER_BUILDX_VERSION).$(LOCAL_OS)-$(LOCAL_ARCH)" > "$(LOCALBIN)/$(DOCKER_BUILDX)"
	chmod a+x "$(DOCKER_BUILDX)"
	ln -s "$(LOCALBIN)/$(DOCKER_BUILDX)" "$(DOCKER_CLI_PLUGINS)/docker-buildx"



# yq is a lightweight and portable command-line YAML processor
.PHONY: yq
YQ ?= $(LOCALBIN)/yq
yq:
ifeq (,$(wildcard $(YQ)))
	ifeq (, $(shell which yq 2>/dev/null))
		@{ \
		set -e ;\
		mkdir -p $(dir $(YQ)) ;\
		curl -sSLo $(YQ) https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$(LOCAL_OS)_$(LOCAL_ARCH) ;\
		chmod +x $(YQ) ;\
		}
	else
		YQ = $(shell which yq)
	endif
endif