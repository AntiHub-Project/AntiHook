# AntiHook - 自动登录工具

AntiHook 是一个跨平台的协议处理器工具，支持 `kiro://` 和 `anti://` 自定义 URL 协议，实现 OAuth 自动登录流程。

## 支持平台

- ✅ macOS (Intel & Apple Silicon)
- ✅ Windows
- ✅ Linux

## 功能特性

- 🔐 自动处理 OAuth 登录流程
- 🌐 支持自定义协议 (`kiro://` 和 `anti://`)
- 🔄 自动注册协议处理器
- 💻 跨平台支持
- ⚡ 快速响应
- 📦 **构建时自动配置后端地址**
- 📝 **详细的日志记录系统**
- 🎨 **静默运行，无中间弹框干扰**

## 快速开始

### macOS 安装

#### 方法一：使用构建脚本

```bash
# 克隆仓库
git clone <repository-url>
cd AntiHook

# 构建 macOS 版本
./build.sh darwin

# 安装（构建后的二进制文件会在 build 目录）
# Intel Mac
sudo cp build/antihook-darwin-amd64 /usr/local/bin/antihook

# Apple Silicon Mac
sudo cp build/antihook-darwin-arm64 /usr/local/bin/antihook

# 运行安装程序（注册协议处理器）
antihook
```

#### 方法二：直接编译

```bash
# 克隆仓库
git clone <repository-url>
cd AntiHook

# 编译
go build -o antihook .

# 运行安装
./antihook
```

#### 安装 duti（可选，用于更好的协议处理）

```bash
brew install duti
```

### Windows 安装

```bash
# 克隆仓库
git clone <repository-url>
cd AntiHook

# 构建 Windows 版本
./build.sh windows

# 或者在 Windows 上直接编译
go build -o antihook.exe .

# 运行安装程序（将自动注册协议处理器）
./antihook.exe
```

### Linux 安装

```bash
# 克隆仓库
git clone <repository-url>
cd AntiHook

# 构建
./build.sh linux

# 安装
sudo cp build/antihook-linux-amd64 /usr/local/bin/antihook

# 运行安装
antihook
```

## 使用说明

### 安装协议处理器

直接运行程序即可自动安装：

```bash
# macOS/Linux
./antihook

# Windows
antihook.exe
```

安装后会：
1. 将程序复制到系统目录
2. 注册 `kiro://` 和 `anti://` 协议处理器
3. 添加到系统 PATH（可选）

### 协议使用

#### Kiro 协议

```
kiro://your-callback-url
```

程序会自动将回调 URL 转发到服务器完成登录。

#### Anti 协议

```
anti://?identity=your-bearer-token&is_shared=0
```

参数说明：
- `identity`: Bearer token（可以不带 "Bearer " 前缀）
- `is_shared`: 是否共享（0 或 1，默认 0）

### 恢复原始处理器（仅 Windows）

```bash
antihook.exe --recover
```

### 🆕 构建时配置后端地址（新功能）

AntiHook 支持在构建时自动注入后端地址，打包后的程序会自动使用您配置的生产环境地址，**无需用户手动配置**。

#### 什么是"打包/构建"？

**打包（Build）** 是将源代码编译成可执行文件的过程。在这个过程中，您配置的后端地址会被自动注入到程序中。

```
源代码 + 配置文件 → 构建脚本 → 可执行文件（已包含配置）
```

#### 方法一：使用配置文件（推荐）

**适用场景**：发布生产版本，配置会永久保存

1. 复制配置文件模板：
```bash
cp .build.config.example .build.config
```

2. 编辑 `.build.config` 文件，设置您的后端地址：
```bash
# Kiro 服务器地址（用于 kiro:// 协议回调）
SERVER_URL="https://tunnel.mortis.edu.kg"

# 后端服务器地址（用于 anti:// 协议 OAuth）
BACKEND_URL="https://tunnel.mortis.edu.kg"
```

3. 运行构建，配置会自动注入：
```bash
./build.sh all
```

**完成后**：
- ✅ 生成的可执行文件已包含您的后端地址
- ✅ 用户运行时无需任何配置
- ✅ 自动连接到您指定的服务器

#### 方法二：使用临时环境变量

**适用场景**：测试不同环境，不影响配置文件

```bash
# 临时设置环境变量并构建
SERVER_URL="https://test-server.com" \
BACKEND_URL="https://test-backend.com" \
./build.sh all
```

#### 方法三：运行时环境变量（兼容旧版）

**适用场景**：调试或临时切换服务器

```bash
# Kiro 服务器地址（默认: http://localhost:8045）
export KIRO_SERVER_URL="https://your-kiro-server.com"

# 后端服务器地址（默认: http://localhost:8008）
export BACKEND_URL="https://your-backend-server.com"
```

**注意**：运行时环境变量优先级最高，会覆盖构建时的配置。

#### 验证配置是否生效

构建完成后，验证配置是否正确注入：

```bash
# macOS/Linux - 检查配置是否注入
strings build/antihook-darwin-amd64 | grep "https"

# 应该能看到您配置的地址
# https://tunnel.mortis.edu.kg

# Windows - 使用 PowerShell
Select-String -Path build/antihook-windows-amd64.exe -Pattern "https://" -AllMatches
```

