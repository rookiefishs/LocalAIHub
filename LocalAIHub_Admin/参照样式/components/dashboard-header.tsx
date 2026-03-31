"use client"

import { usePathname } from "next/navigation"
import { SidebarTrigger } from "@/components/ui/sidebar"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { BellIcon, UserIcon, CircleIcon } from "lucide-react"

const pageTitles: Record<string, string> = {
  "/dashboard": "仪表盘",
  "/dashboard/upstreams": "上游管理",
  "/dashboard/models": "虚拟模型管理",
  "/dashboard/routes": "路由管理",
  "/dashboard/keys": "API Key 管理",
  "/dashboard/logs": "日志中心",
  "/dashboard/debug": "调试控制",
  "/dashboard/settings": "系统设置",
}

export function DashboardHeader() {
  const pathname = usePathname()
  const title = pageTitles[pathname] || "管理后台"

  return (
    <header className="flex h-14 items-center gap-4 border-b border-border/50 bg-card/30 px-4">
      <SidebarTrigger />
      <Separator orientation="vertical" className="h-6" />

      <div className="flex-1">
        <h1 className="font-semibold text-lg">{title}</h1>
      </div>

      {/* System Status */}
      <div className="flex items-center gap-2">
        <Badge variant="outline" className="gap-1.5 text-xs font-normal">
          <CircleIcon className="h-2 w-2 fill-green-500 text-green-500" />
          系统正常
        </Badge>
      </div>

      {/* Notifications */}
      <Button variant="ghost" size="icon" className="relative">
        <BellIcon className="h-4 w-4" />
        <span className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-destructive" />
        <span className="sr-only">通知</span>
      </Button>

      {/* User Menu */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <UserIcon className="h-4 w-4" />
            <span className="sr-only">用户菜单</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-48">
          <DropdownMenuLabel>
            <div className="flex flex-col">
              <span className="font-medium">管理员</span>
              <span className="text-xs text-muted-foreground">admin@local</span>
            </div>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem>个人设置</DropdownMenuItem>
          <DropdownMenuItem>帮助文档</DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem className="text-destructive">退出登录</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
