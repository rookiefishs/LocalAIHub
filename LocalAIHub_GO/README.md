# LocalAIHub_GO

<p align="center">
  <img src="../logo.png" alt="LocalAIHub Logo" width="64" height="64" />
</p>

<p align="center">
  LocalAIHub Go 后端服务，提供 AI 网关、审计日志和后台管理 API。
</p>

## 功能特性

- **AI 网关**：支持 OpenAI、Anthropic、Gemini 协议转发
- **路由管理**：模型路由、故障转移、Provider 熔断与恢复
- **上游适配**：支持 Provider、Provider Key、连接探测与健康检查
- **日志审计**：请求日志、审计日志、详情查询、CSV 导出、Token 统计
- **配额控制**：API Key 每日/每月请求与 Token 限制，超限自动禁用
- **配置工具**：配置导入导出、试运行、API 测试工具接口
- **管理 API**：完整的认证、CRUD、状态切换和测试接口

## 技术栈

- Go 1.21+
- 标准库 HTTP Router
- MySQL
- JWT 认证

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8+

### 安装依赖

```bash
go mod tidy
```

### 配置

在 `configs/` 目录下创建配置文件，或基于 `configs/config.example.yaml` 复制一份本地配置。

### 启动服务

```bash
# 开发模式
go run ./cmd/server

# 或编译后运行
go build -o bin/localaihub ./cmd/server
./bin/localaihub
```

服务默认监听 `0.0.0.0:8080`。

## API 前缀

- 管理 API：`/admin/api/v1`
- OpenAI 网关：`/proxy/openai/v1`
- Anthropic 网关：`/proxy/anthropic/v1`
- Gemini 网关：`/proxy/gemini/v1beta`

## 项目结构

```text
LocalAIHub_GO/
├── cmd/                  # 程序入口
├── configs/              # 配置文件
├── internal/
│   ├── app/              # 启动与路由装配
│   ├── module/           # 业务模块
│   │   ├── auth/         # 管理员认证
│   │   ├── gateway/      # AI 网关
│   │   ├── provider/     # 上游管理
│   │   ├── clientkey/    # API Key 管理
│   │   ├── route/        # 路由管理
│   │   ├── log/          # 请求日志/审计日志查询
│   │   ├── audit/        # 审计日志写入
│   │   └── configexport/ # 配置导入导出
│   └── pkg/              # 公共工具
├── migrations/           # 数据库迁移
└── sql/                  # SQL 脚本
```

## 主要模块

### Gateway

处理外部 AI 请求，实现：

- 模型路由选择
- Provider Key 选择与失败上报
- OpenAI / Anthropic / Gemini 转发
- 故障转移与熔断
- 请求日志记录与 Token 用量统计
- API Key 配额累计与禁用

### Admin API

提供后台管理能力，实现：

- 上游管理、虚拟模型、路由管理
- API Key 创建、状态切换、配额管理
- 请求日志、审计日志列表与详情
- 审计日志 CSV 导出
- 配置导入导出与试运行
- API 测试工具

## 调用示例

### OpenAI

```bash
curl -X POST http://localhost:8080/proxy/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

### Gemini

```bash
curl -X POST "http://localhost:8080/proxy/gemini/v1beta/models/gemini-2.0-flash:generateContent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "contents": [
      {"parts": [{"text": "你好"}]}
    ]
  }'
```

## 许可证

MIT
