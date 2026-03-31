"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Field, FieldLabel } from "@/components/ui/field"
import { Switch } from "@/components/ui/switch"
import {
  MoreHorizontalIcon,
  GitBranchIcon,
  CircleIcon,
  ArrowRightIcon,
  RefreshCwIcon,
  LockIcon,
  UnlockIcon,
  AlertTriangleIcon,
  ZapIcon,
  HistoryIcon,
} from "lucide-react"

// Mock data
const routes = [
  {
    id: "route-001",
    virtualModel: "gpt-4o",
    displayName: "GPT-4 Omni",
    currentUpstream: "OpenAI 官方",
    currentModel: "gpt-4o",
    routeStatus: "normal",
    routeMode: "auto",
    candidates: [
      { upstream: "OpenAI 官方", model: "gpt-4o", priority: 1, status: "healthy", latency: 120 },
      { upstream: "中转站 A", model: "gpt-4o", priority: 2, status: "healthy", latency: 89 },
    ],
    lastSwitchAt: null,
    lastSwitchReason: null,
    manualLockUntil: null,
  },
  {
    id: "route-002",
    virtualModel: "gpt-4-turbo",
    displayName: "GPT-4 Turbo",
    currentUpstream: "中转站 A",
    currentModel: "gpt-4-turbo",
    routeStatus: "normal",
    routeMode: "manual",
    candidates: [
      { upstream: "OpenAI 官方", model: "gpt-4-turbo-preview", priority: 1, status: "healthy", latency: 156 },
      { upstream: "中转站 A", model: "gpt-4-turbo", priority: 2, status: "healthy", latency: 78 },
    ],
    lastSwitchAt: "2小时前",
    lastSwitchReason: "手动切换",
    manualLockUntil: "2026-03-31 12:00",
  },
  {
    id: "route-003",
    virtualModel: "claude-sonnet",
    displayName: "Claude 3.5 Sonnet",
    currentUpstream: "Anthropic 官方",
    currentModel: "claude-3-5-sonnet-20241022",
    routeStatus: "breaker",
    routeMode: "auto",
    candidates: [
      { upstream: "Anthropic 官方", model: "claude-3-5-sonnet-20241022", priority: 1, status: "healthy", latency: 134 },
      { upstream: "中转站 B", model: "claude-3-5-sonnet", priority: 2, status: "breaker", latency: 0 },
    ],
    lastSwitchAt: "30分钟前",
    lastSwitchReason: "自动熔断",
    manualLockUntil: null,
  },
  {
    id: "route-004",
    virtualModel: "gemini-pro",
    displayName: "Gemini Pro",
    currentUpstream: "Gemini 官方",
    currentModel: "gemini-pro",
    routeStatus: "degraded",
    routeMode: "auto",
    candidates: [
      { upstream: "Gemini 官方", model: "gemini-pro", priority: 1, status: "degraded", latency: 340 },
    ],
    lastSwitchAt: null,
    lastSwitchReason: null,
    manualLockUntil: null,
  },
]

const statusConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  normal: { label: "正常", color: "text-green-500", bgColor: "bg-green-500/10" },
  degraded: { label: "降级", color: "text-yellow-500", bgColor: "bg-yellow-500/10" },
  breaker: { label: "已熔断", color: "text-orange-500", bgColor: "bg-orange-500/10" },
  down: { label: "不可用", color: "text-red-500", bgColor: "bg-red-500/10" },
  healthy: { label: "健康", color: "text-green-500", bgColor: "bg-green-500/10" },
}

const switchHistory = [
  {
    id: "sw-001",
    time: "30 分钟前",
    model: "claude-sonnet",
    from: "中转站 B",
    to: "Anthropic 官方",
    reason: "自动熔断 - 连续失败次数达到阈值",
    operator: "系统",
  },
  {
    id: "sw-002",
    time: "2 小时前",
    model: "gpt-4-turbo",
    from: "OpenAI 官方",
    to: "中转站 A",
    reason: "手动切换 - 成本优化",
    operator: "admin",
  },
  {
    id: "sw-003",
    time: "1 天前",
    model: "gpt-4o",
    from: "中转站 A",
    to: "OpenAI 官方",
    reason: "自动恢复 - 主上游恢复健康",
    operator: "系统",
  },
]

