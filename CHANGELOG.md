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

## 2026/04/04

- Security: 管理后台登录新增 Refresh Token 刷新机制，前端在访问令牌过期后自动续签；后端配置移除明文管理员密码项，要求通过 `security.admin_password_hash` 提供哈希密码。
- Added: 网关代理新增失败自动切换到后备绑定能力，支持在当前绑定异常时自动切换到可用后备绑定并同步更新当前路由。

## 2026/04/03

- Added: 网关新增统一错误处理 handleProxyError 方法，根据错误类型返回不同状态码和错误信息。
- Added: 请求日志新增 key_name、prompt_tokens、completion_tokens、total_tokens、request_summary 字段。
- Fixed: 日志列表查询移除 TraceID 和 ProviderID 筛选条件，新增 TimeRange 快捷筛选（1h/6h/1d/3d/7d），优化查询性能。
- Fixed: 仪表盘筛选无使用记录 Key 时显示空图表而非"暂无数据"文案，提升用户体验。
- Changed: 日志页面筛选器改为下拉选择（API Key、模型），时间筛选改为快捷时间范围下拉。
- Changed: 日志表格新增 Key、请求内容、Token 列，移除了 TraceID 列。
- Changed: 按钮尺寸默认 h-11 px-6 text-base，secondary 按钮增加边框和阴影悬停效果。
- Changed: 开发环境 API 地址改为 https://www.rookiefish.com/localaihub-api。
- Changed: 全局字号调整为 16px。

## 2026/04/02

- Added: 网关补齐 Gemini 真实转发与 Anthropic 流式透传能力，支持 `generateContent` / `streamGenerateContent`、模型列表、Provider 探测与健康检查，并统一补强 TLS 校验、认证失败审计、请求日志和 Token 统计。
- Added: 审计日志补齐详情接口、CSV 导出、关键词筛选与更可读字段展示，管理后台日志页同步支持详情弹窗、导出按钮和筛选状态修复。
- Fixed: 仪表盘请求趋势与 Token 趋势改为按完整时间桶展示，补齐空小时数据并修正成功/失败统计来源，解决图表线条与数字不对应的问题。
- Docs: 根目录、后端、前端与 `docs/README.md` 文档按当前代码状态更新，补充 Gemini、审计日志、配置导入导出与日志中心等最新能力说明。

- Changed: 拖拽排序改为手动保存按钮，移除自动提交逻辑。虚拟模型上游绑定弹框、上游管理 Provider Keys 弹框、路由管理绑定列表弹框的排序均需点击"保存顺序"按钮后才落库。
- Changed: 虚拟模型上游绑定弹框移除优先级数字显示，改为与上游管理 Key 管理一致的拖拽样式。
- Changed: 侧边栏菜单选中边框改为基于主题的黑/白色（使用 var(--foreground)），选中背景改为 color-mix 混合色。

- Added: 按钮添加鼠标悬停、点击动画和效果（hover:shadow + translate-y + active:scale）。
- Changed: 页面切换动画加大，page-transition 组件改为 y:20 + scale:0.98 入场，duration 0.3s。
- Changed: 仪表盘图表切换时不再清除当前数据，直接过渡到新数据，避免"闪烁"感。切换 Key 或时间范围时保持当前图表显示。
- Changed: 仪表盘空数据时显示"暂无数据"或 skeleton，而非清空。StatCard 数字变化时有 fade 动画。
- Changed: Select 下拉菜单打开和关闭添加 framer-motion 动画，宽度与 trigger 对齐（使用 Radix 变量）。
- Changed: 表格加载状态优化，上游管理和 API Key 列表在请求中显示 spinner，数据返回后关闭 loading。

- Fixed: 暗黑模式下侧边栏菜单悬停阴影不明显问题，在 globals.css 中添加 html.dark 下的白色投影样式。
- Changed: 暗黑模式悬停阴影改为白色（rgba(255,255,255,0.28)），明亮模式使用黑色阴影。侧边栏菜单悬停时同时有背景亮起和 scale 效果。

