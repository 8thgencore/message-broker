# Переменные
CURRENT_DIR=$(shell pwd)
BIN_DIR=$(CURRENT_DIR)/bin
PROTO_DIR=$(CURRENT_DIR)/api/proto
VENDOR_PROTO_DIR=$(CURRENT_DIR)/vendor.protogen

# Go параметры
GOOS?=linux
GOARCH?=amd64

.PHONY: all
all: deps generate build

# ==============================================================================
# Установка зависимостей

.PHONY: deps
deps: deps-global deps-local

.PHONY: deps-global
deps-global:
	go install github.com/air-verse/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/yoheimuta/protolint/cmd/protolint@latest

.PHONY: deps-local
deps-local:
	GOBIN=$(BIN_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(BIN_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	GOBIN=$(BIN_DIR) go install github.com/envoyproxy/protoc-gen-validate@latest
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# ==============================================================================
# Линтинг и форматирование

.PHONY: lint
lint:
	$(BIN_DIR)/golangci-lint run ./internal/... ./cmd/... ./pkg/... -c .golangci.yaml --fix

.PHONY: format
format:
	$(BIN_DIR)/gofumpt -l -w .

.PHONY: protolint
protolint:
	$(BIN_DIR)/protolint lint $(PROTO_DIR)/*

# ==============================================================================
# Генерация кода

.PHONY: generate
generate: gen-proto gen-swagger

.PHONY: gen-proto
gen-proto:
	mkdir -p pkg/pb/broker/v1
	protoc -I $(PROTO_DIR) -I $(VENDOR_PROTO_DIR) \
		--go_out=pkg/pb --go_opt=paths=source_relative \
		--plugin=protoc-gen-go=$(BIN_DIR)/protoc-gen-go \
		--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-go-grpc=$(BIN_DIR)/protoc-gen-go-grpc \
		--grpc-gateway_out=pkg/pb --grpc-gateway_opt=paths=source_relative \
		--plugin=protoc-gen-grpc-gateway=$(BIN_DIR)/protoc-gen-grpc-gateway \
		$(PROTO_DIR)/broker/v1/*.proto

.PHONY: gen-swagger
gen-swagger:
	mkdir -p pkg/swagger
	protoc -I $(PROTO_DIR) -I $(VENDOR_PROTO_DIR) \
		--openapiv2_out=allow_merge=true,merge_file_name=api:pkg/swagger \
		--openapiv2_opt=logtostderr=true \
		--plugin=protoc-gen-openapiv2=$(BIN_DIR)/protoc-gen-openapiv2 \
		$(PROTO_DIR)/**/**/*.proto

# ==============================================================================
# Скачивание вендорного кода

.PHONY: vendor-proto
vendor-proto: vendor-proto-validate vendor-proto-google vendor-proto-openapiv2

.PHONY: vendor-proto-validate
vendor-proto-validate:
	@if [ ! -d $(VENDOR_PROTO_DIR)/validate ]; then \
		mkdir -p $(VENDOR_PROTO_DIR)/validate && \
		git clone --depth=1 https://github.com/envoyproxy/protoc-gen-validate $(VENDOR_PROTO_DIR)/protoc-gen-validate && \
		mv $(VENDOR_PROTO_DIR)/protoc-gen-validate/validate/*.proto $(VENDOR_PROTO_DIR)/validate && \
		rm -rf $(VENDOR_PROTO_DIR)/protoc-gen-validate; \
	fi

.PHONY: vendor-proto-google
vendor-proto-google:
	@if [ ! -d $(VENDOR_PROTO_DIR)/google ]; then \
		git clone --depth=1 https://github.com/googleapis/googleapis $(VENDOR_PROTO_DIR)/googleapis && \
		mkdir -p $(VENDOR_PROTO_DIR)/google/ && \
		mv $(VENDOR_PROTO_DIR)/googleapis/google/api $(VENDOR_PROTO_DIR)/google && \
		rm -rf $(VENDOR_PROTO_DIR)/googleapis; \
	fi

.PHONY: vendor-proto-openapiv2
vendor-proto-openapiv2:
	@if [ ! -d $(VENDOR_PROTO_DIR)/protoc-gen-openapiv2 ]; then \
		mkdir -p $(VENDOR_PROTO_DIR)/protoc-gen-openapiv2/options && \
		git clone --depth=1 https://github.com/grpc-ecosystem/grpc-gateway $(VENDOR_PROTO_DIR)/openapiv2 && \
		mv $(VENDOR_PROTO_DIR)/openapiv2/protoc-gen-openapiv2/options/*.proto $(VENDOR_PROTO_DIR)/protoc-gen-openapiv2/options && \
		rm -rf $(VENDOR_PROTO_DIR)/openapiv2; \
	fi

# ==============================================================================
# Сборка

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/broker cmd/broker/main.go

# ==============================================================================
# Разработка

.PHONY: run
run:
	go run cmd/broker/main.go

.PHONY: dev
dev:
	air

# ==============================================================================
# Очистка

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(VENDOR_PROTO_DIR)
