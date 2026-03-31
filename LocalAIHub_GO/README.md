# LocalAIHub_GO

LocalAIHub Go 后端服务，提供 AI 网关和后台管理 API。

## 功能特性

- **AI 网关**：OpenAI 兼容协议入口，支持流式和非流式响应
- **路由管理**：模型路由、故障转移、负载均衡
- **上游适配**：支持多种 AI 服务商配置
- **日志审计**：请求日志、审计日志、Token 统计
- **管理 API**：完整的 CRUD 操作接口

## 技术栈

- Go 1.21+
- Gin Web Framework
- GORM (SQLite)
- JWT 认证

## 快速开始

### 环境要求

- Go 1.21+
- SQLite

### 安装依赖

```bash
go mod tidy
```

### 配置

在 `configs/` 目录下创建配置文件，或使用默认配置。

### 启动服务

```bash
# 开发模式
go run ./cmd/server

# 或编译后运行
go build -o bin/localaihub ./cmd/server
./bin/localaihub
```

服务默认监听 `0.0.0.0:8080`。

### API 文档

管理 API 默认前缀：`/api/v1`
网关代理默认前缀：`/proxy/openai/v1`

## 项目结构

```
LocalAIHub_GO/
├── cmd/
│   └── server/          # 程序入口
├── configs/             # 配置文件
├── internal/
│   ├── app/            # 应用启动
│   ├── module/          # 业务模块
│   │   ├── admin/       # 管理员认证
│   │   ├── gateway/     # AI 网关
│   │   ├── log/         # 日志模块
│   │   └── ...
│   └── pkg/             # 公共工具
├── migrations/          # 数据库迁移
└── sql/                 # SQL 脚本
```

## 主要模块

### Gateway (AI 网关)

处理外部 AI 请求，实现：
- 协议转换
- 路由选择
- 故障转移
- 日志记录
- Token 统计

### Admin (管理后台)

提供管理 API，实现：
- 上游管理
- 模型管理
- 路由管理
- 密钥管理
- 日志查询

## 调用示例

```bash
# 使用 API Key 调用
curl -X POST http://localhost:8080/proxy/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

## 许可证

MIT
