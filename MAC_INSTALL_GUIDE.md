# AntiHook macOS 安装和使用指南

## 📦 快速安装

### 步骤 1: 运行构建的程序

```bash
# 进入项目目录
cd /Users/xswu/work/project/code/AntiHook

# 根据您的 Mac 芯片选择对应版本：

# Intel Mac (x86_64)
./build/antihook-darwin-amd64

# Apple Silicon Mac (M1/M2/M3)
./build/antihook-darwin-arm64
```

**首次运行时**，macOS 可能会提示"无法验证开发者"，请按以下步骤操作：

1. 打开 **系统设置** → **隐私与安全性**
2. 在"安全性"部分找到被阻止的 antihook 程序
3. 点击"仍要打开"
4. 在弹出的对话框中确认"打开"

### 步骤 2: 完成安装

程序运行后会自动：
1. ✅ 将自身复制到 `~/.local/bin/Antihub/`
2. ✅ 注册 `kiro://` 和 `anti://` 协议处理器
3. ✅ 添加到系统 PATH（需要重启终端生效）
4. ✅ 显示"Hooked successfully!"提示

## 🔧 如何 Hook Kiro 协议

### Hook 原理

AntiHook 会接管 `kiro://` 协议的处理，当您在浏览器中点击 `kiro://` 链接时：

1. **原始流程**：浏览器 → Kiro 官方应用
2. **Hook 后**：浏览器 → AntiHook → 您的后端服务器

### 协议处理流程

```
用户点击授权
    ↓
浏览器打开: https://prod.us-east-1.auth.desktop.kiro.dev/login...
    ↓
用户完成 Google 授权
    ↓
浏览器重定向: kiro://kiro.kiroAgent/authenticate-success?code=xxx
    ↓
macOS 调用 AntiHook 处理 kiro:// 协议
    ↓
AntiHook 将完整 URL 发送到: https://api.mortis.edu.kg/api/kiro/oauth/callback
    ↓
后端处理登录逻辑
    ↓
显示"Login successful!"
```

## 🧪 测试 Hook 是否生效

### 方法 1: 直接测试协议

```bash
# 手动触发 kiro:// 协议
open "kiro://test-callback?code=test123&state=test456"
```

**预期结果**：
- 终端输出日志：`Received kiro:// callback: kiro://test-callback?code=test123...`
- 弹出对话框：显示登录状态

### 方法 2: 检查协议注册

```bash
# 查看已注册的协议处理器
defaults read ~/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist | grep -A 5 "kiro"
```

### 方法 3: 查看已安装的文件

```bash
# 检查程序是否安装到正确位置
ls -la ~/.local/bin/Antihub/antihook

# 检查是否可执行
~/.local/bin/Antihub/antihook --help
```

## 📝 使用说明

### 正常使用流程

1. **打开需要登录的应用或网站**
2. **点击"使用 Kiro 登录"按钮**
3. **浏览器会打开授权页面**
4. **完成 Google 授权**
5. **浏览器会重定向到 `kiro://` 协议**
6. **AntiHook 自动接管处理**
7. **显示"Login successful!"**

### 查看调试日志

程序会在终端输出详细日志：

```bash
# 在终端中运行（可以看到日志）
~/.local/bin/Antihub/antihook "kiro://your-callback-url"

# 日志示例：
Received kiro:// callback: kiro://kiro.kiroAgent/authenticate-success?code=xxx
Posting to: https://api.mortis.edu.kg/api/kiro/oauth/callback
Request body: {"callback_url":"kiro://..."}
Response status: 200
Response body: {...}
Login successful!
```

## 🔄 重新安装或更新

```bash
# 1. 重新运行安装
./build/antihook-darwin-amd64

# 2. 如果需要清理旧版本
rm -rf ~/.local/bin/Antihub/antihook

# 3. 重新安装
./build/antihook-darwin-amd64
```

## ⚠️ 常见问题

### Q1: 点击 kiro:// 链接没有反应？

**解决方案**：
```bash
# 1. 重新注册协议
./build/antihook-darwin-amd64

# 2. 重启浏览器

# 3. 检查程序是否有执行权限
chmod +x ~/.local/bin/Antihub/antihook
```

### Q2: 提示"无法打开，因为无法验证开发者"？

**解决方案**：
```bash
# 移除 macOS 的隔离属性
xattr -d com.apple.quarantine ./build/antihook-darwin-amd64

# 或者通过系统设置允许
# 系统设置 → 隐私与安全性 → 点击"仍要打开"
```

### Q3: 如何恢复原始的 Kiro 处理器？

**解决方案**：
```bash
# 卸载 AntiHook 的协议注册
# 方法1: 删除程序（系统会自动清理）
rm -rf ~/.local/bin/Antihub/

# 方法2: 重新安装 Kiro 官方应用
# Kiro 会重新注册协议处理器
```

### Q4: 登录时一直卡在授权页面？

**解决方案**：
1. **查看终端日志**，确认是否收到回调
2. **检查后端地址**是否正确：
   ```bash
   # 查看当前配置
   strings ~/.local/bin/Antihub/antihook | grep "mortis.edu.kg"
   ```
3. **测试后端连接**：
   ```bash
   curl -v https://api.mortis.edu.kg/api/kiro/oauth/callback
   ```
4. 参考 [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md) 详细排查

## 🎯 高级配置

### 使用自定义后端地址

如果需要临时使用不同的后端地址：

```bash
# 设置环境变量
export KIRO_SERVER_URL="https://your-server.com"

# 然后触发协议
open "kiro://your-callback"
```

### 查看程序版本信息

```bash
strings ~/.local/bin/Antihub/antihook | grep -E "BuildVersion|BuildTime"
```

## 🚀 完整测试流程

```bash
# 1. 安装
./build/antihook-darwin-amd64

# 2. 验证安装
ls -la ~/.local/bin/Antihub/antihook

# 3. 测试协议（会看到详细日志）
~/.local/bin/Antihub/antihook "kiro://test?code=abc123"

# 4. 查看日志输出，确认是否正常工作
```

## 📞 获取帮助

如果遇到问题：
1. 查看终端的详细日志输出
2. 参考 [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md)
3. 检查 `~/.local/bin/Antihub/` 目录权限
4. 确认后端服务器地址可访问

---

**当前配置**：
- KIRO_SERVER_URL: `https://api.mortis.edu.kg`
- BACKEND_URL: `https://tunnel.mortis.edu.kg`
- 构建版本: `1.0.0`