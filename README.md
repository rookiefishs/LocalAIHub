# LocalAIHub

LocalAIHub 是一个自托管的本地 AI 网关与管理后台组合项目。
它的目标很明确：把多家模型上游统一接进一个网关里，再配一套后台把模型、密钥、路由、日志和配额集中管起来。

## 项目定位

这个仓库是 LocalAIHub 的总入口，负责组织两个核心子项目：

- `LocalAIHub_GO`：Go 后端服务，提供 AI 网关与管理 API
- `LocalAIHub_Admin`：Next.js 管理后台，负责配置、监控和运营操作

适合的使用场景：

- 统一接入 OpenAI / Anthropic / Gemini 等协议上游
- 给不同应用或客户分发独立 API Key
- 对模型、路由、配额和上游 Key 做集中管理
- 记录请求日志、审计日志、Token 消耗和模型使用情况
- 在单独部署模型能力之外，加一层可运营、可观察、可控的网关层

## 技术栈

### 后端

- Go 1.23
- MySQL
- JWT
- YAML 配置
- 标准库 HTTP 服务

### 前端

- Next.js 16
- React 19
- TypeScript
- Tailwind CSS
- Radix UI / 自定义 UI 组件
- Recharts

## 仓库结构

```text
LocalAIHub/
├── LocalAIHub_GO/       # Go 后端，网关与管理 API
├── LocalAIHub_Admin/    # Next.js 后台管理系统
├── docs/                # 项目文档
├── CHANGELOG.md         # 更新日志
└── README.md            # 项目总览
```

## 核心能力

### AI 网关

- OpenAI 兼容协议转发
- Anthropic 消息协议转发
- Gemini `generateContent` / `streamGenerateContent` 转发
- 模型路由与绑定管理
- 上游失败切换与健康检查
- 请求日志与 Token 用量统计

### 管理后台

- 仪表盘统计
- 上游 Provider 与 Provider Key 管理
- 虚拟模型管理
- 路由管理与手动切换
- API Key 创建、停用、配额配置
- 请求日志 / 审计日志查看与导出
- 配置导入导出与试运行
- API 测试工具与初始化引导

## 快速开始

### 1. 启动后端

```bash
cd LocalAIHub_GO
go mod tidy
go run ./cmd/server
```

默认监听：`0.0.0.0:8080`

### 2. 启动前端

```bash
cd LocalAIHub_Admin
npm install
npm run dev
```

默认地址：`http://localhost:3000`

### 3. 首次配置流程

1. 登录后台
2. 配置上游 Provider 和 Provider Key
3. 创建虚拟模型并绑定上游
4. 配置路由策略
5. 创建客户端 API Key
6. 在测试页验证调用链路
7. 让业务侧接入网关地址

## 接口入口

后端默认提供这些前缀：

- 管理 API：`/admin/api/v1`
- OpenAI 网关：`/proxy/openai/v1`
- Anthropic 网关：`/proxy/anthropic/v1`
- Gemini 网关：`/proxy/gemini/v1beta`

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
      {"parts": [{"text": "你好"}]}
    ]
  }'
```

## 相关文档

- [后端 README](./LocalAIHub_GO/README.md)
- [管理后台 README](./LocalAIHub_Admin/README.md)
- [文档目录](./docs/README.md)
- [更新日志](./CHANGELOG.md)

## 开发说明

- 根目录 README 只负责项目总览，子项目的启动细节、配置说明和模块说明分别看各自 README。
- 这个项目的核心不是单纯“转发请求”，而是把网关、路由、配额、日志和后台运营能力整合在一起。
- 如果你要改协议适配、路由或日志链路，优先看 `LocalAIHub_GO`；如果你要改后台交互和页面逻辑，优先看 `LocalAIHub_Admin`。

## License

MIT
