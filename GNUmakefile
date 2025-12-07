TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=github.com
NAMESPACE=nautobot
NAME=nautobot
BINARY=terraform-provider-${NAME}
VERSION=3.0.0-beta
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

# Inject VERSION into main.version
LDFLAGS=-X 'main.version=$(VERSION)'

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

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

local: install
	sed -i "s|version =.*|version = \"${VERSION}\"|" test/provider.tf
	cd test; tofu init -upgrade && tofu apply -auto-approve; cd ..

docs-gen:
	sed -i "s|version =.*|version = \"${VERSION}\"|" README.md
	go generate ./...

tag: local
	git add .
	git commit -m "Bump to version ${VERSION}"
	git tag -a -m "Bump to version ${VERSION}" v${VERSION}
	git push --follow-tag
