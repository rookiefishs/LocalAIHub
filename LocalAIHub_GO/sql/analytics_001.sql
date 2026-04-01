-- 高级统计分析功能 - 数据库变更

-- 模型单价配置表
CREATE TABLE IF NOT EXISTS model_pricing (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_code VARCHAR(128) NOT NULL,
    provider_id BIGINT NULL,
    prompt_price_per_1k DECIMAL(10,6) NOT NULL DEFAULT 0,
    completion_price_per_1k DECIMAL(10,6) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    remark VARCHAR(255) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_model_pricing_model (model_code, provider_id),
    KEY idx_model_pricing_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 示例数据
INSERT INTO model_pricing (model_code, provider_id, prompt_price_per_1k, completion_price_per_1k, currency, remark) VALUES
('gpt-4o', NULL, 0.015, 0.03, 'USD', 'OpenAI GPT-4o'),
('gpt-4o-mini', NULL, 0.0006, 0.0012, 'USD', 'OpenAI GPT-4o-mini'),
('claude-3-opus', NULL, 0.015, 0.075, 'USD', 'Claude 3 Opus'),
('claude-3-sonnet', NULL, 0.003, 0.015, 'USD', 'Claude 3 Sonnet'),
('gemini-pro', NULL, 0.00125, 0.005, 'USD', 'Gemini Pro');
