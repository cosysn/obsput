.PHONY: test build clean release all e2e e2e-test

# ... existing content ...

# Version info
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build directory
BUILD_DIR := build/$(VERSION)

# Platforms
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

test:
	go test ./...

build: \
	$(BUILD_DIR)/linux/amd64/obsput \
	$(BUILD_DIR)/linux/arm64/obsput \
	$(BUILD_DIR)/darwin/amd64/obsput \
	$(BUILD_DIR)/darwin/arm64/obsput \
	$(BUILD_DIR)/windows/amd64/obsput.exe

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)/linux/amd64
	mkdir -p $(BUILD_DIR)/linux/arm64
	mkdir -p $(BUILD_DIR)/darwin/amd64
	mkdir -p $(BUILD_DIR)/darwin/arm64
	mkdir -p $(BUILD_DIR)/windows/amd64

$(BUILD_DIR)/linux/amd64/obsput: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $@ main.go

$(BUILD_DIR)/linux/arm64/obsput: $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $@ main.go

$(BUILD_DIR)/darwin/amd64/obsput: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $@ main.go

$(BUILD_DIR)/darwin/arm64/obsput: $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $@ main.go

$(BUILD_DIR)/windows/amd64/obsput.exe: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $@ main.go

release: build $(BUILD_DIR)/obsput-$(VERSION)-linux-amd64.zip
release: $(BUILD_DIR)/obsput-$(VERSION)-linux-arm64.zip
release: $(BUILD_DIR)/obsput-$(VERSION)-darwin-amd64.zip
release: $(BUILD_DIR)/obsput-$(VERSION)-darwin-arm64.zip
release: $(BUILD_DIR)/obsput-$(VERSION)-windows-amd64.zip
	cd $(BUILD_DIR) && zip -q obsput-$(VERSION)-all.zip \
		linux/amd64/obsput \
		linux/arm64/obsput \
		darwin/amd64/obsput \
		darwin/arm64/obsput \
		windows/amd64/obsput.exe

$(BUILD_DIR)/obsput-$(VERSION)-linux-amd64.zip: $(BUILD_DIR)/linux/amd64/obsput
	cd $(BUILD_DIR)/linux/amd64 && zip -q ../../obsput-$(VERSION)-linux-amd64.zip obsput

$(BUILD_DIR)/obsput-$(VERSION)-linux-arm64.zip: $(BUILD_DIR)/linux/arm64/obsput
	cd $(BUILD_DIR)/linux/arm64 && zip -q ../../obsput-$(VERSION)-linux-arm64.zip obsput

$(BUILD_DIR)/obsput-$(VERSION)-darwin-amd64.zip: $(BUILD_DIR)/darwin/amd64/obsput
	cd $(BUILD_DIR)/darwin/amd64 && zip -q ../../obsput-$(VERSION)-darwin-amd64.zip obsput

$(BUILD_DIR)/obsput-$(VERSION)-darwin-arm64.zip: $(BUILD_DIR)/darwin/arm64/obsput
	cd $(BUILD_DIR)/darwin/arm64 && zip -q ../../obsput-$(VERSION)-darwin-arm64.zip obsput

$(BUILD_DIR)/obsput-$(VERSION)-windows-amd64.zip: $(BUILD_DIR)/windows/amd64/obsput.exe
	cd $(BUILD_DIR)/windows/amd64 && zip -q ../../obsput-$(VERSION)-windows-amd64.zip obsput.exe

clean:
	rm -rf build/

all: clean release

e2e:
	MINIO_ENDPOINT=localhost:9000 \
	MINIO_AK=admin \
	MINIO_SK=password \
	go test ./cmd/... -run TestE2E -v

e2e-setup: docker-compose.yaml scripts/start-minio.sh
	./scripts/start-minio.sh
	mc alias set myminio http://localhost:9000 admin password || true
	mc mb myminio/test-bucket --ignore-existing || true

e2e-clean:
	docker-compose down -v 2>/dev/null || docker compose down -v 2>/dev/null || true
