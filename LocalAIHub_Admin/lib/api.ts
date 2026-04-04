import { getToken, getRefreshToken, setToken, clearToken } from '@/lib/auth'
import type {
  LoginResponse,
  DashboardData,
  Provider,
  ProviderKey,
  Model,
  ModelBinding,
  ClientKey,
  ClientKeyTestResult,
  RouteState,
  RequestLog,
  AuditLog,
  PaginatedResponse,
} from '@/lib/types'

const isDev = process.env.NODE_ENV === 'development'
const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || (isDev ? 'https://www.rookiefish.com/localaihub-api' : '/localaihub-api')

type RequestOptions = RequestInit & {
  auth?: boolean
}

async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return false

  try {
    const res = await fetch(`${API_BASE}/admin/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!res.ok) return false

    const data = await res.json()
    if (data.code === 0 && data.data) {
      setToken(data.data.token, data.data.refresh_token)
      return true
    }
    return false
  } catch {
    return false
  }
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers || {})
  headers.set('Content-Type', 'application/json')
  if (options.auth !== false) {
    const token = getToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
  }

  let response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    cache: 'no-store',
  })

  if (response.status === 401 && options.auth !== false) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      const newToken = getToken()
      headers.set('Authorization', `Bearer ${newToken}`)
      response = await fetch(`${API_BASE}${path}`, {
        ...options,
        headers,
        cache: 'no-store',
      })
    } else {
      clearToken()
      window.location.href = '/login'
      throw new Error('Session expired')
    }
  }

  const payload = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(payload?.message || payload?.error?.message || `Request failed: ${response.status}`)
  }

  if (payload && typeof payload === 'object' && 'code' in payload) {
    return payload.data as T
  }

  return payload as T
}

export async function apiDownload(path: string): Promise<void> {
  const headers = new Headers()
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const response = await fetch(`${API_BASE}${path}`, {
    method: 'GET',
    headers,
    cache: 'no-store',
  })

  if (!response.ok) {
    const payload = await response.json().catch(() => ({}))
    throw new Error(payload?.message || payload?.error?.message || `Request failed: ${response.status}`)
  }

  const blob = await response.blob()
  const disposition = response.headers.get('Content-Disposition') || ''
  const fileNameMatch = disposition.match(/filename="?([^";]+)"?/i)
  const fileName = fileNameMatch?.[1] || 'download.bin'
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  link.click()
  URL.revokeObjectURL(url)
}

export const api = {
  login: (body: { username: string; password: string }) => apiRequest<LoginResponse>('/admin/api/v1/auth/login', { method: 'POST', auth: false, body: JSON.stringify(body) }),
  refresh: (body: { refresh_token: string }) => apiRequest<LoginResponse>('/admin/api/v1/auth/refresh', { method: 'POST', auth: false, body: JSON.stringify(body) }),
  me: () => apiRequest<{ user: { id: number; username: string; status: string } }>('/admin/api/v1/auth/me'),
  dashboard: (query = '') => apiRequest<DashboardData>(`/admin/api/v1/dashboard/overview${query ? `?${query}` : ''}`),
  providers: (query = '') => apiRequest<PaginatedResponse<Provider>>(`/admin/api/v1/providers${query ? `?${query}` : ''}`),
  createProvider: (body: Partial<Provider>) => apiRequest<Provider>('/admin/api/v1/providers', { method: 'POST', body: JSON.stringify(body) }),
  updateProvider: (id: number, body: Partial<Provider>) => apiRequest<Provider>(`/admin/api/v1/providers/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteProvider: (id: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/providers/${id}`, { method: 'DELETE' }),
  updateProviderStatus: (id: number, enabled: boolean) => apiRequest<{ success: boolean }>(`/admin/api/v1/providers/${id}/status`, { method: 'POST', body: JSON.stringify({ enabled }) }),
  testProvider: (id: number) => apiRequest<{ success: boolean; message?: string; auth_type?: string; auth_auto_detected?: boolean; tested_url?: string; latency_ms?: number }>(`/admin/api/v1/providers/${id}/test-connection`, { method: 'POST', body: JSON.stringify({}) }),
  providerKeys: (id: number) => apiRequest<PaginatedResponse<ProviderKey>>(`/admin/api/v1/providers/${id}/keys`),
  createProviderKey: (id: number, body: Partial<ProviderKey>) => apiRequest<ProviderKey>(`/admin/api/v1/providers/${id}/keys`, { method: 'POST', body: JSON.stringify(body) }),
  deleteProviderKey: (providerId: number, keyId: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/providers/${providerId}/keys/${keyId}`, { method: 'DELETE' }),
  updateProviderKeyStatus: (providerId: number, keyId: number, status: string) => apiRequest<{ success: boolean }>(`/admin/api/v1/providers/${providerId}/keys/${keyId}/status`, { method: 'POST', body: JSON.stringify({ status }) }),
  testProviderKey: (providerId: number, keyId: number) => apiRequest<{ success: boolean; message?: string; error?: string; auth_type?: string; auth_auto_detected?: boolean; tested_url?: string }>(`/admin/api/v1/providers/${providerId}/keys/${keyId}/test`, { method: 'POST', body: JSON.stringify({}) }),
  updateProviderKey: (providerId: number, keyId: number, body: Partial<ProviderKey>) => apiRequest<ProviderKey>(`/admin/api/v1/providers/${providerId}/keys/${keyId}`, { method: 'PUT', body: JSON.stringify(body) }),
  updateProviderKeyPriority: (providerId: number, keyId: number, priority: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/providers/${providerId}/keys/${keyId}/priority`, { method: 'POST', body: JSON.stringify({ priority }) }),
  models: (query = '') => apiRequest<PaginatedResponse<Model>>(`/admin/api/v1/models${query ? `?${query}` : ''}`),
  createModel: (body: Partial<Model>) => apiRequest<Model>('/admin/api/v1/models', { method: 'POST', body: JSON.stringify(body) }),
  updateModel: (id: number, body: Partial<Model>) => apiRequest<Model>(`/admin/api/v1/models/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteModel: (id: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/models/${id}`, { method: 'DELETE' }),
  modelBindings: (id: number) => apiRequest<PaginatedResponse<ModelBinding>>(`/admin/api/v1/models/${id}/bindings`),
  createModelBinding: (id: number, body: Partial<ModelBinding>) => apiRequest<ModelBinding>(`/admin/api/v1/models/${id}/bindings`, { method: 'POST', body: JSON.stringify(body) }),
  updateModelBinding: (modelId: number, bindingId: number, body: Partial<ModelBinding>) => apiRequest<ModelBinding>(`/admin/api/v1/models/${modelId}/bindings/${bindingId}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteModelBinding: (modelId: number, bindingId: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/models/${modelId}/bindings/${bindingId}`, { method: 'DELETE' }),
  testModelBinding: (modelId: number, bindingId: number) => apiRequest<{ success?: boolean; message?: string; model?: string; tested_url?: string; auth_type?: string }>(`/admin/api/v1/models/${modelId}/bindings/${bindingId}/test`, { method: 'POST', body: JSON.stringify({}) }),
  routes: (query = '') => apiRequest<PaginatedResponse<RouteState>>(`/admin/api/v1/routes${query ? `?${query}` : ''}`),
  switchRoute: (id: number, body: { target_binding_id: number; manual_lock?: boolean; reason?: string; lock_until?: string | null }) => apiRequest<{ success: boolean }>(`/admin/api/v1/routes/${id}/switch`, { method: 'POST', body: JSON.stringify(body) }),
  unlockRoute: (id: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/routes/${id}/unlock`, { method: 'POST', body: JSON.stringify({}) }),
  deleteRoute: (id: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/routes/${id}`, { method: 'DELETE' }),
  clientKeys: (query = '') => apiRequest<PaginatedResponse<ClientKey>>(`/admin/api/v1/client-keys${query ? `?${query}` : ''}`),
  createClientKey: (body: Partial<ClientKey>) => apiRequest<ClientKey>('/admin/api/v1/client-keys', { method: 'POST', body: JSON.stringify(body) }),
  getClientKey: (id: number) => apiRequest<ClientKey>(`/admin/api/v1/client-keys/${id}`),
  testClientKey: (id: number) => apiRequest<ClientKeyTestResult>(`/admin/api/v1/client-keys/${id}/test`, { method: 'POST', body: JSON.stringify({}) }),
  updateClientKey: (id: number, body: Partial<ClientKey>) => apiRequest<ClientKey>(`/admin/api/v1/client-keys/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteClientKey: (id: number) => apiRequest<{ success: boolean }>(`/admin/api/v1/client-keys/${id}`, { method: 'DELETE' }),
  updateClientKeyStatus: (id: number, status: string) => apiRequest<{ success: boolean }>(`/admin/api/v1/client-keys/${id}/status`, { method: 'POST', body: JSON.stringify({ status }) }),
  getClientKeyQuota: (id: number) => apiRequest<ClientKey>(`/admin/api/v1/client-keys/${id}/quota`),
  updateClientKeyQuota: (id: number, body: Partial<ClientKey>) => apiRequest<ClientKey>(`/admin/api/v1/client-keys/${id}/quota`, { method: 'PUT', body: JSON.stringify(body) }),
  testRequest: (body: { api_key: string; model: string; messages: { role: string; content: string }[]; stream?: boolean; temperature?: number; max_tokens?: number }) => apiRequest<{ success: boolean; response?: string; error?: string; key_status?: string; latency_ms?: number }>('/admin/api/v1/tools/test-request', { method: 'POST', body: JSON.stringify(body) }),
  exportConfig: (query = '') => apiRequest<any>(`/admin/api/v1/config/export${query ? `?${query}` : ''}`),
  importConfig: (body: any) => apiRequest<any>('/admin/api/v1/config/import', { method: 'POST', body: JSON.stringify(body) }),
  requestLogs: (query = '') => apiRequest<PaginatedResponse<RequestLog>>(`/admin/api/v1/logs/requests${query ? `?${query}` : ''}`),
  requestLogDetail: (id: number) => apiRequest<RequestLog>(`/admin/api/v1/logs/requests/${id}`),
  auditLogs: (query = '') => apiRequest<PaginatedResponse<AuditLog>>(`/admin/api/v1/audit-logs${query ? `?${query}` : ''}`),
  auditLogDetail: (id: number) => apiRequest<AuditLog>(`/admin/api/v1/audit-logs/${id}`),
  downloadAuditLogs: (query = '') => apiDownload(`/admin/api/v1/audit-logs/export${query ? `?${query}` : ''}`),
}
