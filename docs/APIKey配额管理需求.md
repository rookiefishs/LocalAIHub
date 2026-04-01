# API Key 配额管理需求文档

## 概述

本文档描述 LocalAIHub API Key 配额管理功能的需求，实现对 API Key 的请求次数和 Token 消耗配额控制，超出配额后直接禁用 API Key。

## 功能列表

### 1. 配额类型

| 类型 | 描述 | 计量周期 |
|------|------|----------|
| daily_request_limit | 每日请求次数限制 | 每日 00:00 重置 |
| monthly_request_limit | 每月请求次数限制 | 每月 1日 重置 |
| daily_token_limit | 每日 Token 限制 | 每日 00:00 重置 |
| monthly_token_limit | 每月 Token 限制 | 每月 1日 重置 |

### 2. 配额检查逻辑

- 请求到达网关时，先检查配额
- 如果超出任一配额限制，直接返回 429 状态码
- 同时将 API Key 状态改为 disabled
- 记录配额超限日志

### 3. 数据库变更

```sql
ALTER TABLE api_client ADD COLUMN daily_request_limit INT NULL COMMENT '每日请求次数限制';
ALTER TABLE api_client ADD COLUMN monthly_request_limit INT NULL COMMENT '每月请求次数限制';
ALTER TABLE api_client ADD COLUMN daily_token_limit INT NULL COMMENT '每日Token限制';
ALTER TABLE api_client ADD COLUMN monthly_token_limit INT NULL COMMENT '每月Token限制';
ALTER TABLE api_client ADD COLUMN current_daily_requests INT NOT NULL DEFAULT 0;
ALTER TABLE api_client ADD COLUMN current_monthly_requests INT NOT NULL DEFAULT 0;
ALTER TABLE api_client ADD COLUMN current_daily_tokens INT NOT NULL DEFAULT 0;
ALTER TABLE api_client ADD COLUMN current_monthly_tokens INT NOT NULL DEFAULT 0;
ALTER TABLE api_client ADD COLUMN quota_reset_at DATE NULL COMMENT '配额重置日期';
ALTER TABLE api_client ADD COLUMN quota_disabled_at DATETIME(3) NULL COMMENT '配额超限禁用时间';

-- 配额使用记录表
CREATE TABLE IF NOT EXISTS api_client_quota_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    client_id BIGINT NOT NULL,
    quota_type VARCHAR(16) NOT NULL COMMENT 'daily/monthly',
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    request_count INT NOT NULL DEFAULT 0,
    token_count INT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_quota_log_client_period (client_id, period_start, period_end)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 后端 API 设计

### 配额管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /admin/api/v1/client-keys/{id}/quota | 获取配额配置 |
| PUT | /admin/api/v1/client-keys/{id}/quota | 设置配额 |
| GET | /admin/api/v1/client-keys/quota-usage | 批量配额使用情况 |
| POST | /admin/api/v1/client-keys/{id}/quota/reset | 重置配额使用量 |
| POST | /admin/api/v1/client-keys/quota/reset-all | 重置所有配额（定时任务） |

### API 详情

#### GET /admin/api/v1/client-keys/{id}/quota

**响应示例**

```json
{
    "code": 0,
    "data": {
        "daily_request_limit": 1000,
        "monthly_request_limit": 30000,
        "daily_token_limit": 1000000,
        "monthly_token_limit": 30000000,
        "current_daily_requests": 456,
        "current_monthly_requests": 12345,
        "current_daily_tokens": 234567,
        "current_monthly_tokens": 1234567,
        "quota_reset_at": "2024-01-01",
        "status": "active"
    }
}
```

#### PUT /admin/api/v1/client-keys/{id}/quota

**请求体**

```json
{
    "daily_request_limit": 1000,
    "monthly_request_limit": 30000,
    "daily_token_limit": 1000000,
    "monthly_token_limit": 30000000
}
```

## 前端页面设计

### 1. API Key 管理页面

在现有 API Key 管理页面添加：

- 配额使用进度条（每日请求、每日 Token）
- 超配额警告提示

### 2. 创建/编辑 Key 弹窗

添加配额设置区域：

```
┌─────────────────────────────────────────────┐
│ 配额设置（可选）                              │
├─────────────────────────────────────────────┤
│ 每日请求限制: [________] 次                   │
│ 每月请求限制: [________] 次                   │
│ 每日Token限制: [________]                    │
│ 每月Token限制: [________]                    │
└─────────────────────────────────────────────┘
```

### 3. 配额使用视图

新增独立页面显示所有 Key 的配额使用情况：

- 表格：Key名称、每日请求/限制、每日Token/限制、状态
- 可按状态筛选（正常/接近限额/已禁用）
- 一键重置按钮

## 待完成

- [ ] 数据库表变更
- [ ] 后端配额检查中间件
- [ ] 定时任务重置配额
- [ ] 前端配额配置
- [ ] 单元测试
