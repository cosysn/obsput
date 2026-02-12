# 配置文件设计

## 目录结构

```
/path/to/
├── obsput           # 二进制文件
└── .obsput/
    └── obsput.yaml  # 配置文件
```

## 配置路径

**优先级：**
1. 二进制文件所在目录：`{executable_dir}/.obsput/obsput.yaml`

**不再使用：**
- `~/.obsput/obsput.yaml`

## 配置模板

```yaml
configs: []
```

## 使用流程

### 首次运行

```bash
$ ./obsput upload ./app
✗ No OBS configurations configured

Please add OBS configuration:
  ./obsput obs add --name prod --endpoint "obs.xxx.com" --bucket "bucket" --ak "xxx" --sk "xxx"
```

### 添加配置

```bash
$ ./obsput obs add --name prod \
  --endpoint "obs.cn-east-1.myhuaweicloud.com" \
  --bucket "my-bucket" \
  --ak "your-access-key" \
  --sk "your-secret-key"

✓ Added OBS config: prod
```

### 查看配置

```bash
$ ./obsput obs list
NAME    ENDPOINT                             BUCKET     STATUS
prod    obs.cn-east-1.myhuaweicloud.com      my-bucket  active
```

### 删除配置

```bash
$ ./obsput obs remove prod
✓ Removed OBS config: prod
```

## 实现方式

### 1. 获取二进制目录

```go
import (
    "os"
    "path/filepath"
)

func GetConfigPath() (string, error) {
    execPath, err := os.Executable()
    if err != nil {
        return "", err
    }
    dir := filepath.Dir(execPath)
    return filepath.Join(dir, ".obsput", "obsput.yaml"), nil
}
```

### 2. 配置文件模板

```go
package config

const DefaultConfig = `configs: []
`

func (c *Config) Ensure(path string) error {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return c.Save(path)
    }
    return nil
}
```

### 3. 命令执行时检查配置

```go
func getConfigPath() string {
    // 实现获取配置路径
}

func getConfig() (*config.Config, error) {
    path, err := getConfigPath()
    if err != nil {
        return nil, err
    }

    cfg, err := config.Load(path)
    if err != nil {
        return nil, err
    }

    if len(cfg.Configs) == 0 {
        return nil, fmt.Errorf("No OBS configurations configured")
    }

    return cfg, nil
}
```

## 修改文件

| 文件 | 修改内容 |
|------|---------|
| `cmd/root.go` | 添加 `getConfigPath()` 函数 |
| `cmd/upload.go` | 使用新路径，添加配置检查 |
| `cmd/list.go` | 使用新路径，添加配置检查 |
| `cmd/delete.go` | 使用新路径，添加配置检查 |
| `cmd/download.go` | 使用新路径，添加配置检查 |
| `cmd/obs.go` | 添加 `obs init` 命令 |
| `pkg/config/config.go` | 添加 `Ensure()` 方法 |

## 命令变更

| 命令 | 说明 |
|------|------|
| `obs add` | 添加 OBS 配置 |
| `obs list` | 列出所有配置 |
| `obs get <name>` | 查看单个配置 |
| `obs remove <name>` | 删除配置 |
| `obs init` | 初始化配置文件 |
