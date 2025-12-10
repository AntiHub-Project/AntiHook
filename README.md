# AntiHook

## 构建

```bash
# macOS
./build.sh darwin

# Windows
./build.sh windows

# Linux
./build.sh linux
```

## 开发

### 依赖安装

```bash
go mod download
```

### 编译

```bash
go build -o antihook .
```
### 使用

对于MacOS，使用前请先安装duti：

```bash
brew install duti
```

对于Windows，请运行AntiHook至少一次。

如果要移除Hook，请运行：
```bash
antihook --recover
```


