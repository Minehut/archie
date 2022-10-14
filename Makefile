PWD := $(shell pwd)
VERSION := $(shell date +%Y%m%d%H%M%S)

GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
GOPATH := $(shell go env GOPATH)

# If the first argument is "run" pass args
ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

run:
	@echo "Running archie"
	@CGO_ENABLED=0 go run -a -ldflags "-s -w -X archie/archie.Version=dev-$(VERSION)" . $(RUN_ARGS)

build-go-local:
	@echo "Building archie binary in './dist/archie'"
	@rm -rf ./dist
	@CGO_ENABLED=0 go build -a -o ./dist/archie -ldflags "-s -w -X archie/archie.Version=dev-$(VERSION)" .

build-go-linux-arm64:
	@echo "Building archie binary in './dist/archie'"
	@rm -rf ./dist
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a -o ./dist/archie -ldflags "-s -w -X archie/archie.Version=dev-$(VERSION)" .

build-go-linux-amd64:
	@echo "Building archie binary in './dist/archie'"
	@rm -rf ./dist
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o ./dist/archie -ldflags "-s -w -X archie/archie.Version=dev-$(VERSION)" .

docker-linux-arm64: build-go-linux-arm64
	@docker build --pull -f Dockerfile.dev --build-arg "arch=arm64/v8"  -t "archie:latest" .

docker-linux-amd64: build-go-linux-amd64
	@docker build --pull -f Dockerfile.dev --build-arg "arch=amd64" -t "archie:latest" .

build-goreleaser:
	@echo "Building archie binary in './dist/archie"
	@goreleaser build --single-target --skip-validate --rm-dist

install-go: build-go-local
	@echo "Installing archie binary in '$(GOPATH)/bin/archie'"
	@mkdir -p $(GOPATH)/bin && cp -f $(PWD)/archie $(GOPATH)/bin/archie

install-goreleaser: build-goreleaser
	@echo "Installing archie binary in '$(GOPATH)/bin/archie'"
	@mkdir -p $(GOPATH)/bin && cp -f $(PWD)/dist/archie_$(GOOS)_$(GOARCH)/archie $(GOPATH)/bin/archie

release-snapshot:
	@goreleaser release --rm-dist --snapshot

release-skip-publish:
	@goreleaser release --rm-dist --skip-validate --skip-publish

release:
	@goreleaser release --rm-dist --skip-validate

helm-install:
	@helm plugin install --version master https://github.com/sonatype-nexus-community/helm-nexus-push.git || \
	helm repo add nexus https://packages.slgg.io/repository/helm-hosted --username "$$USER@superleague.com"

helm-release:
	@$(eval PASSWORD=$(shell bash -c 'read -s -p "Nexus SSO password for $$USER@superleague.com: " passwd; echo $$passwd'))
	@echo
	@cd helm && helm nexus-push nexus ./archie -u "$$USER@superleague.com" -p "$(PASSWORD)"
