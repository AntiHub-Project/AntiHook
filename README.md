# AntiHook

## 构建

### 构建所有平台

```bash
./build.sh all
```

### 构建特定平台

```bash
# macOS
./build.sh darwin

# Windows
./build.sh windows

# Linux
./build.sh linux
```

构建产物位于 `build/` 目录：
- `antihook-darwin-amd64` - macOS Intel
- `antihook-darwin-arm64` - macOS Apple Silicon
- `antihook-windows-amd64.exe` - Windows 64位
- `antihook-linux-amd64` - Linux 64位


## 开发

### 依赖安装

```bash
go mod download
```

### 本地测试

```bash
# 编译
go build -o antihook .

# 测试
./antihook "kiro://test"
./antihook "anti://?identity=test-token&is_shared=0"
```
