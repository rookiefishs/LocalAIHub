'use client'

import { useState, useEffect, useContext } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { FiLogOut, FiSun, FiMoon, FiRefreshCw } from 'react-icons/fi'
import { GoDotFill } from 'react-icons/go'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { clearToken } from '@/lib/auth'
import { useRefresh } from './refresh-context'

const pageTitles: Record<string, string> = {
  '/dashboard': '仪表盘',
  '/dashboard/wizard': '快捷流程',
  '/dashboard/analytics': '统计分析',
  '/dashboard/upstreams': '上游管理',
  '/dashboard/models': '虚拟模型管理',
  '/dashboard/routes': '路由管理',
  '/dashboard/keys': 'API Key 管理',
  '/dashboard/logs': '日志中心',
  '/dashboard/help': '使用教程',
}

export function DashboardHeader() {
  const pathname = usePathname()
  const router = useRouter()
  const [theme, setTheme] = useState<'light' | 'dark'>('light')
  const [refreshing, setRefreshing] = useState(false)
  const title = pageTitles[pathname] || '管理后台'
  const [confirmOpen, setConfirmOpen] = useState(false)
  const { triggerRefresh } = useRefresh()

  useEffect(() => {
    const saved = localStorage.getItem('theme') as 'light' | 'dark' | null
    if (saved) {
      setTheme(saved)
      document.documentElement.classList.toggle('dark', saved === 'dark')
    } else {
      document.documentElement.classList.add('dark')
    }
  }, [])

  function toggleTheme() {
    const next = theme === 'light' ? 'dark' : 'light'
    setTheme(next)
    localStorage.setItem('theme', next)
    document.documentElement.classList.toggle('dark', next === 'dark')
  }

  return (
    <>
      <header className="header-shell sticky top-0 z-20 mb-2 flex h-14 items-center gap-4 px-4" style={{ borderRadius: '10px' }}>
        <div className="flex-1">
          <h1 className="text-base font-medium" style={{ color: 'var(--foreground)' }}>{title}</h1>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" loading={refreshing} onClick={async () => { setRefreshing(true); triggerRefresh(); setTimeout(() => setRefreshing(false), 1000) }}>
            <FiRefreshCw className="mr-1.5 h-3.5 w-3.5" />
            刷新
          </Button>
          <Button variant="secondary" disabled className="!opacity-100 !cursor-default">
            <GoDotFill className="mr-1.5 h-3 w-3 text-emerald-400" />
            正常
          </Button>
          <Button variant="secondary" onClick={toggleTheme}>
            {theme === 'light' ? <FiMoon className="h-4 w-4" /> : <FiSun className="h-4 w-4" />}
          </Button>
          <Button variant="destructive" size="icon" onClick={() => setConfirmOpen(true)}>
            <FiLogOut className="h-4 w-4" />
          </Button>
        </div>
      </header>

      <ConfirmDialog
        open={confirmOpen}
        title="确认退出登录"
        description="退出后将清除当前管理员登录状态，并返回登录页。"
        confirmText="确认退出"
        confirmVariant="destructive"
        onCancel={() => setConfirmOpen(false)}
        onConfirm={() => {
          clearToken()
          setConfirmOpen(false)
          router.push('/login')
        }}
      />
    </>
  )
}