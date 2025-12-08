# AntiHook 故障排查指南

## 问题：登录授权页面卡住

如果在使用 `anti://` 协议登录时，浏览器一直停留在授权页面无法完成登录，请按以下步骤排查：

### 1. 检查回调服务器端口

OAuth 回调需要在本地 `42532` 端口启动服务器。

**检查端口是否被占用：**

```bash
# macOS/Linux
lsof -i :42532

# Windows
netstat -ano | findstr :42532
```

**如果端口被占用，请关闭占用进程：**

```bash
# macOS/Linux
kill <PID>

# Windows
taskkill /PID <PID> /F
```

### 2. 检查防火墙设置

确保防火墙允许本地 `42532` 端口的连接。

**macOS：**
- 系统设置 → 网络 → 防火墙 → 防火墙选项
- 确保允许 antihook 的传入连接

**Windows：**
- 控制面板 → Windows Defender 防火墙 → 允许应用通过防火墙
- 添加 antihook.exe 到允许列表

### 3. 检查后端 API 地址

确认后端地址是否正确配置并且可访问。

**当前配置的地址：**
```bash
KIRO_SERVER_URL: https://api.mortis.edu.kg
BACKEND_URL: https://tunnel.mortis.edu.kg
```

**测试后端连接：**

```bash
# 测试授权接口
curl -v -X POST https://tunnel.mortis.edu.kg/api/plugin-api/oauth/authorize \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"is_shared": 0}'
```

如果返回错误，说明后端地址配置有问题。

### 4. 查看详细日志

最新版本已添加调试日志。运行程序时会在终端输出：

```
Received OAuth callback: http://localhost:42532/oauth-callback?code=xxx&state=yyy
```

**如果没有看到这条日志：**
- 说明浏览器没有成功回调到本地服务器
- 检查浏览器是否阻止了 localhost 重定向
- 尝试在浏览器中手动访问 `http://localhost:42532/oauth-callback`

### 5. 常见问题解决方案

#### 问题 A：浏览器显示"无法连接到 localhost:42532"

**原因：** 回调服务器未启动或启动失败

**解决方案：**
1. 确认没有其他程序占用 42532 端口
2. 以管理员/root 权限运行程序
3. 检查系统日志查看错误信息

#### 问题 B：浏览器完成授权但没有弹出"登录成功"提示

**原因：** 回调 URL 构造不正确或后端接口返回错误

**解决方案：**
1. 检查终端输出的回调 URL 是否完整
2. 确认后端 `/api/plugin-api/oauth/callback` 接口正常
3. 查看后端日志了解详细错误

#### 问题 C：显示"OAuth timeout"错误

**原因：** 授权超时（默认超时时间由后端 `expires_in` 字段控制）

**解决方案：**
1. 加快授权操作速度
2. 联系后端管理员增加超时时间
3. 检查网络连接是否稳定

### 6. 手动测试 OAuth 流程

可以手动测试完整的 OAuth 流程：

```bash
# 1. 获取授权 URL
curl -X POST https://tunnel.mortis.edu.kg/api/plugin-api/oauth/authorize \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"is_shared": 0}'

# 响应示例：
# {
#   "success": true,
#   "data": {
#     "auth_url": "https://...",
#     "state": "...",
#     "expires_in": 300
#   }
# }

# 2. 在浏览器中打开 auth_url

# 3. 完成授权后，浏览器会重定向到：
# http://localhost:42532/oauth-callback?code=xxx&state=yyy

# 4. 程序会自动调用回调接口：
curl -X POST https://tunnel.mortis.edu.kg/api/plugin-api/oauth/callback \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"callback_url": "http://localhost:42532/oauth-callback?code=xxx&state=yyy"}'
```

### 7. 版本更新说明

**v1.0.1 修复内容：**
- ✅ 修复回调 URL 构造问题（使用 `RequestURI()` 而不是 `String()`）
- ✅ 增加服务器启动等待时间（100ms → 500ms）
- ✅ 添加调试日志输出
- ✅ 改进错误处理

**如果问题仍然存在：**

请提供以下信息以便进一步诊断：
1. 操作系统和版本
2. 使用的是哪个协议（`kiro://` 还是 `anti://`）
3. 终端输出的完整日志
4. 浏览器控制台的错误信息（F12 → Console）
5. 后端服务器的日志

### 8. 临时解决方案

如果急需使用，可以尝试：

1. **使用环境变量覆盖配置：**
   ```bash
   export BACKEND_URL="https://your-working-backend.com"
   ./antihook "anti://?identity=YOUR_TOKEN&is_shared=0"
   ```

2. **直接使用 kiro:// 协议（如果适用）：**
   ```bash
   ./antihook "kiro://your-callback-url"
   ```

3. **检查是否有代理或 VPN 干扰：**
   - 临时关闭代理/VPN
   - 添加 localhost 到代理例外列表

## 联系支持

如果以上方法都无法解决问题，请提供详细的错误信息和日志，我们会尽快协助解决。