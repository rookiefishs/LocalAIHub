# LocalAIHub

<p align="center">
  <img src="logo.png" alt="LocalAIHub Logo" width="128" height="128" />
</p>

<p align="center">
  本地 AI 中转网关与管理后台，提供 OpenAI 兼容协议的 AI 请求转发和管理功能。
</p>

## 项目简介

LocalAIHub 是一个自托管的 AI 网关解决方案，允许你：

- 统一管理多个 AI 服务商（OpenAI、Anthropic、Gemini 等）
- 为不同应用或客户分发独立的 API Key
- 监控请求量和 Token 消耗
- 自动故障转移和负载均衡
- 完整的审计日志

## 技术架构

```
LocalAIHub/
├── LocalAIHub_GO/        # Go 后端 (Gin + GORM)
└── LocalAIHub_Admin/     # Next.js 管理前台
```

- **后端**：Go + Gin + SQLite，提供 AI 网关和管理 API
- **前台**：Next.js 16 + React 19 + TypeScript + Tailwind CSS

## 功能特性

### AI 网关
- OpenAI 兼容协议 (/proxy/openai/v1/*)
- 流式和非流式响应
- 自动故障转移
- Token 用量统计

### 管理后台
- 仪表盘：请求量、Token 消耗、模型分布统计
- 上游管理：AI 服务商配置
- 虚拟模型：模型定义和参数配置
- 路由管理：路由策略和故障转移规则
- API Key：客户端密钥管理
- 日志中心：请求日志和审计日志
- 使用教程：新手指南

## 快速开始

### 后端启动

```bash
cd LocalAIHub_GO
go mod tidy
go run ./cmd/server
```

服务默认监听 `0.0.0.0:8080`

### 前端启动

```bash
cd LocalAIHub_Admin
npm install
npm run dev
```

服务默认运行在 `http://localhost:3000`

### 首次使用

1. 访问管理后台 `http://localhost:3000`
2. 使用默认账号登录（首次启动时创建）
3. 在「上游管理」添加 AI 服务商
4. 在「虚拟模型」创建模型并绑定上游
5. 在「API Key」创建客户端密钥
6. 使用密钥调用 AI 接口

## 调用示例

```bash
curl -X POST http://localhost:8080/proxy/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

## 文档

- [后端 README](./LocalAIHub_GO/README.md)
- [前端 README](./LocalAIHub_Admin/README.md)
- [更新日志](./CHANGELOG.md)

## 许可证

MIT
