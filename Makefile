# The name of the executable (default is current directory name)
BINARY := mosoly-ledger-bridge
ARTIFACTS_DIR := artifacts
ASSETS_DIR = assets

VERSION ?= vlatest
COMMIT := $(shell git rev-parse HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_TIME := $(shell TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ')
M = $(shell printf "\033[32;1m▶▶▶▶▶\033[0m")

GO_LDFLAGS := "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitHash=$(COMMIT) -X main.Branch=$(BRANCH)"

TOOLS = golang.org/x/tools/cmd/goimports \
         github.com/Masterminds/glide \
         golang.org/x/lint/golint \
         honnef.co/go/tools/cmd/staticcheck \

.PHONY: all
all: clean test ; $(info $(M) building executable…) @ ## Build program binary
	mkdir -p $(ARTIFACTS_DIR)
	go build --ldflags=$(GO_LDFLAGS) -o $(ARTIFACTS_DIR)/$(BINARY)
	echo export VERSION=$(VERSION) > $(ARTIFACTS_DIR)/VERSION

.PHONY: test
test: vendor fmt-check vet static-check lint ; $(info $(M) running tests…)
	go test -timeout 20s -race -v $$(glide novendor)

.PHONY: tools
tools: ; $(info $(M) building tools…)
	go get -v $(TOOLS)

.PHONY: vendor
vendor: tools ; $(info $(M) retrieving dependencies…)
	glide install

.PHONY: fmt-check
fmt-check: vendor tools ; $(info $(M) checking formattation…)
	gofiles=$$(go list -f {{.Dir}} $$(glide novendor) | grep -v mock) && [ -z "$$gofiles" ] || unformatted=$$(for d in $$gofiles; do goimports -l $$d/*.go; done) && [ -z "$$unformatted" ] || (echo >&2 "Go files must be formatted with goimports. Following files has problem:\n$$unformatted" && false)

.PHONY: fmt
fmt: vendor tools ; $(info $(M) formatting the code…)
	gofiles=$$(go list -f {{.Dir}} $$(glide novendor) | grep -v mock) && [ -z "$$gofiles" ] || for d in $$gofiles; do goimports -l -w $$d/*.go; done

.PHONY: vet
vet: vendor tools ; $(info $(M) checking correctness of the code…)
	go vet $$(glide novendor)

.PHONY: static-check
static-check: vendor tools ; $(info $(M) detecting bugs and inefficiencies in code…)
	staticcheck $$(glide novendor)

.PHONY: lint
lint: vendor tools ; $(info $(M) running golint…)
	for i in $$(go list ./... | grep -v /vendor/); do golint $$i; done

clean:
	rm -rf $(ARTIFACTS_DIR)
