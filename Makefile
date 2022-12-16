.PHONY: gobuild
	go build ./...

.PHONY: gotest
gotest:
	go test -v ./...

.PHONY: docs-serve
docs-serve:
	mkdocs serve -f docs/mkdocs.yaml