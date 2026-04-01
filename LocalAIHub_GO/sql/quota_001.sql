-- API Key 配额管理功能数据库变更

-- 为 api_client 表添加配额相关字段
ALTER TABLE api_client ADD COLUMN daily_request_limit INT NULL COMMENT '每日请求次数限制';
ALTER TABLE api_client ADD COLUMN monthly_request_limit INT NULL COMMENT '每月请求次数限制';
ALTER TABLE api_client ADD COLUMN daily_token_limit BIGINT NULL COMMENT '每日Token限制';
ALTER TABLE api_client ADD COLUMN monthly_token_limit BIGINT NULL COMMENT '每月Token限制';
ALTER TABLE api_client ADD COLUMN current_daily_requests INT NOT NULL DEFAULT 0 COMMENT '当日已使用请求数';
ALTER TABLE api_client ADD COLUMN current_monthly_requests INT NOT NULL DEFAULT 0 COMMENT '当月已使用请求数';
ALTER TABLE api_client ADD COLUMN current_daily_tokens BIGINT NOT NULL DEFAULT 0 COMMENT '当日已使用Token数';
ALTER TABLE api_client ADD COLUMN current_monthly_tokens BIGINT NOT NULL DEFAULT 0 COMMENT '当月已使用Token数';
ALTER TABLE api_client ADD COLUMN quota_reset_at DATE NULL COMMENT '配额重置日期';
ALTER TABLE api_client ADD COLUMN quota_disabled_at DATETIME(3) NULL COMMENT '配额超限禁用时间';

-- 创建配额使用记录表
CREATE TABLE IF NOT EXISTS api_client_quota_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    client_id BIGINT NOT NULL,
    quota_type VARCHAR(16) NOT NULL COMMENT 'daily/monthly',
    period_date DATE NOT NULL,
    request_count INT NOT NULL DEFAULT 0,
    token_count BIGINT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_quota_log_client_date (client_id, period_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
