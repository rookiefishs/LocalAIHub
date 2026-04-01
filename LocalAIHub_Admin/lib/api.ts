import { getToken } from '@/lib/auth'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || (process.env.NODE_ENV === 'development' ? 'http://127.0.0.1:3334' : '/localaihub-api')

type RequestOptions = RequestInit & {
  auth?: boolean
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers || {})
  headers.set('Content-Type', 'application/json')
  if (options.auth !== false) {
    const token = getToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    cache: 'no-store',
  })

  const payload = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(payload?.message || payload?.error?.message || `Request failed: ${response.status}`)
  }

  if (payload && typeof payload === 'object' && 'code' in payload) {
    return payload.data as T
  }

  return payload as T
}

export const api = {
  login: (body: { username: string; password: string }) => apiRequest<{ user: { id: number; username: string }; token: string }>('/admin/api/v1/auth/login', { method: 'POST', auth: false, body: JSON.stringify(body) }),
  me: () => apiRequest<{ user: { id: number; username: string; status: string } }>('/admin/api/v1/auth/me'),
  dashboard: (query = '') => apiRequest<any>(`/admin/api/v1/dashboard/overview${query ? `?${query}` : ''}`),
  providers: (query = '') => apiRequest<any>(`/admin/api/v1/providers${query ? `?${query}` : ''}`),
  createProvider: (body: any) => apiRequest<any>('/admin/api/v1/providers', { method: 'POST', body: JSON.stringify(body) }),
  updateProvider: (id: number, body: any) => apiRequest<any>(`/admin/api/v1/providers/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteProvider: (id: number) => apiRequest<any>(`/admin/api/v1/providers/${id}`, { method: 'DELETE' }),
  updateProviderStatus: (id: number, enabled: boolean) => apiRequest<any>(`/admin/api/v1/providers/${id}/status`, { method: 'POST', body: JSON.stringify({ enabled }) }),
  testProvider: (id: number) => apiRequest<any>(`/admin/api/v1/providers/${id}/test-connection`, { method: 'POST', body: JSON.stringify({}) }),
  providerKeys: (id: number) => apiRequest<any>(`/admin/api/v1/providers/${id}/keys`),
  createProviderKey: (id: number, body: any) => apiRequest<any>(`/admin/api/v1/providers/${id}/keys`, { method: 'POST', body: JSON.stringify(body) }),
  deleteProviderKey: (providerId: number, keyId: number) => apiRequest<any>(`/admin/api/v1/providers/${providerId}/keys/${keyId}`, { method: 'DELETE' }),
  updateProviderKeyStatus: (providerId: number, keyId: number, status: string) => apiRequest<any>(`/admin/api/v1/providers/${providerId}/keys/${keyId}/status`, { method: 'POST', body: JSON.stringify({ status }) }),
  models: (query = '') => apiRequest<any>(`/admin/api/v1/models${query ? `?${query}` : ''}`),
  createModel: (body: any) => apiRequest<any>('/admin/api/v1/models', { method: 'POST', body: JSON.stringify(body) }),
  updateModel: (id: number, body: any) => apiRequest<any>(`/admin/api/v1/models/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteModel: (id: number) => apiRequest<any>(`/admin/api/v1/models/${id}`, { method: 'DELETE' }),
  modelBindings: (id: number) => apiRequest<any>(`/admin/api/v1/models/${id}/bindings`),
  createModelBinding: (id: number, body: any) => apiRequest<any>(`/admin/api/v1/models/${id}/bindings`, { method: 'POST', body: JSON.stringify(body) }),
  updateModelBinding: (modelId: number, bindingId: number, body: any) => apiRequest<any>(`/admin/api/v1/models/${modelId}/bindings/${bindingId}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteModelBinding: (modelId: number, bindingId: number) => apiRequest<any>(`/admin/api/v1/models/${modelId}/bindings/${bindingId}`, { method: 'DELETE' }),
  testModelBinding: (modelId: number, bindingId: number) => apiRequest<any>(`/admin/api/v1/models/${modelId}/bindings/${bindingId}/test`, { method: 'POST', body: JSON.stringify({}) }),
  routes: (query = '') => apiRequest<any>(`/admin/api/v1/routes${query ? `?${query}` : ''}`),
  switchRoute: (id: number, body: any) => apiRequest<any>(`/admin/api/v1/routes/${id}/switch`, { method: 'POST', body: JSON.stringify(body) }),
  unlockRoute: (id: number) => apiRequest<any>(`/admin/api/v1/routes/${id}/unlock`, { method: 'POST', body: JSON.stringify({}) }),
  deleteRoute: (id: number) => apiRequest<any>(`/admin/api/v1/routes/${id}`, { method: 'DELETE' }),
  clientKeys: (query = '') => apiRequest<any>(`/admin/api/v1/client-keys${query ? `?${query}` : ''}`),
  createClientKey: (body: any) => apiRequest<any>('/admin/api/v1/client-keys', { method: 'POST', body: JSON.stringify(body) }),
  getClientKey: (id: number) => apiRequest<any>(`/admin/api/v1/client-keys/${id}`),
  testClientKey: (id: number) => apiRequest<any>(`/admin/api/v1/client-keys/${id}/test`, { method: 'POST', body: JSON.stringify({}) }),
  updateClientKey: (id: number, body: any) => apiRequest<any>(`/admin/api/v1/client-keys/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteClientKey: (id: number) => apiRequest<any>(`/admin/api/v1/client-keys/${id}`, { method: 'DELETE' }),
  updateClientKeyStatus: (id: number, status: string) => apiRequest<any>(`/admin/api/v1/client-keys/${id}/status`, { method: 'POST', body: JSON.stringify({ status }) }),
  getClientKeyQuota: (id: number) => apiRequest<any>(`/admin/api/v1/client-keys/${id}/quota`),
  updateClientKeyQuota: (id: number, body: any) => apiRequest<any>(`/admin/api/v1/client-keys/${id}/quota`, { method: 'PUT', body: JSON.stringify(body) }),
  requestLogs: (query = '') => apiRequest<any>(`/admin/api/v1/logs/requests${query ? `?${query}` : ''}`),
  requestLogDetail: (id: number) => apiRequest<any>(`/admin/api/v1/logs/requests/${id}`),
  auditLogs: (query = '') => apiRequest<any>(`/admin/api/v1/logs/audit${query ? `?${query}` : ''}`),
}
