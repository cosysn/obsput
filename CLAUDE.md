# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go test ./...            # Run all unit tests
go test -run TestName ./...  # Run specific test
make build               # Build for current platform
make release             # Build all 5 platforms with ZIP archives
make clean               # Remove build artifacts
make e2e                 # Run E2E tests against MinIO
make e2e-setup           # Start MinIO via docker-compose for testing
make e2e-clean           # Stop MinIO container
```

## Project Overview

**obsput** is a CLI tool written in Go that uploads binaries to Huawei Cloud OBS (Object Storage Service). It uses the Huawei Cloud SDK which is S3-compatible.

## Architecture

### CLI Layer (`cmd/`)
Cobra-based commands with shared flags (`--profile/-p` for targeting specific OBS configs):
- `obsput put <file>` - Upload with automatic version tracking
- `obsput list` - List uploaded versions (supports `--json` output)
- `obsput delete <version>` - Delete specific version
- `obsput download <version>` - Show download commands
- `obsput obs [add/get/list/remove/init]` - Manage OBS configurations
- `obsput version` - Show build version info

### Configuration
- Stored at `{binary_dir}/.obsput/obsput.yaml`
- Supports multiple named OBS configurations
- Environment variable fallback (OBS_ENDPOINT, OBS_BUCKET, OBS_AK, OBS_SK)

### OBS Client (`pkg/obs/`)
- Uses `huaweicloud-sdk-go-obs` (S3-compatible SDK)
- Path-style URL for IP addresses/localhost, virtual-hosted-style for domain names
- Signed URL generation with 24h expiry
- TCP connection testing before operations

### Version Format
`v{version}-{commit}-{date}-{timestamp}-{counter}` (e.g., `v1.0.0-abc123-20260214-153045-1`)

## Key Implementation Details

- **Parallel uploads**: The `put` command uploads to multiple OBS endpoints concurrently
- **Config location**: Config is always relative to the binary location, not current working directory
- **MinIO for testing**: E2E tests run against MinIO via docker-compose (set `MINIO_ENDPOINT`, `MINIO_AK`, `MINIO_SK`)
- **Git-based versioning**: Version info is generated at build time using git commit and timestamp
