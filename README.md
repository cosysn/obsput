# obsput

Upload binaries to Huawei Cloud OBS with CLI tool. Supports multi-OBS configuration, CI/CD integration, and cross-platform builds.

## Features

- **Multi-OBS Support**: Upload to multiple OBS endpoints simultaneously
- **Progress Bar**: Real-time upload progress with speed display
- **Version Management**: Automatic version tracking with git commit ID
- **CI/CD Ready**: Works in GitHub Actions, GitLab CI, etc.
- **Cross-Platform**: Linux, macOS, Windows support

## Installation

### Download Pre-built Binary

Download from [Releases](https://github.com/cosysn/obsput/releases):

```bash
# Linux amd64
wget https://github.com/cosysn/obsput/releases/download/v0.3.0/obsput-v0.3.0-linux-amd64.zip
unzip obsput-v0.3.0-linux-amd64.zip

# macOS arm64
wget https://github.com/cosysn/obsput/releases/download/v0.3.0/obsput-v0.3.0-darwin-arm64.zip
unzip obsput-v0.3.0-darwin-arm64.zip

# Windows
wget https://github.com/cosysn/obsput/releases/download/v0.3.0/obsput-v0.3.0-windows-amd64.zip
unzip obsput-v0.3.0-windows-amd64.zip
```

### Build from Source

```bash
# Clone
git clone https://github.com/cosysn/obsput.git
cd obsput

# Build all platforms
make release

# Or build for current platform
go build -o obsput main.go
```

## Configuration

The configuration file is automatically created in the same directory as the binary:

```
/path/to/
├── obsput           # Binary
└── .obsput/
    └── obsput.yaml  # Config file
```

### Add OBS Configuration

```bash
./obsput obs add --name prod \
  --endpoint "obs.cn-east-1.myhuaweicloud.com" \
  --bucket "my-bucket" \
  --ak "your-access-key" \
  --sk "your-secret-key"
```

### Configure Multiple OBS

```bash
# Add first OBS
./obsput obs add --name cn-east \
  --endpoint "obs.cn-east-1.myhuaweicloud.com" \
  --bucket "bucket-cn" \
  --ak "ak1" \
  --sk "sk1"

# Add second OBS
./obsput obs add --name cn-south \
  --endpoint "obs.cn-south-1.myhuaweicloud.com" \
  --bucket "bucket-south" \
  --ak "ak2" \
  --sk "sk2"
```

### Manage Configurations

```bash
# List all configurations
./obsput obs list

# Get single configuration
./obsput obs get prod

# Remove configuration
./obsput obs remove prod

# Initialize config (if not exists)
./obsput obs init
```

## Usage

### Upload Binary

```bash
# Simple upload
./obsput upload ./bin/myapp

# Upload with prefix
./obsput upload ./bin/myapp --prefix releases

# Upload with specific OBS
./obsput upload ./bin/myapp --name prod
```

Output:
```
Uploading: ./bin/myapp
Version: v1.0.0-abc123-20260212-143000

[prod]
  Uploaded: https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp
  MD5: abc123def456...
  Size: 12.5MB
  Speed: 2.3MB/s

Download:
  URL: https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp

Commands:
  curl -O https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp
  wget https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp

1 completed, 0 failed
```

### List Versions

```bash
# Table format (default)
./obsput list

# JSON format
./obsput list -o json
```

Output:
```
[prod]
VERSION                          SIZE    DATE            COMMIT    DOWNLOAD_URL
v1.0.0-abc123-20260212-143000   12.5MB  2026-02-12      abc123    https://...
v1.0.1-def456-20260213-150000   13.2MB  2026-02-13      def456    https://...
```

### Delete Version

```bash
# Delete from all OBS
./obsput delete v1.0.0-abc123-20260212-143000

# Delete from specific OBS
./obsput delete v1.0.0-abc123-20260212-143000 --name prod
```

Output:
```
[prod] Deleting v1.0.0-abc123-20260212-143000...
[prod] Deleted: v1.0.0-abc123-20260212-143000
```

### Download Info

```bash
./obsput download v1.0.0-abc123-20260212-143000
```

Output:
```
[prod]
Version: v1.0.0-abc123-20260212-143000
URL: https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp

Commands:
  curl -O https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp
  wget https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp
  curl -o myapp https://bucket.obs.cn-east-1.myhuaweicloud.com/releases/v1.0.0-abc123-20260212-143000/myapp
```

### Version Info

```bash
./obsput version
```

Output:
```
Version: v0.3.0
Commit: abc1234
Date: 2026-02-12
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Build and Upload to OBS

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build
        run: |
          make build

      - name: Upload to OBS
        run: |
          cd build/v*/linux/amd64
          ./obsput upload obsput --prefix releases
        env:
          OBS_ENDPOINT: ${{ secrets.OBS_ENDPOINT }}
          OBS_BUCKET: ${{ secrets.OBS_BUCKET }}
          OBS_AK: ${{ secrets.OBS_AK }}
          OBS_SK: ${{ secrets.OBS_SK }}
```

### GitLab CI

```yaml
stages:
  - build
  - deploy

build:
  stage: build
  script:
    - make build
  artifacts:
    paths:
      - build/v*/linux/amd64/obsput

deploy:
  stage: deploy
  script:
    - cd build/v*/linux/amd64
    - ./obsput upload obsput --prefix releases
  environment:
    name: production
```

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /app/build/v*/linux/amd64/obsput /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/obsput"]
```

## Environment Variables (Alternative)

You can also use environment variables instead of config file:

```bash
export OBS_ENDPOINT="obs.cn-east-1.myhuaweicloud.com"
export OBS_BUCKET="my-bucket"
export OBS_AK="your-access-key"
export OBS_SK="your-secret-key"
```

## Build Commands

```bash
# Run tests
make test

# Build current platform
make build

# Build all platforms (Linux, macOS, Windows)
make release

# Clean build artifacts
make clean

# Full clean + release
make all
```

## Project Structure

```
obsput/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── upload.go          # Upload command
│   ├── list.go            # List command
│   ├── delete.go          # Delete command
│   ├── download.go        # Download command
│   └── obs.go             # Config management
├── pkg/                    # Packages
│   ├── config/            # Configuration
│   ├── obs/               # OBS client
│   ├── version/           # Version generator
│   ├── output/            # Formatter
│   └── progress/          # Progress bar
├── scripts/               # Utility scripts
├── docker-compose.yaml    # MinIO for testing
├── Makefile               # Build automation
└── README.md              # This file
```

## License

MIT
