GIT_TAG := $(shell echo $(shell git describe --tags || git branch --show-current) | sed 's/^v//')
COMMIT  := $(shell git log -1 --format='%H')
BUILD_DATE	:= $(shell date '+%Y-%m-%d')
IS_SUDO_USER := $(shell if [ "$(shell whoami)" = "root" ] || [ "$(shell groups | grep -e 'sudo' -e 'admin' -e 'google-sudoers' | wc -l)" = "1" ]; then echo "1"; fi)
GO_BIN := $(shell echo $(shell which go || echo "/usr/local/go/bin/go" ))

###############################################################################
###                                Build flags                              ###
###############################################################################

LD_FLAGS = -X github.com/bcdevtools/node-management/constants.VERSION=$(GIT_TAG) \
            -X github.com/bcdevtools/node-management/constants.COMMIT_HASH=$(COMMIT) \
            -X github.com/bcdevtools/node-management/constants.BUILD_DATE=$(BUILD_DATE)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

###############################################################################
###                                  HTML                                   ###
###############################################################################

html: client/html
	@echo "Embedding HTML..."
	@statik -src=client/html -dest=client/ -f
	@echo "Embedded successfully"
.PHONY: html

###############################################################################
###                                  Build                                  ###
###############################################################################

build: go.sum
	@echo "Building Node Management binary..."
	@echo "Flags $(BUILD_FLAGS)"
	@go build -mod=readonly $(BUILD_FLAGS) -o build/nmngd ./cmd/nmngd
	@echo "Builded Node Management successfully"
.PHONY: build

###############################################################################
###                                 Install                                 ###
###############################################################################

install: go.sum
	@echo "Installing Node Management binary..."
	@echo "Flags $(BUILD_FLAGS)"
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/nmngd
	@echo " [v] Installed in GOPATH/bin"
	@if [ "$(shell uname)" = "Linux" ] && [ "$(IS_SUDO_USER)" = "1" ]; then \
		sudo mv $(shell $(GO_BIN) env GOPATH)/bin/nmngd /usr/local/bin/nmngd; \
		echo " [v] Installed as global command"; \
	else \
		echo " [x] (Skipped) Install as global command"; \
	fi
	@echo "Installed Node Management successfully"
.PHONY: install