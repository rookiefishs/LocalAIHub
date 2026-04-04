export interface User {
  id: number
  username: string
  status?: string
}

export interface LoginResponse {
  user: User
  token: string
  refresh_token: string
}

export interface HourlyStat {
  hour: string
  count: number
  success: number
  avg_latency: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

export interface ModelStat {
  model_code: string
  count: number
}

export interface KeyStat {
  key_name: string
  request_count: number
  total_tokens: number
  success_rate: number
}

export interface KeyTrend {
  hour: string
  key_name: string
  count: number
  tokens: number
}

export interface KeyModelStat {
  key_name: string
  model_code: string
  count: number
}

export interface DashboardData {
  request_count: number
  success_count: number
  failure_count: number
  success_rate: number
  avg_latency_ms: number
  active_upstream_count: number
  open_circuit_count: number
  debug_session_count: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  request_trend: HourlyStat[]
  model_distribution: ModelStat[]
  key_stats: KeyStat[]
  key_trend: KeyTrend[]
  key_model_distribution: KeyModelStat[]
}

export interface Provider {
  id: number
  name: string
  provider_type: string
  service_type?: string
  base_url: string
  auth_type: string
  new_key?: string
  enabled: boolean
  status: string
  health_status?: string
  timeout_ms?: number
  remark?: string
  created_at: string
}

export interface ProviderKey {
  id: number
  provider_id: number
  key_masked: string
  status: string
  priority: number
  remark?: string
  created_at: string
  last_test_result?: string
}

export interface Model {
  id: number
  model_code: string
  display_name: string
  protocol_family?: string
  capability_flags?: string[]
  visible: boolean
  status: string
  sort_order?: number
  description?: string
  default_params_json?: Record<string, any>
  created_at: string
}

export interface ModelBinding {
  id: number
  virtual_model_id: number
  provider_id: number
  provider_name?: string
  provider_key_id?: number
  upstream_model_name: string
  priority: number
  enabled: boolean
  is_same_name?: boolean
}

export interface ClientKeyTestAttempt {
  model: string
  success: boolean
  error?: string
}

export interface ClientKeyTestResult {
  ok?: boolean
  model?: string
  url?: string
  attempts?: ClientKeyTestAttempt[]
}

export interface ClientKey {
  id: number
  name: string
  key_prefix: string
  plain_key?: string
  status: string
  remark?: string
  expires_at?: string
  allowed_models?: number[]
  daily_request_limit?: number
  monthly_request_limit?: number
  daily_token_limit?: number
  monthly_token_limit?: number
  current_daily_requests?: number
  current_monthly_requests?: number
  current_daily_tokens?: number
  current_monthly_tokens?: number
  created_at: string
  last_used_at?: string
}

export interface RouteState {
  virtual_model_id: number
  model_code: string
  display_name: string
  current_binding_id?: number
  current_binding_name?: string
  route_status: string
  manual_lock: boolean
  lock_until?: string
  last_switch_reason?: string
  last_switch_at?: string
}

export interface RequestLog {
  id: number
  trace_id: string
  protocol_type: string
  client_id?: number
  key_name?: string
  virtual_model_id?: number
  virtual_model_code?: string
  virtual_model_name?: string
  requested_model?: string
  binding_id?: number
  route_name?: string
  provider_id?: number
  provider_name?: string
  upstream_model_name?: string
  status_code?: number
  success: boolean
  latency_ms?: number
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens?: number
  error_code?: string
  error_message?: string
  request_summary?: string
  response_summary?: string
  created_at: string
}

export interface AuditLog {
  id: number
  admin_user_id: number
  action: string
  target_type: string
  target_id?: number
  change_summary?: string
  ip_address?: string
  user_agent?: string
  request_id?: string
  created_at: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}
