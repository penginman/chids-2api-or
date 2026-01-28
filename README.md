# Orchids-2api 文档

## 项目简介

**Orchids-2api** (orchids-api) 是一个 Go 语言编写的 API 代理服务器，提供多账号管理与负载均衡代理功能，兼容 Claude API 格式的请求转发。

### 核心功能

- 多账号管理与负载均衡代理
- 兼容 Claude API 格式的请求转发
- 将请求代理到 Orchids 后端服务
- 提供 Web 管理界面


## 文档目录

| 文档 | 描述 |
|------|------|
| [架构设计](./docs/architecture.md) | 目录结构、核心组件、请求流程、数据模型 |
| [API 接口](./docs/api-reference.md) | 所有端点列表、请求/响应格式、认证说明 |
| [部署指南](./docs/deployment.md) | Docker 构建、本地开发、生产部署 |
| [配置说明](./docs/configuration.md) | 环境变量、配置文件格式 |
| [GitHub Actions 使用指南](./docs/github-actions-guide.md) | 自动注册与推送配置、使用说明 |

## 快速开始

```bash
# 本地开发
go mod download
go run ./cmd/server/main.go

# Docker 部署
./build.sh
docker compose up -d
```

## 主要特性

1. **多账号管理** - 支持添加、编辑、删除多个 Orchids 账号
2. **负载均衡** - 加权随机算法分配请求
3. **故障转移** - 账号失败时自动切换
4. **模型映射** - 透明映射 Claude 模型到上游模型
5. **工具调用** - 完整支持 Claude Tool Use
6. **流式响应** - SSE 实时响应
7. **Token 计数** - 估算输入/输出 Token
8. **调试日志** - 详细的请求/响应日志
9. **管理界面** - Web UI 管理账号
10. **导入导出** - 账号配置备份恢复
11. **自动注册** - GitHub Actions 自动注册账号并推送

## GitHub Actions 自动注册

本项目支持通过 GitHub Actions 自动注册 Orchids 账号并推送到远程服务器。

### 快速开始

1. 配置 GitHub Secrets:
   - `PUSH_API_URL`: 推送目标 API 地址
   - `PUSH_API_USER`: Basic Auth 用户名
   - `PUSH_API_PASS`: Basic Auth 密码

2. 触发方式:
   - 定时触发: 每天自动运行 (可配置)
   - 手动触发: 在 Actions 页面手动运行

3. 详细说明请参考: [GitHub Actions 使用指南](./docs/github-actions-guide.md)
