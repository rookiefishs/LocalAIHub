"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarFooter,
  SidebarSeparator,
} from "@/components/ui/sidebar"
import {
  LayoutDashboardIcon,
  ServerIcon,
  BoxIcon,
  GitBranchIcon,
  KeyIcon,
  FileTextIcon,
  BugIcon,
  SettingsIcon,
  NetworkIcon,
  LogOutIcon,
} from "lucide-react"
import { Button } from "@/components/ui/button"

const mainNavItems = [
  {
    title: "仪表盘",
    href: "/dashboard",
    icon: LayoutDashboardIcon,
  },
  {
    title: "上游管理",
    href: "/dashboard/upstreams",
    icon: ServerIcon,
  },
  {
    title: "虚拟模型",
    href: "/dashboard/models",
    icon: BoxIcon,
  },
  {
    title: "路由管理",
    href: "/dashboard/routes",
    icon: GitBranchIcon,
  },
  {
    title: "API Key",
    href: "/dashboard/keys",
    icon: KeyIcon,
  },
]

const systemNavItems = [
  {
    title: "日志中心",
    href: "/dashboard/logs",
    icon: FileTextIcon,
  },
  {
    title: "调试控制",
    href: "/dashboard/debug",
    icon: BugIcon,
  },
  {
    title: "系统设置",
    href: "/dashboard/settings",
    icon: SettingsIcon,
  },
]

export function AppSidebar() {
  const pathname = usePathname()

  return (
    <Sidebar>
      <SidebarHeader className="p-4">
        <Link href="/dashboard" className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-primary/10 border border-primary/20">
            <NetworkIcon className="h-5 w-5 text-primary" />
          </div>
          <div>
            <h2 className="font-semibold text-sm">AI Gateway</h2>
            <p className="text-xs text-muted-foreground">管理后台</p>
          </div>
        </Link>
      </SidebarHeader>

      <SidebarSeparator />

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>核心功能</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainNavItems.map((item) => (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton
                    asChild
                    isActive={pathname === item.href}
                    tooltip={item.title}
                  >
                    <Link href={item.href}>
                      <item.icon className="h-4 w-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>系统管理</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {systemNavItems.map((item) => (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton
                    asChild
                    isActive={pathname === item.href}
                    tooltip={item.title}
                  >
                    <Link href={item.href}>
                      <item.icon className="h-4 w-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="p-4">
        <Button variant="ghost" className="w-full justify-start" asChild>
          <Link href="/login">
            <LogOutIcon className="mr-2 h-4 w-4" />
            退出登录
          </Link>
        </Button>
      </SidebarFooter>
    </Sidebar>
  )
}
