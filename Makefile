
all: build

prepare:
#	go mod init grpc-gateway-example
	go get google.golang.org/grpc
	go get github.com/grpc-ecosystem/grpc-gateway/v2
	go get github.com/golang/protobuf/protoc-gen-go
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2


PROTO_DIR := proto
OUT_DIR := pkg/service
GOOGLEAPIS_DIR := $(PROTO_DIR)/googleapis

PROTO_FILES := $(PROTO_DIR)/service.proto

build:
	protoc -I $(PROTO_DIR) -I $(GOOGLEAPIS_DIR) $(PROTO_FILES) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(OUT_DIR) --grpc-gateway_opt=paths=source_relative

clean:
	rm -rf $(OUT_DIR)/*.pb.go
