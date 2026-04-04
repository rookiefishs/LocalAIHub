CREATE DATABASE localaihub CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE localaihub;

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS admin_user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    password_algo VARCHAR(32) NOT NULL DEFAULT 'bcrypt',
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    last_login_at DATETIME(3) NULL,
    last_login_ip VARCHAR(64) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_admin_user_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_login_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    admin_user_id BIGINT NULL,
    username VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL,
    ip_address VARCHAR(64) NULL,
    user_agent VARCHAR(255) NULL,
    result_message VARCHAR(255) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_admin_login_log_admin_user_created_at (admin_user_id, created_at),
    KEY idx_admin_login_log_username_created_at (username, created_at),
    CONSTRAINT fk_admin_login_log_admin_user FOREIGN KEY (admin_user_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS api_client (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    key_prefix VARCHAR(32) NOT NULL,
    api_key_hash VARCHAR(255) NOT NULL,
    plain_key VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    remark VARCHAR(255) NULL,
    last_used_at DATETIME(3) NULL,
    expires_at DATETIME(3) NULL,
    allowed_models_json JSON NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_api_client_key_prefix (key_prefix),
    KEY idx_api_client_status (status),
    KEY idx_api_client_expires_at (expires_at),
    KEY idx_api_client_last_used_at (last_used_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS api_client_model (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    client_id BIGINT NOT NULL,
    virtual_model_id BIGINT NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_api_client_model_unique (client_id, virtual_model_id),
    KEY idx_api_client_model_client_id (client_id),
    KEY idx_api_client_model_virtual_model_id (virtual_model_id),
    CONSTRAINT fk_api_client_model_client FOREIGN KEY (client_id) REFERENCES api_client(id) ON DELETE CASCADE,
    CONSTRAINT fk_api_client_model_virtual_model FOREIGN KEY (virtual_model_id) REFERENCES virtual_model(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS provider (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    provider_type VARCHAR(32) NOT NULL,
    service_type VARCHAR(32) NOT NULL DEFAULT 'official',
    base_url VARCHAR(255) NOT NULL,
    auth_type VARCHAR(32) NOT NULL DEFAULT 'bearer',
    timeout_ms INT NOT NULL DEFAULT 60000,
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    health_status VARCHAR(16) NOT NULL DEFAULT 'healthy',
    last_health_check_at DATETIME(3) NULL,
    last_health_message VARCHAR(255) NULL,
    remark VARCHAR(255) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    KEY idx_provider_type (provider_type),
    KEY idx_provider_enabled (enabled),
    KEY idx_provider_health_status (health_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS provider_key (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    provider_id BIGINT NOT NULL,
    key_masked VARCHAR(64) NOT NULL,
    secret_encrypted TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'enabled',
    priority INT NOT NULL DEFAULT 1,
    fail_count INT NOT NULL DEFAULT 0,
    last_used_at DATETIME(3) NULL,
    last_error_at DATETIME(3) NULL,
    last_error_message VARCHAR(255) NULL,
    remark VARCHAR(255) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    KEY idx_provider_key_provider_id (provider_id),
    KEY idx_provider_key_status (status),
    KEY idx_provider_key_provider_priority (provider_id, priority),
    CONSTRAINT fk_provider_key_provider FOREIGN KEY (provider_id) REFERENCES provider(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS virtual_model (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_code VARCHAR(128) NOT NULL,
    display_name VARCHAR(128) NOT NULL,
    protocol_family VARCHAR(32) NOT NULL,
    capability_flags JSON NOT NULL,
    visible TINYINT(1) NOT NULL DEFAULT 1,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    sort_order INT NOT NULL DEFAULT 0,
    description VARCHAR(255) NULL,
    default_params_json JSON NOT NULL,
    remark VARCHAR(255) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_virtual_model_model_code (model_code),
    KEY idx_virtual_model_visible_status (visible, status),
    KEY idx_virtual_model_sort_order (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS virtual_model_binding (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    virtual_model_id BIGINT NOT NULL,
    provider_id BIGINT NOT NULL,
    provider_key_id BIGINT NULL,
    upstream_model_name VARCHAR(128) NOT NULL,
    priority INT NOT NULL DEFAULT 1,
    is_same_name TINYINT(1) NOT NULL DEFAULT 0,
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    capability_snapshot_json JSON NOT NULL,
    param_override_json JSON NOT NULL,
    remark VARCHAR(255) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_virtual_model_binding_unique (virtual_model_id, provider_id, upstream_model_name),
    KEY idx_virtual_model_binding_model_enabled_priority (virtual_model_id, enabled, priority),
    KEY idx_virtual_model_binding_provider_id (provider_id),
    CONSTRAINT fk_virtual_model_binding_model FOREIGN KEY (virtual_model_id) REFERENCES virtual_model(id) ON DELETE CASCADE,
    CONSTRAINT fk_virtual_model_binding_provider FOREIGN KEY (provider_id) REFERENCES provider(id) ON DELETE CASCADE,
    CONSTRAINT fk_virtual_model_binding_provider_key FOREIGN KEY (provider_key_id) REFERENCES provider_key(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS route_state (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    virtual_model_id BIGINT NOT NULL,
    current_binding_id BIGINT NULL,
    route_status VARCHAR(16) NOT NULL DEFAULT 'normal',
    manual_locked TINYINT(1) NOT NULL DEFAULT 0,
    lock_until DATETIME(3) NULL,
    last_switch_reason VARCHAR(255) NULL,
    last_switch_at DATETIME(3) NULL,
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_route_state_virtual_model_id (virtual_model_id),
    CONSTRAINT fk_route_state_model FOREIGN KEY (virtual_model_id) REFERENCES virtual_model(id) ON DELETE CASCADE,
    CONSTRAINT fk_route_state_binding FOREIGN KEY (current_binding_id) REFERENCES virtual_model_binding(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS route_switch_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    virtual_model_id BIGINT NOT NULL,
    from_binding_id BIGINT NULL,
    to_binding_id BIGINT NULL,
    trigger_type VARCHAR(32) NOT NULL,
    operator_admin_id BIGINT NULL,
    reason VARCHAR(255) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_route_switch_log_virtual_model_created_at (virtual_model_id, created_at),
    KEY idx_route_switch_log_trigger_type (trigger_type),
    CONSTRAINT fk_route_switch_log_model FOREIGN KEY (virtual_model_id) REFERENCES virtual_model(id) ON DELETE CASCADE,
    CONSTRAINT fk_route_switch_log_from_binding FOREIGN KEY (from_binding_id) REFERENCES virtual_model_binding(id) ON DELETE SET NULL,
    CONSTRAINT fk_route_switch_log_to_binding FOREIGN KEY (to_binding_id) REFERENCES virtual_model_binding(id) ON DELETE SET NULL,
    CONSTRAINT fk_route_switch_log_admin FOREIGN KEY (operator_admin_id) REFERENCES admin_user(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS circuit_breaker_state (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    provider_id BIGINT NOT NULL,
    virtual_model_id BIGINT NOT NULL,
    state VARCHAR(16) NOT NULL DEFAULT 'closed',
    failure_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failure_rate DECIMAL(5,4) NULL,
    last_failure_at DATETIME(3) NULL,
    next_retry_at DATETIME(3) NULL,
    last_reason VARCHAR(255) NULL,
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_circuit_breaker_provider_virtual_model (provider_id, virtual_model_id),
    KEY idx_circuit_breaker_state_state (state),
    KEY idx_circuit_breaker_state_next_retry_at (next_retry_at),
    CONSTRAINT fk_circuit_breaker_provider FOREIGN KEY (provider_id) REFERENCES provider(id) ON DELETE CASCADE,
    CONSTRAINT fk_circuit_breaker_model FOREIGN KEY (virtual_model_id) REFERENCES virtual_model(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS debug_session (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    scope_type VARCHAR(32) NOT NULL,
    scope_value VARCHAR(128) NULL,
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    start_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    end_at DATETIME(3) NOT NULL,
    operator_admin_id BIGINT NOT NULL,
    reason VARCHAR(255) NULL,
    closed_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    KEY idx_debug_session_enabled (enabled),
    KEY idx_debug_session_scope (scope_type, scope_value),
    KEY idx_debug_session_end_at (end_at),
    CONSTRAINT fk_debug_session_admin FOREIGN KEY (operator_admin_id) REFERENCES admin_user(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS request_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    trace_id VARCHAR(64) NOT NULL,
    protocol_type VARCHAR(32) NOT NULL,
    client_id BIGINT NULL,
    virtual_model_id BIGINT NULL,
    virtual_model_code VARCHAR(128) NULL,
    requested_model VARCHAR(128) NULL,
    binding_id BIGINT NULL,
    provider_id BIGINT NULL,
    provider_key_id BIGINT NULL,
    upstream_model_name VARCHAR(128) NULL,
    request_summary_json JSON NULL,
    response_summary_json JSON NULL,
    status_code INT NULL,
    success TINYINT(1) NOT NULL DEFAULT 0,
    latency_ms INT NULL,
    prompt_tokens INT NULL,
    completion_tokens INT NULL,
    total_tokens INT NULL,
    error_code VARCHAR(64) NULL,
    error_message VARCHAR(255) NULL,
    is_debug_logged TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_request_log_trace_id (trace_id),
    KEY idx_request_log_created_at (created_at),
    KEY idx_request_log_client_created_at (client_id, created_at),
    KEY idx_request_log_virtual_model_created_at (virtual_model_code, created_at),
    KEY idx_request_log_requested_model_created_at (requested_model, created_at),
    KEY idx_request_log_binding_created_at (binding_id, created_at),
    KEY idx_request_log_provider_created_at (provider_id, created_at),
    KEY idx_request_log_success_created_at (success, created_at),
    CONSTRAINT fk_request_log_client FOREIGN KEY (client_id) REFERENCES api_client(id) ON DELETE SET NULL,
    CONSTRAINT fk_request_log_virtual_model FOREIGN KEY (virtual_model_id) REFERENCES virtual_model(id) ON DELETE SET NULL,
    CONSTRAINT fk_request_log_binding FOREIGN KEY (binding_id) REFERENCES virtual_model_binding(id) ON DELETE SET NULL,
    CONSTRAINT fk_request_log_provider FOREIGN KEY (provider_id) REFERENCES provider(id) ON DELETE SET NULL,
    CONSTRAINT fk_request_log_provider_key FOREIGN KEY (provider_key_id) REFERENCES provider_key(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    admin_user_id BIGINT NOT NULL,
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(64) NOT NULL,
    target_id BIGINT NULL,
    change_summary_json JSON NULL,
    ip_address VARCHAR(64) NULL,
    user_agent VARCHAR(255) NULL,
    request_id VARCHAR(64) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_audit_log_admin_user_created_at (admin_user_id, created_at),
    KEY idx_audit_log_target (target_type, target_id),
    KEY idx_audit_log_action_created_at (action, created_at),
    CONSTRAINT fk_audit_log_admin FOREIGN KEY (admin_user_id) REFERENCES admin_user(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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

CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_group VARCHAR(64) NOT NULL,
    config_key VARCHAR(128) NOT NULL,
    config_value_json JSON NOT NULL,
    description VARCHAR(255) NULL,
    updated_by BIGINT NULL,
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_system_config (config_group, config_key),
    CONSTRAINT fk_system_config_admin FOREIGN KEY (updated_by) REFERENCES admin_user(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO admin_user (username, password_hash, password_algo, status)
VALUES ('admin', '$2a$10$7sM8J6wWmE0ax4qP1fM4HOdM3V7T3rP3r0nJQJ0WQ0l8KjYV0t5Nu', 'bcrypt', 'active');

INSERT IGNORE INTO system_config (config_group, config_key, config_value_json, description)
VALUES
    ('gateway', 'default_timeout_ms', JSON_OBJECT('value', 60000), 'Default upstream timeout in milliseconds'),
    ('gateway', 'request_body_limit_mb', JSON_OBJECT('value', 5), 'Maximum request body size in MB'),
    ('circuit_breaker', 'failure_threshold', JSON_OBJECT('value', 5), 'Consecutive failure threshold'),
    ('circuit_breaker', 'failure_rate_threshold', JSON_OBJECT('value', 0.5), 'Failure rate threshold'),
    ('circuit_breaker', 'open_duration_seconds', JSON_OBJECT('value', 60), 'Circuit open duration in seconds'),
    ('circuit_breaker', 'half_open_probe_count', JSON_OBJECT('value', 2), 'Half-open probe count'),
    ('logs', 'request_log_retention_days', JSON_OBJECT('value', 7), 'Request log retention days'),
    ('logs', 'audit_log_retention_days', JSON_OBJECT('value', 30), 'Audit log retention days'),
    ('logs', 'debug_log_retention_days', JSON_OBJECT('value', 3), 'Debug log retention days'),
    ('health_check', 'enabled', JSON_OBJECT('value', true), '是否启用健康检查'),
    ('health_check', 'interval_minutes', JSON_OBJECT('value', 3), '健康检查间隔（分钟）'),
    ('health_check', 'retention_days', JSON_OBJECT('value', 30), '健康检查日志保留天数');

INSERT IGNORE INTO virtual_model (model_code, display_name, protocol_family, capability_flags, visible, status, sort_order, description, default_params_json)
VALUES
    ('gpt-4o', 'GPT-4o', 'openai', JSON_ARRAY('text', 'chat'), 1, 'active', 10, 'OpenAI family virtual text model', JSON_OBJECT()),
    ('claude-sonnet', 'Claude Sonnet', 'anthropic', JSON_ARRAY('text', 'chat'), 1, 'active', 20, 'Anthropic family virtual text model', JSON_OBJECT()),
    ('gemini-pro', 'Gemini Pro', 'gemini', JSON_ARRAY('text', 'chat'), 1, 'active', 30, 'Gemini family virtual text model', JSON_OBJECT());
