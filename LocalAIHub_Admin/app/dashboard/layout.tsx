'use client'

import { ReactNode, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/app-shell'
import { getToken } from '@/lib/auth'
import { RefreshProvider } from '@/components/refresh-context'

export default function DashboardLayout({ children }: { children: ReactNode }) {
  const router = useRouter()

  useEffect(() => {
    if (!getToken()) {
      router.replace('/login')
    }
  }, [router])

  return (
    <RefreshProvider>
      <AppShell>{children}</AppShell>
    </RefreshProvider>
  )
}
