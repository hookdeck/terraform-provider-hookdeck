default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

generate:
	go generate ./...

generate-codegen:
	go generate ./cmd/codegen/...

generate-tfdocs:
	go generate ./cmd/tfdocs/...

enable-git-hooks:
	git config --local include.path ../.gitconfig
	$(warning REMEMBER, YOU MUST HAVE REVIEWED THE CUSTOM HOOKS!)

download:
	echo Download go.mod dependencies
	go mod download

install-tools: download
	echo Installing tools from tools.go
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
