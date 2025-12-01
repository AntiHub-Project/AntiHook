# Auto login for AntiHub beta

## 协议支持

本项目支持两种协议：

### 1. kiro:// 协议
用于 Kiro 应用的回调处理。

### 2. anti:// 协议
用于 OAuth 授权登录流程。

## anti:// 协议使用说明

### URL 格式

```
anti://antigravity?identity=<token>&is_shared=<0|1>
```

### 参数说明

- **identity**: OAuth 授权令牌（必需）
  - 格式：`sk-xxxxx` 或 `Bearer sk-xxxxx`（程序会自动添加 `Bearer ` 前缀）
  - 示例：`sk-qG4ptW4ktJQ2ANStka72lopuSbKQKYeuaMaH5nt8uPhspmR`

- **is_shared**: 是否共享（可选参数）
  - 值：`0` 或 `1`
  - 默认：`0`

### URL 示例

1. **标准格式**：
```
anti://antigravity?identity=sk-qG4ptW4ktJQ2ANStka72lopuSbKQKYeuaMaH5nt8uPhspmR&is_shared=1
```

2. **不带 is_shared 参数**：
```
anti://antigravity?identity=sk-qG4ptW4ktJQ2ANStka72lopuSbKQKYeuaMaH5nt8uPhspmR
```

3. **identity 带 Bearer 前缀**：
```
anti://antigravity?identity=Bearer%20sk-qG4ptW4ktJQ2ANStka72lopuSbKQKYeuaMaH5nt8uPhspmR&is_shared=1
```

### 工作流程

1. 用户点击 `anti://` 链接
2. 程序解析 identity 和 is_shared 参数
3. 向 `${serverurl}/api/oauth/authorize` 发送 POST 请求
4. 获取 OAuth 授权 URL
5. 在本地启动 HTTP 服务器监听 42532 端口
6. 打开浏览器访问授权 URL
7. 用户完成授权后，浏览器重定向到 `http://localhost:42532/oauth-callback`
8. 程序接收回调并向 `${serverurl}/api/oauth/callback/manual` 发送回调信息
9. 显示成功弹窗

### 服务器配置

默认服务器地址：`http://localhost:8045`

可通过环境变量 `KIRO_SERVER_URL` 自定义服务器地址：
```bash
set KIRO_SERVER_URL=https://your-server.com
```

### API 端点

1. **授权请求**：`POST ${serverurl}/api/oauth/authorize`
   - Header: `Authorization: Bearer <token>`
   - Body: `{"is_shared": 0 或 1}`
   - Response: 
     ```json
     {
       "success": true,
       "data": {
         "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?...",
         "state": "uuid",
         "expires_in": 300
       }
     }
     ```

2. **回调处理**：`POST ${serverurl}/api/oauth/callback/manual`
   - Header: `Authorization: Bearer <token>`
   - Body: `{"callback_url": "http://localhost:42532/oauth-callback?state=...&code=..."}`

## 安装

运行程序将自动：
1. 将程序复制到 `%LOCALAPPDATA%\Antihub\antihook.exe`
2. 注册 `kiro://` 和 `anti://` 协议处理器
3. 添加到系统 PATH

## 恢复原始协议

```bash
antihook.exe -recover
```

## 构建

```bash
go build -o antihook.exe .
```

## 依赖

- Go 1.21+
- golang.org/x/sys/windows/registry