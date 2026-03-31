'use client'

import { useState } from 'react'
import type { ReactNode } from 'react'
import { AppSidebar } from '@/components/app-sidebar'
import { DashboardHeader } from '@/components/dashboard-header'
import { PageTransition } from '@/components/page-transition'

export function AppShell({ children }: { children: ReactNode }) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  return (
    <div className="h-screen overflow-hidden lg:flex">
      <div className="transition-all duration-300 ease-in-out" style={{ width: sidebarCollapsed ? 64 : 220, flexShrink: 0 }}>
        <AppSidebar collapsed={sidebarCollapsed} onToggle={() => setSidebarCollapsed(!sidebarCollapsed)} />
      </div>
      <div className="flex-1 min-w-0 overflow-hidden p-3 md:p-4">
        <DashboardHeader />
        <main className="h-[calc(100vh-6rem)] overflow-y-auto pr-1">
          <PageTransition>{children}</PageTransition>
        </main>
      </div>
    </div>
  )
}
