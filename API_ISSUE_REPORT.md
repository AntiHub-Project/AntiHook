# API 端点问题诊断报告

## 🔴 问题描述

Hook 已成功安装并能正确拦截 `kiro://` 协议，但后端 API 返回 404 错误。

## 📊 实际日志

```
=== 2025-12-08 15:02:22 ===
Received kiro:// callback: kiro://kiro.kiroAgent/authenticate-success?code=f5488d79-d050-453c-aa33-a24c8d25ca71&state=b1b52ff8-e8ee-4468-bd50-7b70b496fc87
Posting to: https://api.mortis.edu.kg/api/kiro/oauth/callback
Request body: {"callback_url":"kiro://kiro.kiroAgent/authenticate-success?code=...&state=..."}
Response status: 404
Response body: {"detail":"Not Found"}
```

## ✅ 工作正常的部分

1. **协议拦截** ✓ - kiro:// 协议被正确拦截
2. **回调接收** ✓ - 收到了完整的授权回调 URL
3. **配置注入** ✓ - 后端地址正确配置为 `https://api.mortis.edu.kg`
4. **网络请求** ✓ - HTTP 请求成功发送到后端
5. **日志记录** ✓ - 详细日志已记录到文件

## ❌ 问题所在

**后端 API 端点不存在**：`/api/kiro/oauth/callback` 返回 404

已测试的路径（均返回 404）：
- `/api/kiro/oauth/callback` ❌
- `/oauth/callback` ❌
- `/api/oauth/kiro/callback` ❌

## 🔍 需要确认的信息

请联系后端开发人员确认以下信息：

### 1. 正确的 API 端点
- 完整的 API URL 是什么？
- 是否需要版本号（如 `/v1/`, `/v2/`）？
- 路径的准确格式？

### 2. 请求格式
- 当前发送的格式：
  ```json
  {
    "callback_url": "kiro://kiro.kiroAgent/authenticate-success?code=xxx&state=yyy"
  }
  ```
- 后端期望的格式是否正确？
- 是否需要其他参数？

### 3. 认证要求
- 是否需要 API Key？
- 是否需要 Bearer Token？
- 是否需要其他 HTTP Headers？

### 4. 域名确认
- `https://api.mortis.edu.kg` 是否是正确的后端域名？
- 是否应该使用 `https://tunnel.mortis.edu.kg`？

## 🛠️ 可能的解决方案

### 方案 1：使用正确的 API 路径

如果后端开发人员提供了正确的路径，例如：
```
https://api.mortis.edu.kg/api/v1/kiro/callback
```

**修改步骤：**

1. 编辑 `main.go` 第 204 行：
```go
// 修改前
apiURL := serverURL + "/api/kiro/oauth/callback"

// 修改后
apiURL := serverURL + "/api/v1/kiro/callback"  // 使用正确的路径
```

2. 重新构建：
```bash
./build.sh darwin
./build/antihook-darwin-amd64
```

### 方案 2：使用不同的服务器地址

如果应该使用 `tunnel.mortis.edu.kg` 而不是 `api.mortis.edu.kg`：

1. 修改 `.build.config`：
```bash
SERVER_URL="https://tunnel.mortis.edu.kg"
```

2. 重新构建并安装。

### 方案 3：添加认证信息

如果需要 API Key 或其他认证：

修改 `main.go` 的 `postCallback` 函数，添加必要的 headers：
```go
req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer YOUR_API_KEY")  // 添加认证
req.Header.Set("X-API-Key", "YOUR_KEY")  // 或其他方式
```

## 📞 联系后端团队

请向后端团队提供以下信息：

**问题**：Kiro OAuth 回调接口返回 404

**详细信息**：
- 请求 URL: `https://api.mortis.edu.kg/api/kiro/oauth/callback`
- 请求方法: POST
- Content-Type: application/json
- 请求体:
  ```json
  {
    "callback_url": "kiro://kiro.kiroAgent/authenticate-success?code=xxx&state=yyy"
  }
  ```
- 响应: `{"detail":"Not Found"}`

**需要确认**：
1. 正确的 API 端点路径是什么？
2. 请求格式是否正确？
3. 是否需要认证信息？

## 📝 测试命令

获得正确信息后，可以使用以下命令测试：

```bash
# 测试 API 端点
curl -X POST https://api.mortis.edu.kg/CORRECT_PATH \
  -H "Content-Type: application/json" \
  -d '{"callback_url":"kiro://test"}'

# 如果需要认证
curl -X POST https://api.mortis.edu.kg/CORRECT_PATH \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"callback_url":"kiro://test"}'
```

## 🎯 下一步行动

1. **立即**：联系后端开发人员获取正确的 API 信息
2. **获得信息后**：按照上述方案修改配置
3. **修改后**：重新构建并测试
4. **测试**：查看 `~/.config/antihook/kiro.log` 确认成功

## 📂 相关文件

- 日志文件: `~/.config/antihook/kiro.log`
- 配置文件: `.build.config`
- 主程序: `main.go` (第 189-220 行)
- 构建产物: `build/antihook-darwin-amd64`

---

**总结**：Hook 功能完全正常，唯一的问题是后端 API 端点配置不正确。需要后端团队提供正确的 API 信息。