# LocalAIHub_Admin

LocalAIHub 管理后台，基于 Next.js App Router 构建。

## 功能特性

- **仪表盘**：请求量统计、Token 消耗、模型分布
- **上游管理**：添加/编辑/删除 AI 服务商配置
- **虚拟模型**：创建模型、绑定上游、设置参数
- **路由管理**：模型路由、锁定/解锁、故障转移
- **API Key**：创建客户端密钥、设置权限
- **日志中心**：请求日志、审计日志查询
- **使用教程**：新手入门指南

## 技术栈

- Next.js 16 (App Router)
- React 19
- TypeScript
- Tailwind CSS
- shadcn/ui
- Recharts

## 快速开始

### 环境要求

- Node.js 18+
- pnpm / npm / yarn

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

服务默认运行在 `http://localhost:3000`。

### 构建生产版本

```bash
npm run build
npm start
```

## 项目结构

```
LocalAIHub_Admin/
├── app/
│   ├── dashboard/       # 管理页面
│   │   ├── keys/        # API Key 管理
│   │   ├── logs/        # 日志中心
│   │   ├── models/      # 虚拟模型
│   │   ├── routes/      # 路由管理
│   │   ├── upstreams/   # 上游管理
│   │   ├── help/        # 使用教程
│   │   └── page.tsx     # 仪表盘
│   └── login/           # 登录页
├── components/
│   ├── ui/              # UI 组件
│   └── pagination-bar.tsx
├── lib/
│   └── api.ts           # API 调用
└── ...
```

## 页面说明

| 页面 | 路径 | 说明 |
|------|------|------|
| 仪表盘 | `/dashboard` | 统计概览 |
| 上游管理 | `/dashboard/upstreams` | AI 服务商配置 |
| 虚拟模型 | `/dashboard/models` | 模型定义 |
| 路由管理 | `/dashboard/routes` | 路由策略 |
| API Key | `/dashboard/keys` | 客户端密钥 |
| 日志中心 | `/dashboard/logs` | 请求/审计日志 |
| 使用教程 | `/dashboard/help` | 新手指南 |

## 环境变量

```bash
# .env.local
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

## 调用示例

在「API Key」页面点击「使用」按钮获取调用信息：

```bash
curl -X POST YOUR_BASE_URL/proxy/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

## 许可证

MIT
