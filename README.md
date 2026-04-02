# LocalAIHub

<p align="center">
  <img src="logo.png" alt="LocalAIHub Logo" width="128" height="128" />
</p>

<p align="center">
  本地 AI 中转网关与管理后台，提供多协议 AI 请求转发、统一配置管理和运营审计能力。
</p>

## 项目简介

LocalAIHub 是一个自托管的 AI 网关解决方案，允许你：

- 统一管理 OpenAI、Anthropic、Gemini 协议上游
- 为不同应用或客户分发独立的 API Key
- 管理 API Key 配额、过期时间和模型权限
- 监控请求量、延迟、Token 消耗和模型分布
- 导入导出核心配置，并保留操作审计记录

说明：Gemini 协议入口已接入网关，当前以 `generateContent` / `streamGenerateContent` 为主；模型管理侧也支持配置 Gemini 协议模型。

## 技术架构

```text
LocalAIHub/
├── LocalAIHub_GO/        # Go 后端服务
└── LocalAIHub_Admin/     # Next.js 管理后台
```

- **后端**：Go + 标准库 HTTP + MySQL，提供 AI 网关和管理 API
- **前端**：Next.js 16 + React 19 + TypeScript + Tailwind CSS

## 功能特性

### AI 网关

- OpenAI 兼容协议：`/proxy/openai/v1/*`
- Anthropic 消息协议：`/proxy/anthropic/v1/messages`
- Gemini 生成协议：`/proxy/gemini/v1beta/models/*:generateContent`
- 自动故障转移、路由切换、Provider Key 健康上报
- 请求日志、Token 统计、客户端配额累计

### 管理后台

- 仪表盘：请求量、Token、模型分布、API Key 维度统计
- 上游管理：Provider、Provider Key、连接测试、优先级管理
- 虚拟模型：模型定义、协议族、上游绑定、排序与测试
- 路由管理：手动切换、解锁、故障转移规则
- API Key：创建、状态控制、使用示例、每日/每月配额设置
- 日志中心：请求日志、审计日志、详情查看、CSV 导出
- API 测试工具：选 Key、选模型、直接发起测试请求
- 配置导入导出：支持全量或按模块导出、导入和试运行
- 快捷流程与使用教程：帮助快速完成初始化配置

## 快速开始

### 后端启动

```bash
cd LocalAIHub_GO
go mod tidy
go run ./cmd/server
```

服务默认监听 `0.0.0.0:8080`。

### 前端启动

```bash
cd LocalAIHub_Admin
npm install
npm run dev
```

服务默认运行在 `http://localhost:3000`。

### 首次使用

1. 访问管理后台 `http://localhost:3000`
2. 使用默认账号登录
3. 在「上游管理」配置 AI 服务商和 Provider Key
4. 在「虚拟模型」创建模型并绑定上游
5. 在「API Key」创建客户端密钥并配置可用模型/配额
6. 在「API 测试」验证调用链路
7. 用生成的客户端密钥调用网关接口

## 调用示例

### OpenAI 兼容调用

```bash
curl -X POST http://localhost:8080/proxy/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

### Gemini 调用

```bash
curl -X POST "http://localhost:8080/proxy/gemini/v1beta/models/gemini-2.0-flash:generateContent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "contents": [
      {
        "parts": [{"text": "你好"}]
      }
    ]
  }'
```

## 文档

- [后端 README](./LocalAIHub_GO/README.md)
- [前端 README](./LocalAIHub_Admin/README.md)
- [文档索引](./docs/README.md)
- [更新日志](./CHANGELOG.md)

## 许可证

MIT
