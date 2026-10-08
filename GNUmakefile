default: testacc

# Test configuration
TEST ?= ./...
RUN ?= TestAcc
TESTARGS ?=

# Function to run tests with optional .env.test loading
define run_test
	@if [ -z "$$HOOKDECK_API_KEY" ] || [ -z "$$TF_ACC" ]; then \
		if [ -f .env.test ]; then \
			echo "📋 Loading from .env.test..."; \
			set -a; . ./.env.test; set +a; \
		else \
			echo "❌ Missing configuration:"; \
			[ -z "$$HOOKDECK_API_KEY" ] && echo "   - HOOKDECK_API_KEY not set"; \
			[ -z "$$TF_ACC" ] && echo "   - TF_ACC not set"; \
			echo ""; \
			echo "💡 To fix:"; \
			echo "   1. Copy .env.test.example to .env.test"; \
			echo "   2. Add your HOOKDECK_API_KEY to .env.test"; \
			echo "   OR"; \
			echo "   3. Export variables in your shell"; \
			exit 1; \
		fi; \
	fi; \
	echo "🧪 Running tests on $(1)..."; \
	go test $(1) -v -run $(2) $(TESTARGS) -timeout 120m
endef

# Run acceptance tests
.PHONY: testacc
testacc:
	$(call run_test,$(TEST),$(RUN))

generate:
	go generate ./...

generate-tfdocs:
	go generate ./cmd/tfdocs/...

enable-git-hooks:
	git config --local include.path ../.gitconfig
	$(warning REMEMBER, YOU MUST HAVE REVIEWED THE CUSTOM HOOKS!)
