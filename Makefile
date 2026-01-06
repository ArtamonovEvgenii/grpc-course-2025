-include local/local.env
export

SERVER_BIN_NAME=server
CLIENT_BIN_NAME=client


# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-35s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


.PHONY: generate-api
generate-api: ### generate grpc related files
	@easyp generate


.PHONY: lint-api
lint-api: ### lint api definition using easyp
	@easyp lint --path api


.PHONY: bin-server
bin-server: ### build grpc server
	$(info build $(SERVER_BIN_NAME) ...)
	@go build -o $(SERVER_BIN_NAME) ./cmd/server


.PHONY: run-server
run-server: bin-server ### run grpc server
	$(info run $(SERVER_BIN_NAME) ...)
	@./$(SERVER_BIN_NAME)


.PHONY: bin-client
bin-client: ### build grpc server
	$(info build $(CLIENT_BIN_NAME) ...)
	@go build -o $(CLIENT_BIN_NAME) ./cmd/client


.PHONY: run-client
run-client: bin-client ### run grpc server
	$(info run $(CLIENT_BIN_NAME) ...)
	@./$(CLIENT_BIN_NAME)
