.PHONY: generate
generate:
	protoc -I api/proto \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=. \
		--openapiv2_opt=allow_merge=true,merge_file_name=api \
		api/proto/broker/v1/*.proto

.PHONY: run
run:
	go run cmd/broker/main.go

.PHONY: build
build:
	go build -o bin/broker cmd/broker/main.go 