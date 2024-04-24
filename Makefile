ANSI_BOLD := $(if $NO_COLOR,$(shell tput bold 2>/dev/null),)
ANSI_RESET := $(if $NO_COLOR,$(shell tput sgr0 2>/dev/null),)

# Switch this to podman if you are using that in place of docker
CONTAINERTOOL := docker

MODULE_NAME := $(lastword $(subst /, ,$(shell go list -m)))
VERSION := $(if $(shell git status --porcelain 2>/dev/null),latest,$(shell git rev-parse HEAD))

##@ Build

.PHONY: build
build: .cloudbees/testing/action.yml ## Build the container image
	@echo "$(ANSI_BOLD)⚡️ Building container image ...$(ANSI_RESET)"
	@$(CONTAINERTOOL) build --rm -t $(MODULE_NAME):$(VERSION) -t $(MODULE_NAME):latest -f Dockerfile .
	@echo "✅ Container image built$(ANSI_RESET)"

build-test: ## Build the test container image
	@echo "$(ANSI_BOLD)⚡️ Building test container image ...$(ANSI_RESET)"
	@$(CONTAINERTOOL) build --target testing -t $(MODULE_NAME):$(VERSION)-testing -t $(MODULE_NAME):latest-testing -f Dockerfile .
	@echo "$(ANSI_BOLD)✅ Container test image built and tagged as $(MODULE_NAME):latest-testing"

manual-test: build-test ## Interactively runs the test container to allow manual testing
	@echo '$(ANSI_BOLD)ℹ️  To set-up credentials, run something like:$(ANSI_RESET)'
	@echo '       configure-git-global-credentials configure \'
	@echo '            --provider github \'
	@echo '            --repositories "*/*" \'
	@echo '            --token ...your.token.here...'
	@echo '$(ANSI_BOLD)ℹ️  Critical test cases:$(ANSI_RESET)'
	@echo '      1. `git clone ...some.private.repo.url.here...`'
	@echo '      2. Using `go mod download` from source that has private module dependencies'
	@$(CONTAINERTOOL) run --rm -ti $(MODULE_NAME):$(VERSION)-testing

.PHONY: test
test: ## Runs unit tests
	@echo "$(ANSI_BOLD)⚡️ Running unit tests ...$(ANSI_RESET)"
	@go test ./...
	@echo "$(ANSI_BOLD)✅ Unit tests passed$(ANSI_RESET)"

.PHONY: verify
verify: format sync test ## Verifies that the committed code is formatted, all files are in sync and the unit tests pass
	@if [ "`git status --porcelain 2>/dev/null`x" = "x" ] ; then \
	  echo "$(ANSI_BOLD)✅ Git workspace is clean$(ANSI_RESET)" ; \
	else \
	  echo "$(ANSI_BOLD)❌ Git workspace is dirty$(ANSI_RESET)" ; \
	  exit 1 ; \
	fi

.cloudbees/testing/action.yml: action.yml Makefile ## Ensures that the test version of the action.yml is in sync with the production version
	@echo "$(ANSI_BOLD)⚡️ Updating $@ ...$(ANSI_RESET)"
	@sed -e 's|docker://public.ecr.aws/l7o7z1g8/actions/|docker://020229604682.dkr.ecr.us-east-1.amazonaws.com/actions/|g' < action.yml > .cloudbees/testing/action.yml

.cloudbees/workflows/workflow.yml: Dockerfile ## Ensures that the workflow uses the same version of go as the Dockerfile
	@echo "$(ANSI_BOLD)⚡️ Updating $@ ...$(ANSI_RESET)"
	@IMAGE=$$(sed -ne 's/FROM[ \t]*golang:\([^ \t]*\)-alpine[0-9.]*[ \t].*/\1/p' Dockerfile) ; \
	sed -e 's|\(uses:[ \t]*docker://golang:\)[^ \t]*|\1'"$$IMAGE"'|;' < $@ > $@.bak ; \
	mv -f $@.bak $@

.PHONY: sync
sync: .cloudbees/testing/action.yml .cloudbees/workflows/workflow.yml ## Updates action.yml so that the container tag matches the VERSION file
	@echo "$(ANSI_BOLD)✅ All files synchronized$(ANSI_RESET)"

.PHONY: check-git-status
check-git-status: ## Checks if there are any uncommitted changes in the repository
	@echo "$(ANSI_BOLD)⚡️ Checking for uncommitted changes...$(ANSI_RESET)"
	@if ! git diff-index --quiet HEAD --; then \
		echo "❌ There are uncommitted changes in the repository." ; \
		git status ; \
		exit 1; \
	fi
	@echo "$(ANSI_BOLD)✅ No uncommitted changes in the repository$(ANSI_RESET)"

.PHONY: format
format: ## Applies the project code style
	@echo "$(ANSI_BOLD)⚡️ Applying project code style ...$(ANSI_RESET)"
	@gofmt -w .
	@echo "$(ANSI_BOLD)✅ Project code style applied$(ANSI_RESET)"

##@ Miscellaneous

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

