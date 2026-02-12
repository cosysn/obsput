# obsput 设计文档

## 概述

将构建产物上传到华为云OBS，生成下载链接，支持CLI和CI/CD集成。

## 功能特性

| 命令 | 功能 |
|------|------|
| `upload <file>` | 并发上传到多个OBS，支持进度条、MD5校验 |
| `list` | 查询历史文件，表格/JSON输出 |
| `delete <version>` | 删除指定版本，支持 `--name` 指定OBS |
| `download <version>` | 显示下载链接和curl/wget命令 |
| `obs add/set/get/list/rm` | OBS配置管理 |

## 使用示例

```bash
# 上传
obsput upload ./bin/myapp --prefix releases

# 查询历史
obsput list
obsput list -o json

# 删除
obsput delete v1.0.0-abc123-20260212-143000
obsput delete v1.0.0-abc123-20260212-143000 --name prod

# 下载
obsput download v1.0.0-abc123-20260212-143000

# 配置管理
obsput obs add --name prod --endpoint "obs.xxx.com" --bucket "bucket" --ak "ak" --sk "sk"
obsput obs list
```

## 目录结构

```
obsput/
├── cmd/
│   ├── root.go              # 根命令入口
│   ├── upload.go            # upload 子命令
│   ├── list.go              # list 子命令
│   ├── delete.go            # delete 子命令
│   ├── download.go          # download 子命令
│   └── obs.go               # obs 命令组 (add/set/get/list/rm)
├── pkg/
│   ├── config/
│   │   └── config.go        # 配置管理
│   ├── obs/
│   │   └── client.go        # OBS客户端封装
│   ├── version/
│   │   └── generator.go     # 版本号生成
│   ├── output/
│   │   └── formatter.go     # 输出格式化
│   └── progress/
│       └── progress.go      # 进度条
├── configs/
│   └── obsput.yaml          # 配置文件示例
├── main.go
├── go.mod
└── README.md
```

## 技术栈

| 依赖 | 用途 |
|------|------|
| github.com/huaweicloud/huaweicloud-sdk-go-obs | OBS SDK |
| github.com/spf13/cobra | CLI框架 |
| github.com/spf13/viper | 配置管理 |
| github.com/jedib0t/go-pretty/v6 | 表格输出 |

## 配置

### 环境变量

```bash
OBSPUT_ENDPOINT      # OBS域名
OBSPUT_BUCKET        # 桶名
OBSPUT_AK            # Access Key
OBSPUT_SK            # Secret Key
```

### 配置文件 (~/.obsput.yaml)

```yaml
configs:
  - name: obs-cn-east-1
    endpoint: "obs.cn-east-1.myhuaweicloud.com"
    bucket: "bucket-cn"
    ak: "ak1"
    sk: "sk1"
  - name: obs-cn-south-1
    endpoint: "obs.cn-south-1.myhuaweicloud.com"
    bucket: "bucket-south"
    ak: "ak2"
    sk: "sk2"
```

### 命令行参数

```bash
--name value      # 指定OBS名称（用于操作单个OBS）
--prefix value    # 上传路径前缀
-o, --output json # 输出格式 (table/json)
```

## 版本号规则

目录命名：`v<version>-<commit>-<date>-<time>/`

```
v1.0.0-abc123-20260212-143000/
└── myapp
```

## OBS存储结构

```
bucket/
├── releases/
│   └── v1.0.0-abc123-20260212-143000/
│       └── app-linux-amd64
├── v1.0.0-def456-20260212-150000/
│   └── app-linux-amd64
└── ...
```

## 输出示例

### 上传

```bash
$ obsput upload ./bin/myapp --prefix releases

[1/2] prod
███████████████████████████████  100%  12.5MB / 12.5MB  2.3MB/s
✓ Uploaded: https://obs.cn-east-1.myhuaweicloud.com/bucket-cn/releases/v1.0.0-abc123-20260212-143000/myapp
MD5:    abc123def456...

[2/2] staging
✗ Failed: connection timeout

2 completed, 1 failed
```

### 下载链接

```bash
✓ Uploaded: https://obs.xxx.com/bucket/v1.0.0-abc123-20260212-143000/myapp
MD5:    abc123def456...

Download:
  URL: https://obs.xxx.com/bucket/v1.0.0-abc123-20260212-143000/myapp

Commands:
  curl -O https://obs.xxx.com/bucket/v1.0.0-abc123-20260212-143000/myapp
  wget https://obs.xxx.com/bucket/v1.0.0-abc123-20260212-143000/myapp
```

### 列表

```bash
$ obsput list

[obs-cn-east-1]
VERSION                          SIZE    DATE            COMMIT    DOWNLOAD_URL
v1.0.0-abc123-20260212-143000    12.5MB  2026-02-12      abc123    https://...

[obs-cn-south-1]
VERSION                          SIZE    DATE            COMMIT    DOWNLOAD_URL
v1.0.0-abc123-20260212-143000    12.5MB  2026-02-12      abc123    https://...
```

### 配置管理

```bash
$ obsput obs add --name prod --endpoint "obs.xxx.com" --bucket "bucket" --ak "ak" --sk "sk"

$ obsput obs list
NAME            ENDPOINT                         BUCKET       STATUS
prod            obs.cn-east-1.myhuaweicloud.com bucket-prod  active

$ obsput obs get prod
Name:    prod
Endpoint: obs.cn-east-1.myhuaweicloud.com
Bucket:  bucket-prod
AK:      ********1234
SK:      ************5678
```

## 构建配置

### 目录结构

```
build/
└── v1.0.0/
    ├── linux/
    │   ├── amd64/
    │   │   └── obsput
    │   └── arm64/
    │       └── obsput
    ├── darwin/
    │   ├── amd64/
    │   │   └── obsput
    │   └── arm64/
    │       └── obsput
    ├── windows/
    │   └── amd64/
    │       └── obsput.exe
    ├── obsput-v1.0.0-linux-amd64.zip
    ├── obsput-v1.0.0-linux-arm64.zip
    ├── obsput-v1.0.0-darwin-amd64.zip
    ├── obsput-v1.0.0-darwin-arm64.zip
    ├── obsput-v1.0.0-windows-amd64.zip
    └── obsput-v1.0.0-all.zip
```

### Makefile

见 [build.md](./build.md)

## 测试

### 测试框架

- `testing` - Go标准测试框架

### 测试覆盖

| 包 | 测试文件 | 测试内容 |
|---|---------|---------|
| pkg/config | config_test.go | 配置文件读写、增删改查 |
| pkg/obs | client_test.go | OBS客户端连接、上传/删除 |
| pkg/version | generator_test.go | 版本号生成 |
| pkg/output | formatter_test.go | 输出格式化 |
| pkg/progress | progress_test.go | 进度条显示 |

### Git Hook

```bash
#!/bin/bash
# .githooks/pre-commit
go test ./... || exit 1
```

## CI/CD集成

```yaml
# GitHub Actions
- name: Upload to OBS
  run: |
    go run cmd/obsput/main.go upload ./bin/app \
      --endpoint=${{ secrets.OBS_ENDPOINT }} \
      --bucket=${{ secrets.OBS_BUCKET }} \
      --ak=${{ secrets.OBS_AK }} \
      --sk=${{ secrets.OBS_SK }}
```

## .gitignore

```
# Binaries
obsput
obsput.exe

# Build outputs
build/
dist/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Go
*.exe
*.exe~
*.test
*.out
*.prof
*.coverprofile

# Environment
.env
.env.local

# Logs
*.log

# Temporary
tmp/
temp/
```
