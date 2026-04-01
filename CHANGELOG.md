<!--
 * @Author: rookiefish <3209605851@qq.com>
 * @Date: 2026-03-30
 * @LastEditTime: 2026-04-01
 * @LastEditors: rookiefish <3209605851@qq.com>
 * @Description: LocalAIHub 项目更新日志
-->

# Changelog

本文档用于记录项目的重要更新。

- Added: 新增功能或组件
- Changed: 现有功能调整或优化
- Fixed: 问题修复或漏洞修补
- Deprecated: 即将废弃的功能
- Removed: 已移除的功能
- Security: 安全相关更新
- Performance: 性能优化
- Docs: 文档更新
- Refactor: 代码重构
- Reverted: 回滚操作

## 2026/04/01

- Added: 仪表盘新增 API Key 筛选功能，请求趋势和 Token 趋势图表支持按不同 API Key 筛选，默认显示全部 Key 的汇总数据，图表标题显示当前选中的 Key 名称。
- Added: 新增高级统计分析页面，包含成本统计（按供应商/模型）、Token 消耗趋势、同比环比对比、供应商费用占比饼图、模型费用排行等功能。
- Added: 新增统计分析侧边栏菜单入口，后端实现相应的 analytics API。
- Added: 生成待实现功能需求文档，包括高级统计分析、告警通知系统、API Key 配额管理、操作审计日志、配置导入导出、API 测试工具。
- Changed: 统一前端所有表单控件圆角为 10px（与卡片一致），包括 Button、Input、Textarea、Select、Dialog、AlertDialog、ConfirmDialog、Modal、Tabs 等组件。
- Changed: Select 下拉菜单增加悬停样式（data-[highlighted]）和选中样式（data-[state=checked]），提升交互体验。
- Changed: 仪表盘 StatCard 增加点击跳转功能和悬停动画效果（scale + shadow），5 个统计卡片分别跳转到日志或上游管理页面。
- Changed: 快捷流程页面的步进指示器模块放入卡片中并居中显示，容器最大宽度调整为 max-w-5xl。
- Fixed: 修复 Go 后端重复导入包的编译错误（healthcheck service）。
- Fixed: 删除 gateway_repository.go 中重复声明的方法（CountSuccessRequests、AvgLatency、SumTokens 带参数版本冲突）。

## 2026/03/31

- Docs: 新增根目录、`LocalAIHub_GO/README.md` 与 `LocalAIHub_Admin/README.md` 项目说明文档，并统一接入项目 Logo 展示。
- Security: 补充根目录及后端 `.gitignore` 规则，忽略本地配置文件、环境变量、数据库文件、日志与可执行文件，避免敏感信息误传仓库。
- Changed: 管理后台登录页、站点图标与侧边栏品牌区统一替换为 `LocalAIHub` Logo，收口项目视觉标识。
- Added: 分页器新增每页条数选择器（10/30/50/100），支持在所有管理页面自定义每页显示数量。
- Added: 新增「使用教程」页面，包含快速开始指南、调用示例和注意事项，可从侧边栏「系统管理」入口访问。
- Changed: 管理后台 API Key 弹窗改为使用 `window.location.origin` 获取当前域名，修复生产环境显示 127.0.0.1 的问题。
- Fixed: 修复日志分页查询返回错误总数的问题，现在正确返回实际记录数。
- Fixed: 修复 Token 使用统计未记录的问题，现在会从上游响应中提取 token 用量并保存。

## 2026/03/30

- Added: 初始化 LocalAIHub 项目，包含 Go 后端网关和 Next.js 管理前台。
- Added: 实现上游管理功能，支持添加/编辑/删除 AI 服务商配置。
- Added: 实现虚拟模型管理，支持创建模型、绑定上游、设置优先级。
- Added: 实现路由管理，支持模型路由、锁定/解锁、手动切换上游。
- Added: 实现 API Key 管理，支持创建、编辑、删除客户端密钥，设置有效期和模型权限。
- Added: 实现日志中心，支持请求日志和审计日志查询。
- Added: 实现仪表盘，展示请求量、Token 消耗、模型分布等统计信息。
- Added: 实现登录认证和权限验证。
- Added: 代理网关支持 OpenAI 兼容协议，支持流式和非流式响应。
- Added: 实现路由自动故障转移，当前端失败时自动切换到下一个上游。

### 项目初始化

- 初始化 LocalAIHub_GO Go 后端项目，使用 Gin 框架。
- 初始化 LocalAIHub_Admin Next.js 管理前台，使用 App Router + TypeScript + Tailwind CSS + shadcn/ui。
- 集成 SQLite 数据库存储配置和日志。
- 实现 JWT 认证和 RBAC 权限管理。
- 实现请求日志记录和审计日志。
- 实现 Token 用量统计和模型分布统计。
