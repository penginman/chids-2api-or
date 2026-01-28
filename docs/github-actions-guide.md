# GitHub Actions 自动注册使用指南

## 概述

本项目包含一个 GitHub Actions 工作流,可以自动注册 Orchids 账号并推送到远程 API 服务器。

## 功能特性

- 自动批量注册 Orchids 账号
- 将注册成功的账号推送到指定的 API 地址
- 支持定时触发和手动触发
- 完整的日志输出和错误处理

## 配置步骤

### 1. 配置 GitHub Secrets

在你的 GitHub 仓库中,进入 `Settings` → `Secrets and variables` → `Actions`,添加以下 Secrets:

| Secret 名称 | 说明 | 是否必需 | 示例 |
|------------|------|---------|------|
| `PUSH_API_URL` | 推送目标 API 地址 | 必需 | `https://your-server.com/api/accounts` |
| `PUSH_API_USER` | Basic Auth 用户名 | 必需 | `admin` |
| `PUSH_API_PASS` | Basic Auth 密码 | 必需 | `your_password` |
| `SOCKS_PROXY` | SOCKS5 代理地址 | 可选 | `socks5://proxy.example.com:1080` |

### 2. 触发方式

#### 方式一: 定时触发 (Cron)

工作流默认配置为每天 UTC 00:00 (北京时间早上 8:00) 自动运行。

可以修改 [.github/workflows/auto-register.yml](.github/workflows/auto-register.yml) 中的 cron 表达式:

```yaml
schedule:
  - cron: '0 0 * * *'  # 每天 UTC 00:00
```

常用 cron 表达式:
- `0 */6 * * *` - 每 6 小时执行一次
- `0 0 * * 1` - 每周一执行
- `0 0 1 * *` - 每月 1 号执行

#### 方式二: 手动触发

1. 进入 GitHub 仓库的 `Actions` 标签页
2. 选择 `Auto Register and Push` 工作流
3. 点击 `Run workflow` 按钮
4. 可以自定义参数:
   - **注册数量** (count): 默认 1
   - **并发线程数** (workers): 默认 2
5. 点击 `Run workflow` 开始执行

### 3. 查看执行结果

1. 进入 `Actions` 标签页
2. 点击对应的工作流运行记录
3. 查看详细日志输出,包括:
   - 注册进度和结果
   - 推送结果和状态码
   - 成功/失败统计

## 本地测试 (可选)

### 编译 CLI 工具

```bash
go build -o register-cli ./cmd/register-cli
```

### 运行测试

```bash
# 设置环境变量
export PUSH_API_URL="http://localhost:3002/api/accounts"
export PUSH_API_USER="admin"
export PUSH_API_PASS="admin123"

# 可选: 设置 SOCKS 代理
export SOCKS_PROXY="socks5://127.0.0.1:1080"

# 运行注册
./register-cli --count=1 --workers=1 --headless=true
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|-----|------|-------|
| `--count` | 注册数量 | 5 |
| `--workers` | 并发线程数 | 2 |
| `--headless` | 无头模式 | true |
| `--push-url` | 推送 API 地址 | 从环境变量读取 |
| `--push-user` | 推送用户名 | 从环境变量读取 |
| `--push-pass` | 推送密码 | 从环境变量读取 |
| `--socks-proxy` | SOCKS5 代理地址 | 从环境变量读取 (可选) |

## 推送数据格式

推送到远程 API 的数据格式 (POST `/api/accounts`):

```json
{
  "name": "Auto-xxx",
  "email": "xxx@domain.com",
  "client_cookie": "...",
  "client_uat": "...",
  "session_id": "...",
  "user_id": "...",
  "project_id": "280b7bae-cd29-41e4-a0a6-7f603c43b607",
  "agent_mode": "claude-opus-4.5",
  "weight": 1,
  "enabled": true
}
```

## 认证方式

推送 API 使用 **HTTP Basic Authentication**:

```
Authorization: Basic base64(username:password)
```

## SOCKS 代理配置 (可选)

如果注册或推送过程中遇到网络问题,可以配置 SOCKS5 代理。

### 配置方法

在 GitHub Secrets 中添加:

```
SOCKS_PROXY=socks5://proxy.example.com:1080
```

### 代理格式

- 协议: 仅支持 `socks5://`
- 格式: `socks5://host:port`
- 示例:
  - `socks5://127.0.0.1:1080`
  - `socks5://proxy.example.com:1080`

