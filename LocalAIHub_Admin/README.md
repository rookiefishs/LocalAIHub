# LocalAIHub_Admin

<p align="center">
  <img src="../logo.png" alt="LocalAIHub Logo" width="64" height="64" />
</p>

<p align="center">
  LocalAIHub 管理后台，基于 Next.js App Router 构建。
</p>

## 功能特性

- **仪表盘**：请求量、Token、模型分布、API Key 维度统计
- **上游管理**：Provider、Provider Key、连接测试、优先级管理
- **虚拟模型**：多协议模型定义、绑定、测试与排序
- **路由管理**：手动切换、解锁、故障转移状态查看
- **API Key**：创建、启停、使用示例、每日/每月配额设置
- **日志中心**：请求日志、审计日志、详情查看、CSV 导出
- **API 测试**：直接选 Key 和模型发起测试请求
- **配置导入导出**：按模块导出、导入、试运行
- **快捷流程 / 使用教程**：帮助完成初始化配置

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

```text
LocalAIHub_Admin/
├── app/
│   ├── dashboard/
│   │   ├── keys/         # API Key 管理
│   │   ├── logs/         # 请求日志 / 审计日志
│   │   ├── models/       # 虚拟模型
│   │   ├── routes/       # 路由管理
│   │   ├── settings/     # 配置导入导出
│   │   ├── test/         # API 测试工具
│   │   ├── upstreams/    # 上游管理
│   │   ├── wizard/       # 快捷流程
│   │   ├── help/         # 使用教程
│   │   └── page.tsx      # 仪表盘
│   └── login/            # 登录页
├── components/
│   ├── ui/               # UI 组件
│   └── pagination-bar.tsx
├── lib/
│   └── api.ts            # API 调用
└── ...
```

## 页面说明

| 页面 | 路径 | 说明 |
|------|------|------|
| 仪表盘 | `/dashboard` | 统计概览 |
| 快捷流程 | `/dashboard/wizard` | 引导完成初始化配置 |
| 上游管理 | `/dashboard/upstreams` | AI 服务商与 Key 配置 |
| 虚拟模型 | `/dashboard/models` | 模型定义与上游绑定 |
| 路由管理 | `/dashboard/routes` | 路由策略与手动切换 |
| API Key | `/dashboard/keys` | 客户端密钥与配额 |
| 日志中心 | `/dashboard/logs` | 请求日志 / 审计日志 |
| API 测试 | `/dashboard/test` | 直接测试请求 |
| 配置导入导出 | `/dashboard/settings` | 导入、导出、试运行 |
| 使用教程 | `/dashboard/help` | 新手指南 |

## 环境变量

```bash
# .env.local
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

开发环境若未显式配置，会默认走 `http://127.0.0.1:3334`。

## 调用示例

在「API Key」页面点击「使用」按钮可获取调用信息：

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
