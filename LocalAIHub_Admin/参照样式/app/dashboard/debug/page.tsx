"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Field, FieldLabel } from "@/components/ui/field"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import {
  BugIcon,
  PlayIcon,
  StopCircleIcon,
  ClockIcon,
  AlertTriangleIcon,
  ShieldAlertIcon,
  EyeIcon,
} from "lucide-react"

// Mock active debug session
const activeDebugSession = {
  id: "debug-001",
  scopeType: "global",
  scopeValue: null,
  enabled: true,
  startedAt: "2024-03-30 14:00:00",
  expiresAt: "2024-03-30 15:00:00",
  startedBy: "admin",
  reason: "排查 GPT-4o 响应异常问题",
  remainingMinutes: 28,
}

const debugHistory = [
  {
    id: "debug-001",
    scopeType: "global",
    scopeValue: null,
    startedAt: "2024-03-30 14:00:00",
    endedAt: null,
    duration: "28 分钟",
    startedBy: "admin",
    reason: "排查 GPT-4o 响应异常问题",
    status: "active",
  },
  {
    id: "debug-002",
    scopeType: "model",
    scopeValue: "claude-sonnet",
    startedAt: "2024-03-29 10:30:00",
    endedAt: "2024-03-29 11:30:00",
    duration: "1 小时",
    startedBy: "admin",
    reason: "Claude 模型熔断原因排查",
    status: "completed",
  },
  {
    id: "debug-003",
    scopeType: "api_key",
    scopeValue: "sk-prod-...a3f7",
    startedAt: "2024-03-28 16:00:00",
    endedAt: "2024-03-28 16:30:00",
    duration: "30 分钟",
    startedBy: "admin",
    reason: "生产环境 Key 调用异常",
    status: "completed",
  },
]

const scopeTypeLabels: Record<string, string> = {
  global: "全局",
  model: "指定模型",
  api_key: "指定 API Key",
}