### 注意事项

- 代理配置是可选的,如果不配置则使用直连
- 代理会同时应用于注册和推送过程
- 如果代理配置错误,程序会回退到直连模式并给出警告
- 目前仅支持无认证的 SOCKS5 代理

## 故障排查

### 注册失败

常见原因:
- 临时邮箱服务不可用
- 验证码获取超时
- Cloudflare Turnstile 验证失败
- 网络连接问题

解决方法:
- 检查 GitHub Actions 日志中的详细错误信息
- 确认 Chrome 浏览器正确安装
- 调整并发数 (workers) 避免并发过高
- 尝试配置 SOCKS 代理 (如果是网络问题)

### 推送失败

常见原因:
- API 地址配置错误
- 认证信息不正确
- 远程服务器不可达
- 网络超时

解决方法:
- 检查 GitHub Secrets 配置是否正确
- 确认远程服务器正常运行
- 查看推送响应的 HTTP 状态码和错误信息
- 如果是网络超时,尝试配置 SOCKS 代理

### 工作流执行失败

常见原因:
- Go 编译错误
- Chrome 安装失败
- 依赖下载失败

解决方法:
- 查看 Actions 日志中的具体错误步骤
- 确认 `go.mod` 依赖正确
- 检查网络连接

## 安全建议

1. **不要在代码中硬编码密钥**
   - 所有敏感信息都应通过 GitHub Secrets 配置

2. **使用 HTTPS 推送地址**
   - 确保 API 地址使用 HTTPS 协议保护传输安全

3. **定期更新密钥**
   - 定期更换 Basic Auth 密码

4. **限制访问权限**
   - 确保只有授权人员可以修改 GitHub Secrets

5. **监控执行日志**
   - 定期检查 Actions 日志,发现异常及时处理

## 日志说明

工作流会输出详细的日志信息:

```
========================================
自动注册与推送工具启动
========================================
配置信息:
  注册数量: 1
  并发线程: 2
  无头模式: true
  推送地址: https://***
  推送用户: a***n
========================================

开始批量注册...
[Worker-1][任务#1] 使用 mail.tm 生成邮箱: xxx@xxx.com
...

========================================
批量注册完成!
========================================
总数: 1
成功: 1
失败: 0
耗时: 2m30s
========================================

开始推送注册账号到远程 API...
[推送 1/1] 获取账号信息: xxx@xxx.com
[推送 1/1] 成功: xxx@xxx.com

========================================
推送完成!
========================================
推送总数: 1
推送成功: 1
推送失败: 0
========================================

所有任务完成!
```

## 自定义扩展

### 修改注册数量

编辑 [.github/workflows/auto-register.yml](.github/workflows/auto-register.yml):

```yaml
env:
  REGISTER_COUNT: 5  # 修改为你需要的数量
  REGISTER_WORKERS: 3
```

### 修改定时时间

编辑 cron 表达式:

```yaml
schedule:
  - cron: '0 2 * * *'  # 改为 UTC 02:00
```

### 添加通知

可以在工作流中添加通知步骤 (需要配置相应的 Secrets):

```yaml
- name: Send notification
  if: failure()
  run: |
    # 发送失败通知
    curl -X POST $WEBHOOK_URL -d "Registration failed"
```

## 相关文档

- [GitHub Actions 文档](https://docs.github.com/cn/actions)
- [项目 API 文档](docs/api-reference.md)
- [配置说明](docs/configuration.md)

## 支持

如有问题,请查看:
- GitHub Actions 运行日志
- 项目 Issues 列表
- 相关文档说明
