export const TOKEN_KEY = 'localaihub_admin_token'
export const REFRESH_TOKEN_KEY = 'localaihub_admin_refresh_token'
export const TOKEN_CREATE_TIME_KEY = 'localaihub_admin_token_create_time'

const TOKEN_EXPIRE_MS = 60 * 60 * 1000 // 1 hour

export function getToken() {
  if (typeof window === 'undefined') return ''
  return window.localStorage.getItem(TOKEN_KEY) || ''
}

export function getRefreshToken() {
  if (typeof window === 'undefined') return ''
  return window.localStorage.getItem(REFRESH_TOKEN_KEY) || ''
}

export function setToken(token: string, refreshToken?: string) {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(TOKEN_KEY, token)
  window.localStorage.setItem(TOKEN_CREATE_TIME_KEY, Date.now().toString())
  if (refreshToken) {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  }
}

export function clearToken() {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(TOKEN_KEY)
  window.localStorage.removeItem(REFRESH_TOKEN_KEY)
  window.localStorage.removeItem(TOKEN_CREATE_TIME_KEY)
}

export function isTokenExpired(): boolean {
  if (typeof window === 'undefined') return true
  const createTime = localStorage.getItem(TOKEN_CREATE_TIME_KEY)
  if (!createTime) return true
  return Date.now() - parseInt(createTime) > TOKEN_EXPIRE_MS
}
