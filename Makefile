SHA=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)
DIRTY=false
GO_PACKAGE=$(shell go list)
LDFLAGS=-ldflags "-w -s -X $(GO_PACKAGE)/pkg/util.GitSha=${SHA} -X $(GO_PACKAGE)/pkg/util.Version=${VERSION} -X $(GO_PACKAGE)/pkg/util.Dirty=${DIRTY}"
export GO111MODULE=on

clean: ## clean the repo
	rm aws-oidc 2>/dev/null || true
	go clean
	go clean -testcache
	rm -rf dist 2>/dev/null || true
	rm coverage.out 2>/dev/null || true
	if [ -e /tmp/aws-oidc.lock ]; then \
        rm /tmp/aws-oidc.lock; \
    fi \

setup: # setup development dependencies
	export GO111MODULE=on
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh -s -- v0.9.14
	curl -sfL https://raw.githubusercontent.com/chanzuckerberg/bff/master/download.sh | sh
.PHONY: setup

install:
	go install
.PHONY: install

test:
	CGO_ENABLED=1 go test -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

test-all:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./... -tags=integration
.PHONY: test-all

test-coverage:  ## run the test with proper coverage reporting
	go test  -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out
.PHONY: test-coverage

test-coverage-integration:  ## run the test with proper coverage reporting
	go test -coverprofile=coverage.out -covermode=atomic ./... -tags=integration
	go tool cover -html=coverage.out
.PHONY: test-coverage-all

deps:
	go get -u ./...
	go mod tidy
.PHONY: deps

lint:
	golangci-lint run -E whitespace --exclude-use-default
.PHONY: lint

lint-ci: ## run the fast go linters
	./bin/reviewdog -conf .reviewdog.yml  -reporter=github-pr-review
.PHONY: lint-ci

release: ## run a release
	bff bump
	git push
	goreleaser release --rm-dist
.PHONY: release

release-prerelease: ## release to github as a 'pre-release'
	go build ${LDFLAGS} .
	commit=`git rev-parse --short HEAD`; \
	version=`cat VERSION`; \
	git tag v"$$version"+"$$commit"; \
	git push
	git push --tags
	goreleaser release -f .goreleaser.prerelease.yml --debug --rm-dist
.PHONY: release-prelease

fmt:
	goimports -w -d $$(find . -type f -name '*.go' -not -path "./vendor/*")
.PHONY: fmt

check-mod:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
.PHONY: check-mod
