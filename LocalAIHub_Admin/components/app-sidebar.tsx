'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { FiLayers, FiChevronLeft, FiChevronRight, FiHelpCircle } from 'react-icons/fi'
import { HiOutlineKey, HiOutlineBookOpen } from 'react-icons/hi2'
import { LuLogs } from 'react-icons/lu'
import { MdOutlineRoute, MdOutlineSpaceDashboard } from 'react-icons/md'
import { TbPlugConnected, TbTopologyStarRing3 } from 'react-icons/tb'

const mainNavItems = [
  { title: '仪表盘', href: '/dashboard', icon: MdOutlineSpaceDashboard },
  { title: '上游管理', href: '/dashboard/upstreams', icon: TbPlugConnected },
  { title: '虚拟模型', href: '/dashboard/models', icon: FiLayers },
  { title: '路由管理', href: '/dashboard/routes', icon: MdOutlineRoute },
  { title: 'API Key', href: '/dashboard/keys', icon: HiOutlineKey },
]

const systemNavItems = [
  { title: '日志中心', href: '/dashboard/logs', icon: LuLogs },
  { title: '使用教程', href: '/dashboard/help', icon: HiOutlineBookOpen },
]

interface AppSidebarProps {
  collapsed?: boolean
  onToggle?: () => void
}

export function AppSidebar({ collapsed = false, onToggle }: AppSidebarProps) {
  const pathname = usePathname()

  return (
    <aside className="sidebar-shell hidden h-screen lg:block" style={{ width: '100%' }}>
      <div className="flex flex-col h-full overflow-hidden transition-all duration-300 ease-in-out" style={{ width: collapsed ? 64 : 220 }}>
        <div className="flex items-center justify-between p-3" style={{ borderBottom: '1px solid var(--sidebar-border)' }}>
          <Link href="/dashboard" className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg" style={{ background: 'rgba(56, 201, 212, 0.15)', border: '1px solid rgba(56, 201, 212, 0.25)' }}>
              <TbTopologyStarRing3 className="h-4 w-4" style={{ color: 'var(--primary)' }} />
            </div>
            {!collapsed && (
              <div>
                <h2 className="text-sm font-medium leading-tight" style={{ color: 'var(--sidebar-foreground)' }}>AI Gateway</h2>
                <p className="text-xs" style={{ color: 'var(--muted-foreground)' }}>管理后台</p>
              </div>
            )}
          </Link>
          <button
            onClick={onToggle}
            className="flex h-6 w-6 items-center justify-center rounded transition-colors hover:bg-[var(--sidebar-accent)]"
            style={{ color: 'var(--muted-foreground)' }}
          >
            {collapsed ? <FiChevronRight className="h-3.5 w-3.5" /> : <FiChevronLeft className="h-3.5 w-3.5" />}
          </button>
        </div>

        <div className="flex-1 overflow-y-auto py-3">
          <div className="mb-4 px-3">
            {!collapsed && <div className="mb-2 text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--muted-foreground)' }}>核心功能</div>}
            <div className="space-y-1">
              {mainNavItems.map((item) => {
                const Icon = item.icon
                const active = pathname === item.href
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-all"
                    style={{
                      background: active ? 'rgba(56, 201, 212, 0.14)' : 'transparent',
                      color: active ? '#7de7ef' : 'var(--sidebar-foreground)',
                      justifyContent: collapsed ? 'center' : 'flex-start',
                    }}
                  >
                    <Icon className="h-4 w-4 flex-shrink-0" />
                    {!collapsed && <span>{item.title}</span>}
                  </Link>
                )
              })}
            </div>
          </div>

          <div className="px-3">
            {!collapsed && <div className="mb-2 text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--muted-foreground)' }}>系统管理</div>}
            <div className="space-y-1">
              {systemNavItems.map((item) => {
                const Icon = item.icon
                const active = pathname === item.href
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-all"
                    style={{
                      background: active ? 'rgba(56, 201, 212, 0.14)' : 'transparent',
                      color: active ? '#7de7ef' : 'var(--sidebar-foreground)',
                      justifyContent: collapsed ? 'center' : 'flex-start',
                    }}
                  >
                    <Icon className="h-4 w-4 flex-shrink-0" />
                    {!collapsed && <span>{item.title}</span>}
                  </Link>
                )
              })}
            </div>
          </div>
        </div>
      </div>
    </aside>
  )
}