export default function RoutesPage() {
  const [isSwitchDialogOpen, setIsSwitchDialogOpen] = useState(false)
  const [selectedRoute, setSelectedRoute] = useState<typeof routes[0] | null>(null)

  const handleSwitchClick = (route: typeof routes[0]) => {
    setSelectedRoute(route)
    setIsSwitchDialogOpen(true)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">全局模型路由</h2>
          <p className="text-sm text-muted-foreground">
            查看和管理每个虚拟模型当前的路由状态，支持手动切换和锁定
          </p>
        </div>
        <Button variant="outline" size="sm">
          <RefreshCwIcon className="mr-2 h-4 w-4" />
          刷新状态
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-primary/10">
                <GitBranchIcon className="h-4 w-4 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold">4</p>
                <p className="text-xs text-muted-foreground">总路由数</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green-500/10">
                <CircleIcon className="h-4 w-4 text-green-500" fill="currentColor" />
              </div>
              <div>
                <p className="text-2xl font-bold">2</p>
                <p className="text-xs text-muted-foreground">正常路由</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-orange-500/10">
                <AlertTriangleIcon className="h-4 w-4 text-orange-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">1</p>
                <p className="text-xs text-muted-foreground">熔断中</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue-500/10">
                <LockIcon className="h-4 w-4 text-blue-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">1</p>
                <p className="text-xs text-muted-foreground">手动锁定</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Routes Table */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base">路由列表</CardTitle>
          <CardDescription>当前生效的模型路由配置</CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>状态</TableHead>
                <TableHead>虚拟模型</TableHead>
                <TableHead>当前路由</TableHead>
                <TableHead>候选数</TableHead>
                <TableHead>模式</TableHead>
                <TableHead>最近切换</TableHead>
                <TableHead className="w-12"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {routes.map((route) => {
                const status = statusConfig[route.routeStatus]
                return (
                  <TableRow key={route.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className={`p-1.5 rounded-full ${status.bgColor}`}>
                          <CircleIcon
                            className={`h-2 w-2 ${status.color}`}
                            fill="currentColor"
                          />
                        </div>
                        <span className="text-sm">{status.label}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div>
                        <p className="font-medium font-mono">{route.virtualModel}</p>
                        <p className="text-xs text-muted-foreground">
                          {route.displayName}
                        </p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div>
                        <p className="text-sm">{route.currentUpstream}</p>
                        <p className="text-xs text-muted-foreground font-mono">
                          {route.currentModel}
                        </p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        {route.candidates.map((c, idx) => (
                          <div
                            key={idx}
                            className={`h-2 w-2 rounded-full ${statusConfig[c.status]?.color || 'text-gray-400'}`}
                            style={{ backgroundColor: 'currentColor' }}
                            title={`${c.upstream}: ${statusConfig[c.status]?.label}`}
                          />
                        ))}
                        <span className="text-sm text-muted-foreground ml-1">
                          {route.candidates.length} 个
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={route.routeMode === "manual" ? "default" : "outline"}
                        className="text-xs gap-1"
                      >
                        {route.routeMode === "manual" ? (
                          <>
                            <LockIcon className="h-3 w-3" />
                            手动
                          </>
                        ) : (
                          "自动"
                        )}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {route.lastSwitchAt ? (
                        <div>
                          <p className="text-sm">{route.lastSwitchAt}</p>
                          <p className="text-xs text-muted-foreground">
                            {route.lastSwitchReason}
                          </p>
                        </div>
                      ) : (
                        <span className="text-sm text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <MoreHorizontalIcon className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleSwitchClick(route)}>
                            <ZapIcon className="mr-2 h-4 w-4" />
                            手动切换
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            {route.routeMode === "manual" ? (
                              <>
                                <UnlockIcon className="mr-2 h-4 w-4" />
                                解除锁定
                              </>
                            ) : (
                              <>
                                <LockIcon className="mr-2 h-4 w-4" />
                                锁定路由
                              </>
                            )}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem>
                            <HistoryIcon className="mr-2 h-4 w-4" />
                            查看历史
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Switch History */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <HistoryIcon className="h-4 w-4" />
            切换历史
          </CardTitle>
          <CardDescription>最近的路由切换记录</CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>时间</TableHead>
                <TableHead>模型</TableHead>
                <TableHead>切换路径</TableHead>
                <TableHead>原因</TableHead>
                <TableHead>操作人</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {switchHistory.map((record) => (
                <TableRow key={record.id}>
                  <TableCell className="text-muted-foreground">
                    {record.time}
                  </TableCell>
                  <TableCell className="font-medium font-mono">
                    {record.model}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span className="text-muted-foreground">{record.from}</span>
                      <ArrowRightIcon className="h-3 w-3 text-muted-foreground" />
                      <span className="text-primary">{record.to}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm max-w-[300px] truncate">
                    {record.reason}
                  </TableCell>
                  <TableCell>
                    <Badge variant={record.operator === "系统" ? "secondary" : "outline"} className="text-xs">
                      {record.operator}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Switch Dialog */}
      <Dialog open={isSwitchDialogOpen} onOpenChange={setIsSwitchDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>手动切换路由</DialogTitle>
            <DialogDescription>
              为 {selectedRoute?.virtualModel} 选择新的目标上游
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <Field>
              <FieldLabel>目标上游</FieldLabel>
              <Select defaultValue={selectedRoute?.candidates[0]?.upstream}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {selectedRoute?.candidates.map((c, idx) => (
                    <SelectItem key={idx} value={c.upstream}>
                      <div className="flex items-center gap-2">
                        <CircleIcon
                          className={`h-2 w-2 ${statusConfig[c.status]?.color}`}
                          fill="currentColor"
                        />
                        {c.upstream} - {c.model}
                        <span className="text-muted-foreground">({c.latency}ms)</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">锁定路由</p>
                <p className="text-xs text-muted-foreground">
                  锁定后不会自动切换回其他上游
                </p>
              </div>
              <Switch />
            </div>
            <Field>
              <FieldLabel>锁定时长</FieldLabel>
              <Select defaultValue="1h">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="30m">30 分钟</SelectItem>
                  <SelectItem value="1h">1 小时</SelectItem>
                  <SelectItem value="6h">6 小时</SelectItem>
                  <SelectItem value="24h">24 小时</SelectItem>
                  <SelectItem value="forever">永久</SelectItem>
                </SelectContent>
              </Select>
            </Field>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsSwitchDialogOpen(false)}>
              取消
            </Button>
            <Button>确认切换</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
