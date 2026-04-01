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

- Added: 配置导入导出功能，支持一键导出/导入全部配置（上游、虚拟模型、绑定、API Key），支持覆盖模式和试运行预览。
- Fixed: API 测试工具下拉框过滤掉已禁用的 API Key，避免选择后认证失败。
- Added: API 测试工具页面，支持选择 API Key 和模型发送测试请求，实时显示响应、耗时和 Token 消耗统计。
- Added: API Key 配额管理功能，支持设置每日/每月请求次数和 Token 限制，配额耗尽后自动禁用 Key。
- Added: Gateway 请求处理中增加配额检查和 Token 使用统计。
- Added: 前端 API Key 列表页面新增「配额」按钮，可查看和编辑配额设置。
- Changed: Select 下拉菜单悬停和选中项改为黑底白字样式（自动适配亮色/暗色主题）。
- Fixed: 修复 API Key 页面多个 key 启用/禁用按钮共享 loading 状态的问题，现在每个 key 独立控制 loading。
- Changed: 侧边栏选中项改为黑底白字样式（自动适配亮色/暗色主题）。
- Changed: 快捷流程步进指示器当前步骤改为黑底白字样式（自动适配亮色/暗色主题）。
- Changed: 仪表盘 StatCard 悬停效果改为向上移动而非放大，移除阴影效果。
- Changed: 仪表盘上游统计卡片将启用状态显示在数字后面（如 "3 启用"）。
- Changed: 路由管理页面添加按钮改为仅显示图标，并移至卡片 header 右侧，与虚拟模型列表样式保持一致。
- Changed: 移除「使用教程」页面的标题和介绍区域。
- Removed: 移除「快捷流程」页面的标题和介绍区域。
- Added: 仪表盘新增 API Key 筛选功能，请求趋势和 Token 趋势图表支持按不同 API Key 筛选，默认显示全部 Key 的汇总数据，图表标题显示当前选中的 Key 名称。
- Changed: Select 下拉菜单增加悬停样式（data-[highlighted]）和选中样式（data-[state=checked]），提升交互体验。
- Changed: 仪表盘 StatCard 增加点击跳转功能和悬停动画效果，5 个统计卡片分别跳转到日志或上游管理页面。
- Changed: 快捷流程页面的步进指示器模块放入卡片中并居中显示，容器最大宽度调整为 max-w-5xl。
- Removed: 移除统计分析模块，删除前端 analytics 页面、侧边栏入口、后端 analytics handler 及相关 API（与仪表盘功能重复）。
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
