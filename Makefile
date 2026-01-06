SERVER_BIN_NAME=server


# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-35s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


.PHONY: generate
generate: ### generate grpc related files
	@easyp generate


.PHONY: bin-server
bin-server: ### build grpc server
	$(info build $(SERVER_BIN_NAME) ...)
	@go build -o $(SERVER_BIN_NAME) ./cmd/server


.PHONY: run-server
run-server: bin-server ### run grpc server
	$(info run $(SERVER_BIN_NAME) ...)
	@./$(SERVER_BIN_NAME)
