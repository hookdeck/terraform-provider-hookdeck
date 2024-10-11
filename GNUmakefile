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
