import './globals.css'
import type { ReactNode } from 'react'
import { ToastProvider } from '@/components/ui/toast'

export const metadata = {
  title: 'LocalAIHub - 管理后台',
  description: '本地 AI 中转网关管理后台',
  icons: {
    icon: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect x="8" y="8" width="48" height="48" rx="14" fill="%230F1115"/><path d="M22 22L32 16L42 22V42L32 48L22 42V22Z" stroke="%23FFFFFF" stroke-width="3" stroke-linejoin="round" fill="none"/><path d="M22 22L32 28L42 22" stroke="%23FFFFFF" stroke-width="3" stroke-linejoin="round" fill="none"/><path d="M32 28V48" stroke="%23FFFFFF" stroke-width="3" stroke-linejoin="round" fill="none"/></svg>',
  },
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="zh-CN">
      <body><ToastProvider>{children}</ToastProvider></body>
    </html>
  )
}