export default function DebugPage() {
  const [isStartDialogOpen, setIsStartDialogOpen] = useState(false)
  const [remainingTime, setRemainingTime] = useState(activeDebugSession?.remainingMinutes || 0)

  // Countdown timer simulation
  useEffect(() => {
    if (activeDebugSession?.enabled && remainingTime > 0) {
      const timer = setInterval(() => {
        setRemainingTime((prev) => Math.max(0, prev - 1/60))
      }, 1000)
      return () => clearInterval(timer)
    }
  }, [remainingTime])

  const formatRemainingTime = (minutes: number) => {
    const mins = Math.floor(minutes)
    const secs = Math.floor((minutes - mins) * 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  return (
    <div className="space-y-6">
      {/* Active Debug Banner */}
      {activeDebugSession?.enabled && (
        <div className="p-4 rounded-lg bg-yellow-500/10 border border-yellow-500/30">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-yellow-500/20">
                <BugIcon className="h-5 w-5 text-yellow-500" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <p className="font-medium text-yellow-500">调试模式已开启</p>
                  <Badge variant="outline" className="text-yellow-500 border-yellow-500/50">
                    {scopeTypeLabels[activeDebugSession.scopeType]}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">
                  {activeDebugSession.reason}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="text-right">
                <p className="text-sm text-muted-foreground">剩余时间</p>
                <p className="text-2xl font-mono font-bold text-yellow-500">
                  {formatRemainingTime(remainingTime)}
                </p>
              </div>
              <Button variant="destructive" size="sm">
                <StopCircleIcon className="mr-2 h-4 w-4" />
                立即关闭
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">调试控制</h2>
          <p className="text-sm text-muted-foreground">
            临时开启调试模式以获取更详细的日志信息，用于问题排查
          </p>
        </div>
        <Dialog open={isStartDialogOpen} onOpenChange={setIsStartDialogOpen}>
          <DialogTrigger asChild>
            <Button size="sm" disabled={activeDebugSession?.enabled}>
              <PlayIcon className="mr-2 h-4 w-4" />
              开启调试
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>开启临时调试</DialogTitle>
              <DialogDescription>
                调试模式将记录更详细的请求和响应信息，请谨慎使用
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <Field>
                <FieldLabel>生效范围</FieldLabel>
                <Select defaultValue="global">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="global">全局</SelectItem>
                    <SelectItem value="model">指定模型</SelectItem>
                    <SelectItem value="api_key">指定 API Key</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>有效时长</FieldLabel>
                <Select defaultValue="30m">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="15m">15 分钟</SelectItem>
                    <SelectItem value="30m">30 分钟</SelectItem>
                    <SelectItem value="1h">1 小时</SelectItem>
                    <SelectItem value="2h">2 小时</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>开启原因</FieldLabel>
                <Textarea placeholder="请描述开启调试的原因" rows={2} />
              </Field>
              <div className="p-3 rounded-lg bg-destructive/10 border border-destructive/20">
                <div className="flex items-start gap-2">
                  <AlertTriangleIcon className="h-4 w-4 text-destructive mt-0.5" />
                  <div className="text-sm">
                    <p className="font-medium text-destructive">安全提醒</p>
                    <p className="text-muted-foreground">
                      调试模式会记录更多敏感信息，请确保在必要时使用，并及时关闭。
                      调试日志将在 7 天后自动清理。
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsStartDialogOpen(false)}>
                取消
              </Button>
              <Button>开启调试</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Info Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card className="border-border/50 bg-card/50">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <ShieldAlertIcon className="h-4 w-4 text-yellow-500" />
              安全说明
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            <ul className="space-y-1.5 list-disc list-inside">
              <li>调试模式会记录更详细的请求响应内容</li>
              <li>调试日志仍会对敏感字段进行部分脱敏</li>
              <li>调试会话到期后自动关闭</li>
              <li>所有调试操作都会记录审计日志</li>
            </ul>
          </CardContent>
        </Card>

        <Card className="border-border/50 bg-card/50">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <EyeIcon className="h-4 w-4 text-primary" />
              调试日志内容
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            <ul className="space-y-1.5 list-disc list-inside">
              <li>完整请求参数（部分脱敏）</li>
              <li>上游原始响应内容</li>
              <li>参数转换过程详情</li>
              <li>错误堆栈和上下文信息</li>
            </ul>
          </CardContent>
        </Card>

        <Card className="border-border/50 bg-card/50">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <ClockIcon className="h-4 w-4 text-muted-foreground" />
              日志保留策略
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            <ul className="space-y-1.5 list-disc list-inside">
              <li>普通日志保留 30 天</li>
              <li>调试日志保留 7 天</li>
              <li>审计日志保留 90 天</li>
              <li>可在系统设置中调整</li>
            </ul>
          </CardContent>
        </Card>
      </div>

      {/* Debug History */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base">调试历史</CardTitle>
          <CardDescription>最近的调试会话记录</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {debugHistory.map((session) => (
              <div
                key={session.id}
                className="flex items-center justify-between p-4 rounded-lg bg-secondary/30 border border-border/50"
              >
                <div className="flex items-center gap-4">
                  <div className={`p-2 rounded-lg ${
                    session.status === "active" ? "bg-yellow-500/20" : "bg-secondary"
                  }`}>
                    <BugIcon className={`h-4 w-4 ${
                      session.status === "active" ? "text-yellow-500" : "text-muted-foreground"
                    }`} />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <Badge variant={session.status === "active" ? "default" : "secondary"} className="text-xs">
                        {scopeTypeLabels[session.scopeType]}
                        {session.scopeValue && `: ${session.scopeValue}`}
                      </Badge>
                      {session.status === "active" && (
                        <Badge variant="outline" className="text-yellow-500 border-yellow-500/50 text-xs">
                          进行中
                        </Badge>
                      )}
                    </div>
                    <p className="text-sm mt-1">{session.reason}</p>
                    <p className="text-xs text-muted-foreground mt-1">
                      由 {session.startedBy} 于 {session.startedAt} 开启
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-mono">{session.duration}</p>
                  <p className="text-xs text-muted-foreground">
                    {session.status === "active" ? "剩余时间" : "持续时长"}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
