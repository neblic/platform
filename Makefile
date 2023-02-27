GOMODULES = . cmd/neblictl cmd/kafka-sampler 

.PHONY: gobuild
gobuild:
	for gomod in $(GOMODULES); do \
		echo "Building $$gomod" && \
			cd $$gomod && \
			go build ./... && \
			cd -; \
	done

.PHONY: gotest
gotest:
	for gomod in $(GOMODULES); do \
		echo "Testing $$gomod" && \
			cd $$gomod && \
			go test ./... && \
			cd -; \
	done

.PHONY: gomod-update-all
gomod-update-all:
	for gomod in $(GOMODULES); do \
		echo "Updating $$gomod" && \
			cd $$gomod && \
			go get -u ./... && \
			go mod tidy && \
			cd -; \
	done

.PHONY: docs-serve
docs-serve:
	mkdocs serve -f docs/mkdocs.yaml