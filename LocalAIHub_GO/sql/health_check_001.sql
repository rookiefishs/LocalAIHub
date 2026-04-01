-- LocalAIHub 健康检查功能
-- 执行时间: 在 init.sql 之后执行

-- 健康检查日志表
CREATE TABLE IF NOT EXISTS health_check_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    target_type VARCHAR(32) NOT NULL COMMENT 'provider/provider_key/client_key',
    target_id BIGINT NOT NULL,
    target_name VARCHAR(128) NULL COMMENT '名称或标识',
    check_status VARCHAR(16) NOT NULL COMMENT 'enabled/disabled/active',
    previous_status VARCHAR(16) NULL,
    error_message VARCHAR(255) NULL,
    latency_ms INT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_health_check_log_target (target_type, target_id, created_at),
    KEY idx_health_check_log_check_time (created_at),
    KEY idx_health_check_log_target_type (target_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;