- Changed: API 测试工具默认 content 从"你好，请介绍一下自己"改为"1"。
- Changed: 统一优化代码块样式（API 测试响应结果、使用密钥弹框、快捷流程调用示例），添加 border + shadow-inner + 主题适配背景。
- Fixed: 开发环境 API 路径修正为 http://127.0.0.1:3334，生产环境保持 /localaihub-api。
- Fixed: 开发环境 basePath 为空，生产环境为 /localaihub-admin，修复开发环境 404 问题。
- Fixed: Logo 路径改为环境感知式，开发环境加载 /logo.png，生产环境加载 /localaihub-admin/logo.png。

- Added: 新增 Provider 认证探测模块 provider_auth_probe.go，提取连接测试逻辑。
- Removed: 删除 docs 目录下的旧 API 文档文件。

- Fixed: Modal 组件新增 maxWidthClass 属性，支持自定义弹框宽度。

- Fixed: API Key 编辑保存时报错 405，修复 PUT 路由缺失问题，同时修复过期时间更新逻辑。
- Fixed: 修复 API Key 过期时间显示错误，过期或超出使用时间的 Key 过期时间和使用时间列显示红色。
- Changed: 使用密钥弹框中的 Base URL 改为显示到 /proxy/openai/v1，方便复制。
- Changed: 仪表盘支持按 Key 分别显示使用情况，全部 Key 时请求趋势、Token 趋势、模型分布图表分别按 Key 显示不同颜色堆叠显示，选中单个 Key 时显示该 Key 详细请求状态。
- Added: 仪表盘新增后端接口返回各 Key 使用统计数据（key_stats、key_trend、key_model_distribution）。
- Fixed: 仪表盘 API Key 配置页面 baseURL 根据环境区分，开发环境显示 http://127.0.0.1:3334，生产环境显示域名形式。

- Fixed: 侧边栏高亮逻辑优化，使用 mounted 状态确保 SSR 渲染正确，去除 URL 尾部斜杠后匹配，解决仪表盘及子路由高亮异常问题。
- Fixed: 导出配置功能修复，移除不存在的 allowed_models_json 字段查询。
- Changed: 配置导入导出页面复选框改为 Switch 开关组件，文件上传区域改为拖拽式样式，与整体项目主题保持一致。
- Changed: 应用全局尺寸放大，基础字号从 13px 调整为 14px。
- Fixed: 删除数据前增加绑定检查，上游、虚拟模型、路由删除时如果存在关联绑定则阻止删除并返回友好错误提示。
- Changed: 趋势统计时区改为北京时间（UTC+8），请求趋势、Token 统计等图表数据时区自动转换。

- Fixed: 配置导入导出页面表单布局调整为与其他模块一致的样式。
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

- Added: 网关新增统一错误处理 handleProxyError 方法，根据错误类型返回不同状态码和错误信息。
- Added: 请求日志新增 key_name、prompt_tokens、completion_tokens、total_tokens、request_summary 字段。
- Fixed: 日志列表查询移除 TraceID 和 ProviderID 筛选条件，新增 TimeRange 快捷筛选（1h/6h/1d/3d/7d），优化查询性能。
- Fixed: 仪表盘筛选无使用记录 Key 时显示空图表而非"暂无数据"文案，提升用户体验。
- Changed: 日志页面筛选器改为下拉选择（API Key、模型），时间筛选改为快捷时间范围下拉。
- Changed: 日志表格新增 Key、请求内容、Token 列，移除了 TraceID 列。
- Changed: 按钮尺寸默认 h-11 px-6 text-base，secondary 按钮增加边框和阴影悬停效果。
- Changed: 开发环境 API 地址改为 https://www.rookiefish.com/localaihub-api。
- Changed: 全局字号调整为 16px。

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
