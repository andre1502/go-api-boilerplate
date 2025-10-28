REPO_URL := $(shell git config --get remote.origin.url)
REPO_NAME := $(shell basename "$(REPO_URL)" .git)
BRANCH_NAME := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date '+%Y-%m-%dT%H:%M:%S%z')
BUILD_VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.repoName=$(REPO_NAME) -X main.branchName=$(BRANCH_NAME) -X main.commitHash=$(COMMIT_HASH) -X main.buildDate=$(BUILD_DATE) -X main.version=$(BUILD_VERSION) -s -w"
SOURCES ?= $(shell find . -name "*.go" -type f)
GO ?= go

ifneq ($(shell uname), Darwin)
	EXTLDFLAGS = -extldflags "-static" $(null)
else
	EXTLDFLAGS =
endif

# build api
build_api: $(SOURCES)
	$(GO) build -v $(LDFLAGS) -o bin/api cmd/api/main.go

# build scheduler
build_scheduler: $(SOURCES)
	$(GO) build -v $(LDFLAGS) -o bin/scheduler cmd/scheduler/main.go

# install air command
.PHONY: air
air:
	@hash air > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/air-verse/air@latest; \
	fi

# run dev
.PHONY: dev
dev: air
	air --build.cmd "make build_api" --build.bin "bin/api"

# run scheduler
.PHONY: scheduler
scheduler: air
	air --build.cmd "make build_scheduler" --build.bin "bin/scheduler"
