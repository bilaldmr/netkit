.DEFAULT_GOAL := help

##@ general

.PHONY: help
help: ## show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nusage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ docker

.PHONY: docker-build
docker-build: ## build the image
	@./scripts/docker.sh build

.PHONY: docker-up
docker-up: ## build and start the stack in the background
	@./scripts/docker.sh up

.PHONY: docker-down
docker-down: ## stop the stack
	@./scripts/docker.sh down

.PHONY: docker-logs
docker-logs: ## follow service logs
	@./scripts/docker.sh logs

.PHONY: docker-ps
docker-ps: ## list service status
	@./scripts/docker.sh ps

.PHONY: docker-clean
docker-clean: ## remove the stack and its volumes
	@./scripts/docker.sh clean
