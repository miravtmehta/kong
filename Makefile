APP_NAME=kong
UNAME := $(shell uname -s)

## ----------------------------------------------------------------------
## This is a help comment. The purpose of this Makefile is to demonstrate
## a simple help mechanism that uses comments defined alongside the rules
## ----------------------------------------------------------------------

help: ## Show help for each of the Makefile recipes.
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)


clean: ## Removes go executables and  postgres container
	go clean
	rm -rf ${APP_NAME}-darwin
	rm -rf ${APP_NAME}-linux
	docker rm -f ${APP_NAME}-postgres

build: ## Runs clean and setup to build infra and dependencies and create go executables
	$(MAKE) clean
	$(MAKE) setup
	go mod tidy
	go mod download
	GOARCH=amd64 GOOS=darwin go build -o ${APP_NAME}-darwin .
	GOARCH=amd64 GOOS=linux go build -o ${APP_NAME}-linux .

setup: ## Pull latest Postgres images and run as container @ port 5432
	docker pull postgres:latest
	docker run --name ${APP_NAME}-postgres -e POSTGRES_PASSWORD=passo -p 5432:5432 -d postgres

run-mac: ## Run go executable for Mac
	./kong-darwin

run-linux: ## Run go executable for linux
	./kong-linux