## 构建说明

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

## 技术架构

### 目录结构

```
AntiHook/
├── main.go                    # 主程序（跨平台核心逻辑）
├── main_darwin.go             # macOS 特定实现
├── main_windows.go            # Windows 特定实现
├── registry/
│   ├── registry_darwin.go     # macOS 协议注册
│   └── registry_windows.go    # Windows 协议注册
├── build.sh                   # 构建脚本
├── go.mod                     # Go 模块定义
└── README.md                  # 说明文档
```

### 工作流程

1. **协议注册**
   - macOS: 使用 AppleScript 创建协议处理器
   - Windows: 使用注册表注册协议

2. **OAuth 流程**
   - 接收协议 URL
   - 解析参数
   - 启动本地 HTTP 服务器（端口 42532）
   - 打开浏览器进行授权
   - 接收回调完成登录

3. **API 接口**
   - `POST /api/kiro/oauth/callback` - Kiro 回调
   - `POST /api/plugin-api/oauth/authorize` - Anti 授权
   - `POST /api/plugin-api/oauth/callback` - Anti 回调

## macOS 特别说明

### 协议处理器位置

macOS 版本会在以下位置创建协议处理器：

```
~/.config/antihook/kiro_handler.app
~/.config/antihook/anti_handler.app
```

### 首次使用提示

在 macOS 上首次点击 `kiro://` 或 `anti://` 链接时，系统会询问是否允许打开该应用。请选择"始终允许"以获得最佳体验。

### 浏览器刷新

注册协议后可能需要重启浏览器才能使新协议生效。

## 故障排除

### macOS 协议未生效

```bash
# 重新安装
./antihook

# 检查处理器是否存在
ls -la ~/.config/antihook/

# 重启浏览器
```

### Windows 协议未生效

```bash
# 以管理员权限运行
# 右键 -> 以管理员身份运行
antihook.exe

# 检查注册表
# 运行 regedit，查看 HKEY_CURRENT_USER\Software\Classes\kiro
```

## 🆕 日志系统

### 自动日志记录

所有操作都会自动记录到日志文件中，方便调试和排查问题：

**日志文件位置**：`~/.config/antihook/kiro.log`

**日志内容示例**：
```
=== 2025-12-08 15:02:22 ===
Received kiro:// callback: kiro://kiro.kiroAgent/authenticate-success?code=xxx&state=yyy
Posting to: https://tunnel.mortis.edu.kg/api/kiro/oauth/callback
Request body: {"callback_url":"kiro://..."}
Response status: 200
Response body: {"success":true}
Login successful!
```

### 查看日志

```bash
# 查看完整日志
cat ~/.config/antihook/kiro.log

# 实时监控日志
tail -f ~/.config/antihook/kiro.log

# 查看最后 20 行
tail -20 ~/.config/antihook/kiro.log
```

### 清理日志

```bash
# 清空日志文件
> ~/.config/antihook/kiro.log

# 删除日志文件
rm ~/.config/antihook/kiro.log
```

## 🎨 用户体验优化

### 静默运行模式

程序在处理登录时不会显示中间过程弹框，只在以下情况显示提示：

- ✅ **登录成功**：显示"Login successful!"
- ❌ **登录失败**：显示具体错误信息
- 🔄 **处理中**：静默处理，无弹框干扰

这样可以提供更流畅的用户体验，避免不必要的打扰。

### 查看构建信息

构建后的程序包含了构建时的配置信息：

```bash
# macOS/Linux - 查看所有 HTTP(S) 地址
strings build/antihook-darwin-amd64 | grep "https"

# 查看构建版本和时间
strings build/antihook-darwin-amd64 | grep -E "BuildVersion|BuildTime"
```

## 🆕 完整的文档体系

- **[BUILD_GUIDE.md](BUILD_GUIDE.md)** - 详细的构建配置指南和多环境部署
- **[MAC_INSTALL_GUIDE.md](MAC_INSTALL_GUIDE.md)** - macOS 完整安装和使用教程
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - 常见问题排查和解决方案
- **[API_ISSUE_REPORT.md](API_ISSUE_REPORT.md)** - API 问题诊断报告

## 故障排查

### 端口冲突

如果端口 42532 被占用，请关闭占用该端口的程序：

```bash
# macOS/Linux
lsof -i :42532
kill <PID>

# Windows
netstat -ano | findstr :42532
taskkill /PID <PID> /F
```

## 开发

### 依赖安装

```bash
go mod download
```

### 本地测试

```bash
# 编译
go build -o antihook .

# 测试协议
./antihook "kiro://test"
./antihook "anti://?identity=test-token&is_shared=0"
```

### 添加新平台支持

1. 创建 `main_<platform>.go` 文件
2. 实现平台特定函数：
   - `showMessageBox()`
   - `addToPath()`
   - `recoverOriginal()`
   - `openBrowser()`
3. 创建 `registry/registry_<platform>.go`
4. 实现 `ProtocolHandler` 接口
5. 更新构建脚本

## 许可证

[添加许可证信息]

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.0.0
- ✨ 初始版本
- ✅ 支持 macOS、Windows、Linux
- 🔐 实现 OAuth 自动登录
- 📦 提供跨平台构建脚本