import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

const textMeaningMap: Record<string, string> = {
  unauthorized: '未授权，请重新登录',
  forbidden: '无权限访问',
  timeout: '请求超时',
  'context deadline exceeded': '请求超时',
  'deadline exceeded': '请求超时',
  'quota exceeded': '超出配额限制',
  'daily request limit exceeded': '超出每日请求次数限制',
  'monthly request limit exceeded': '超出每月请求次数限制',
  'daily token limit exceeded': '超出每日 Token 限制',
  'monthly token limit exceeded': '超出每月 Token 限制',
  'api key disabled due to quota exceeded': 'API Key 因配额超限被禁用',
  'invalid api key': 'API Key 无效',
  'invalid_api_key': 'API Key 无效',
  'incorrect api key': 'API Key 不正确',
  'authentication failed': '认证失败',
  'auth failed': '认证失败',
  'rate limit exceeded': '触发速率限制',
  rate_limit_exceeded: '触发速率限制',
  'too many requests': '请求过于频繁',
  'upstream request failed': '上游请求失败',
  'upstream error': '上游服务错误',
  'provider not found': '上游不存在',
  'model not found': '模型不存在',
  'route not found': '路由不存在',
  'service unavailable': '服务暂时不可用',
  'internal server error': '服务内部错误',
  'manual lock': '手动锁定',
  manual_locked: '手动锁定',
  switched: '已切换',
  fallback: '已切换到备用线路',
  auto_switch: '自动切换',
  auto_switched: '自动切换',
  'auto switch': '自动切换',
  'health check failed': '健康检查失败',
  'circuit breaker open': '熔断器已打开',
  'provider disabled': '上游已禁用',
  'key disabled': '密钥已禁用',
}

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function describeTextMeaning(value?: string | null): string {
  const text = value?.trim()
  if (!text) return '-'

  const lower = text.toLowerCase()
  const direct = textMeaningMap[lower]
  if (direct) {
    return `${text}（${direct}）`
  }

  const matched = Object.entries(textMeaningMap).find(([key]) => lower.includes(key))
  if (matched) {
    return `${text}（${matched[1]}）`
  }

  return text
}
