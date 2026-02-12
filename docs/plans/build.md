# 构建配置

## Makefile

```makefile
.PHONY: build clean release

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

BUILD_DIR := build/$(VERSION)

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

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

build: $(patsubst %,$(BUILD_DIR)/%/obsput,linux/amd64 linux/arm64 darwin/amd64 darwin/arm64)
build: $(BUILD_DIR)/windows/amd64/obsput.exe

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

release: build
release: $(BUILD_DIR)/obsput-$(VERSION)-linux-amd64.zip
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

clean:
	rm -rf build/
```

## 使用方法

```bash
# 构建所有平台
make build

# 构建并打包
make release

# 清理构建产物
make clean

# 查看版本
./obsput version
```

## 构建产物

```
build/v1.0.0/
├── linux/amd64/obsput
├── linux/arm64/obsput
├── darwin/amd64/obsput
├── darwin/arm64/obsput
├── windows/amd64/obsput.exe
├── obsput-v1.0.0-linux-amd64.zip
├── obsput-v1.0.0-linux-arm64.zip
├── obsput-v1.0.0-darwin-amd64.zip
├── obsput-v1.0.0-darwin-arm64.zip
├── obsput-v1.0.0-windows-amd64.zip
└── obsput-v1.0.0-all.zip
```
