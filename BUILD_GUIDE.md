# AntiHook 构建指南

本指南说明如何配置和构建 AntiHook，使打包后的程序自动使用您的后端地址。

## 🎯 快速开始

### 1. 创建配置文件

```bash
# 复制配置模板
cp .build.config.example .build.config

# 编辑配置文件
vim .build.config  # 或使用您喜欢的编辑器
```

### 2. 配置您的后端地址

在 `.build.config` 文件中设置：

```bash
# Kiro 服务器地址
SERVER_URL="https://your-kiro-server.com"

# 后端服务器地址
BACKEND_URL="https://your-backend-server.com"
```

### 3. 构建程序

```bash
# 构建所有平台
./build.sh all

# 或构建特定平台
./build.sh darwin    # macOS
./build.sh windows   # Windows
./build.sh linux     # Linux
```

构建完成后，生成的二进制文件会自动包含您配置的后端地址！

## 📋 配置优先级

AntiHook 支持三种配置方式，优先级从高到低：

1. **运行时环境变量**（最高优先级）
   ```bash
   export KIRO_SERVER_URL="https://runtime-server.com"
   export BACKEND_URL="https://runtime-backend.com"
   ./antihook
   ```

2. **构建时配置文件**（推荐用于发布）
   ```bash
   # .build.config 文件
   SERVER_URL="https://build-server.com"
   BACKEND_URL="https://build-backend.com"
   ```

3. **构建时环境变量**
   ```bash
   SERVER_URL="https://env-server.com" ./build.sh all
   ```

4. **默认值**（开发环境）
   - SERVER_URL: `http://localhost:8045`
   - BACKEND_URL: `http://localhost:8008`

## 🔧 使用场景

### 场景一：本地开发

不需要任何配置，直接构建即可使用 localhost：

```bash
go build -o antihook .
./antihook
```

### 场景二：生产环境发布

1. 创建 `.build.config` 配置生产环境地址
2. 构建发布版本
3. 分发给用户

```bash
# .build.config
SERVER_URL="https://prod-kiro.example.com"
BACKEND_URL="https://prod-api.example.com"

# 构建
./build.sh all

# 构建产物在 build/ 目录
ls -lh build/
```

### 场景三：多环境构建

为不同环境构建不同版本：

```bash
# 测试环境
SERVER_URL="https://test-api.com" \
BACKEND_URL="https://test-backend.com" \
./build.sh darwin
mv build/antihook-darwin-amd64 build/antihook-darwin-amd64-test

# 生产环境
SERVER_URL="https://prod-api.com" \
BACKEND_URL="https://prod-backend.com" \
./build.sh darwin
mv build/antihook-darwin-amd64 build/antihook-darwin-amd64-prod
```

## 🔍 验证构建配置

构建完成后，可以验证配置是否正确注入：

```bash
# macOS/Linux
strings build/antihook-darwin-amd64 | grep -E "https?://"

# Windows (使用 PowerShell)
Select-String -Path build/antihook-windows-amd64.exe -Pattern "https?://" -AllMatches
```

您应该能看到配置的 URL 地址。

## 📝 注意事项

1. **`.build.config` 文件已添加到 `.gitignore`**
   - 不会被提交到版本控制
   - 可以安全地包含敏感信息

2. **配置文件格式**
   - 使用 Bash 变量格式
   - URL 必须包含协议（http:// 或 https://）
   - 不要在末尾添加斜杠

3. **构建信息**
   - 构建时会自动注入构建时间和版本号
   - 可用于追踪和调试

## 🚀 CI/CD 集成

在 CI/CD 流水线中使用环境变量：

```yaml
# GitHub Actions 示例
- name: Build AntiHook
  env:
    SERVER_URL: ${{ secrets.PROD_SERVER_URL }}
    BACKEND_URL: ${{ secrets.PROD_BACKEND_URL }}
  run: ./build.sh all
```

## 💡 最佳实践

1. **开发环境**：不使用配置文件，使用默认 localhost
2. **测试环境**：使用 `.build.config` 配置测试服务器
3. **生产环境**：使用 CI/CD 环境变量或专门的配置文件
4. **多环境**：为每个环境维护单独的配置文件

## ❓ 常见问题

**Q: 构建后的程序还能通过环境变量修改地址吗？**
A: 可以！运行时环境变量的优先级最高，会覆盖构建时的配置。

**Q: 如何验证程序使用的是哪个地址？**
A: 可以通过抓包工具（如 Charles、Wireshark）查看实际请求的地址。

**Q: 可以不同平台使用不同配置吗？**
A: 可以，在构建特定平台前临时修改 `.build.config` 或使用环境变量。

**Q: 配置文件支持注释吗？**
A: 支持！使用 `#` 开头的行会被忽略。