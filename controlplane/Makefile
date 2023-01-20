.PHONY: gen-proto
gen-proto:
	docker run --rm -v${PWD}:${PWD} -w${PWD} otel/build-protobuf \
	-I/usr/include/google/protobuf --proto_path=${PWD} \
	--go_opt=module=github.com/neblic/platform/controlplane --go_out=${PWD} \
	--go-grpc_opt=module=github.com/neblic/platform/controlplane --go-grpc_out=${PWD} \
	${PWD}/internal/protos/controlplane.proto
