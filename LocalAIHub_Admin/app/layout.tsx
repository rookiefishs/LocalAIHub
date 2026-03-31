import './globals.css'
import type { ReactNode } from 'react'
import { ToastProvider } from '@/components/ui/toast'

export const metadata = {
  title: 'LocalAIHub - 管理后台',
  description: '本地 AI 中转网关管理后台',
  icons: {
    icon: '/logo.png',
  },
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="zh-CN">
      <body><ToastProvider>{children}</ToastProvider></body>
    </html>
  )
}
