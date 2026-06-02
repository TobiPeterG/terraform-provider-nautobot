TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=github.com
NAMESPACE=nautobot
NAME=nautobot
BINARY=terraform-provider-${NAME}
VERSION=3.0.4
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

LDFLAGS=-X 'main.version=$(VERSION)'

TF_PLUGIN_CACHE_DIR?=$(CURDIR)/.terraform.d/plugin-cache

DOCKER_COMPOSE?=docker compose
COMPOSE_NAUTOBOT_FILE?=test/nautobot/docker-compose.yml
COMPOSE_CI_FILE?=test/docker-compose.yml

NAUTOBOT_VER?=3.0.2
PYTHON_VER?=3.12

TF_TOOL?=opentofu
TF_VERSION?=1.11.1
TF_TARGET?=$(if $(filter opentofu,$(TF_TOOL)),with-opentofu,with-terraform)

.PHONY: docs test

default: install

build:
	go build -ldflags "$(LDFLAGS)" -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64  go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386   go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm   go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386     go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64   go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm     go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386   go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386   go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p $(HOME)/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} $(HOME)/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc-docker:
	@mkdir -p "$(TF_PLUGIN_CACHE_DIR)"
	@echo "Running acceptance tests via Docker Compose"
	@echo "  TF_TOOL=$(TF_TOOL) TF_VERSION=$(TF_VERSION) TF_TARGET=$(TF_TARGET)"
	@echo "  NAUTOBOT_VER=$(NAUTOBOT_VER) PYTHON_VER=$(PYTHON_VER)"
	@NAUTOBOT_VER="$(NAUTOBOT_VER)" \
	  PYTHON_VER="$(PYTHON_VER)" \
	  TF_TOOL="$(TF_TOOL)" \
	  TF_VERSION="$(TF_VERSION)" \
	  TF_TARGET="$(TF_TARGET)" \
	  $(DOCKER_COMPOSE) --project-directory . -f "$(COMPOSE_NAUTOBOT_FILE)" -f "$(COMPOSE_CI_FILE)" \
	    up --build --abort-on-container-exit --exit-code-from testrunner

testacc-docker-down:
	@NAUTOBOT_VER="$(NAUTOBOT_VER)" \
	  PYTHON_VER="$(PYTHON_VER)" \
	  TF_TOOL="$(TF_TOOL)" \
	  TF_VERSION="$(TF_VERSION)" \
	  TF_TARGET="$(TF_TARGET)" \
	  $(DOCKER_COMPOSE) --project-directory . -f "$(COMPOSE_NAUTOBOT_FILE)" -f "$(COMPOSE_CI_FILE)" \
	    down -v --remove-orphans

testacc-local:
	@if command -v tofu >/dev/null 2>&1; then \
		cli=tofu; \
	elif command -v terraform >/dev/null 2>&1; then \
		cli=terraform; \
	else \
		echo "Error: neither 'tofu' nor 'terraform' found in PATH."; \
		exit 1; \
	fi; \
	abs=$$(command -v $$cli); \
	mkdir -p "$(TF_PLUGIN_CACHE_DIR)"; \
	echo "Using $$cli ($$abs)"; \
	TF_ACC=1 \
	TF_ACC_TERRAFORM_PATH="$$abs" \
	TF_PLUGIN_CACHE_DIR="$(TF_PLUGIN_CACHE_DIR)" \
	TF_ACC_PROVIDER_NAMESPACE="hashicorp" \
	TF_ACC_PROVIDER_HOST="registry.opentofu.org" \
	go test -p 1 $(TEST) -v $(TESTARGS) -timeout 120m

testacc:
ifeq ($(CI),true)
	@$(MAKE) testacc-docker
else
	@$(MAKE) testacc-local
endif

local: install
	sed -i "s|version =.*|version = \"${VERSION}\"|" local/provider.tf
	cd local; tofu init -upgrade && tofu apply -auto-approve; cd ..

docs:
	sed -i "s|version =.*|version = \"${VERSION}\"|" README.md
	sed -i "s|version =.*|version = \"${VERSION}\"|" examples/provider/provider.tf
	sed -i "s|version =.*|version = \"${VERSION}\"|" local/provider.tf
	go generate ./...

tag: local
	git add .
	git commit -m "Bump to version ${VERSION}"
	git tag -a -m "Bump to version ${VERSION}" v${VERSION}
	git push --follow-tag
