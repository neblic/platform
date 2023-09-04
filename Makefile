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

.PHONY: gen-proto
gen-proto:
	docker run --rm -v${PWD}:${PWD} -w${PWD} otel/build-protobuf \
	-I/usr/include/google/protobuf --proto_path=${PWD} \
	--go_opt=module=github.com/neblic/platform/controlplane --go_out=${PWD}/controlplane \
	--go-grpc_opt=module=github.com/neblic/platform/controlplane --go-grpc_out=${PWD}/controlplane \
	${PWD}/protos/controlplane.proto

	docker run --rm -v${PWD}:${PWD} -w${PWD} otel/build-protobuf \
	-I/usr/include/google/protobuf --proto_path=${PWD} \
	--go_opt=module=github.com/neblic/platform/dataplane --go_out=${PWD}/dataplane \
	--go-grpc_opt=module=github.com/neblic/platform/dataplane --go-grpc_out=${PWD}/dataplane \
	${PWD}/protos/dataplane.proto
