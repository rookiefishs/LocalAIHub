"use client"

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
  ActivityIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  TrendingUpIcon,
  ServerIcon,
  KeyIcon,
  AlertTriangleIcon,
  ArrowRightIcon,
  ZapIcon,
} from "lucide-react"
import { OverviewChart } from "@/components/overview-chart"
import { UpstreamHealthCard } from "@/components/upstream-health-card"

// Mock data
const stats = [
  {
    title: "今日请求数",
    value: "128,543",
    change: "+12.5%",
    trend: "up",
    icon: ActivityIcon,
  },
  {
    title: "成功率",
    value: "99.8%",
    change: "+0.2%",
    trend: "up",
    icon: CheckCircleIcon,
  },
  {
    title: "平均延迟",
    value: "245ms",
    change: "-15ms",
    trend: "up",
    icon: ClockIcon,
  },
  {
    title: "活跃 API Key",
    value: "24",
    change: "+3",
    trend: "up",
    icon: KeyIcon,
  },
]

const recentErrors = [
  {
    id: "err-001",
    time: "2 分钟前",
    model: "gpt-4o",
    upstream: "OpenAI 官方",
    error: "Rate limit exceeded",
    status: "已恢复",
  },
  {
    id: "err-002",
    time: "15 分钟前",
    model: "claude-sonnet",
    upstream: "Anthropic 官方",
    error: "Connection timeout",
    status: "已恢复",
  },
  {
    id: "err-003",
    time: "1 小时前",
    model: "gemini-pro",
    upstream: "Gemini 中转站",
    error: "Authentication failed",
    status: "处理中",
  },
]

const recentSwitches = [
  {
    id: "sw-001",
    time: "5 分钟前",
    model: "gpt-4o",
    from: "中转站 A",
    to: "OpenAI 官方",
    reason: "自动熔断",
    operator: "系统",
  },
  {
    id: "sw-002",
    time: "2 小时前",
    model: "claude-sonnet",
    from: "中转站 B",
    to: "Anthropic 官方",
    reason: "手动切换",
    operator: "admin",
  },
]

const topApiKeys = [
  { name: "生产环境主Key", calls: 45231, percentage: 35 },
  { name: "测试环境Key", calls: 23145, percentage: 18 },
  { name: "内部调试Key", calls: 18432, percentage: 14 },
  { name: "合作方A", calls: 12876, percentage: 10 },
  { name: "合作方B", calls: 8654, percentage: 7 },
]

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <Card key={stat.title} className="border-border/50 bg-card/50">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {stat.title}
              </CardTitle>
              <stat.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <TrendingUpIcon className="h-3 w-3 text-primary" />
                <span className="text-primary">{stat.change}</span> 较昨日
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Charts Row */}
      <div className="grid gap-4 lg:grid-cols-7">
        <Card className="lg:col-span-4 border-border/50 bg-card/50">
          <CardHeader>
            <CardTitle className="text-base">请求趋势</CardTitle>
            <CardDescription>最近 24 小时请求量</CardDescription>
          </CardHeader>
          <CardContent>
            <OverviewChart />
          </CardContent>
        </Card>

        <Card className="lg:col-span-3 border-border/50 bg-card/50">
          <CardHeader>
            <CardTitle className="text-base">上游健康状态</CardTitle>
            <CardDescription>各上游服务可用性</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <UpstreamHealthCard
              name="OpenAI 官方"
              status="healthy"
              latency={120}
              uptime={99.9}
            />
            <UpstreamHealthCard
              name="Anthropic 官方"
              status="healthy"
              latency={156}
              uptime={99.8}
            />
            <UpstreamHealthCard
              name="Gemini 官方"
              status="degraded"
              latency={340}
              uptime={98.5}
            />
            <UpstreamHealthCard
              name="中转站 A"
              status="healthy"
              latency={89}
              uptime={99.7}
            />
            <UpstreamHealthCard
              name="中转站 B"
              status="down"
              latency={0}
              uptime={0}
            />
          </CardContent>
        </Card>
      </div>

      {/* Tables Row */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* Recent Errors */}
        <Card className="border-border/50 bg-card/50">
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle className="text-base flex items-center gap-2">
                <AlertTriangleIcon className="h-4 w-4 text-warning" />
                最近异常事件
              </CardTitle>
              <CardDescription>最近发生的错误和异常</CardDescription>
            </div>
            <Badge variant="outline" className="text-xs">
              3 条记录
            </Badge>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>时间</TableHead>
                  <TableHead>模型</TableHead>
                  <TableHead>上游</TableHead>
                  <TableHead>状态</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {recentErrors.map((error) => (
                  <TableRow key={error.id}>
                    <TableCell className="text-muted-foreground text-xs">
                      {error.time}
                    </TableCell>
                    <TableCell className="font-medium">{error.model}</TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {error.upstream}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={error.status === "已恢复" ? "secondary" : "destructive"}
                        className="text-xs"
                      >
                        {error.status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Recent Switches */}
        <Card className="border-border/50 bg-card/50">
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle className="text-base flex items-center gap-2">
                <ZapIcon className="h-4 w-4 text-primary" />
                最近路由切换
              </CardTitle>
              <CardDescription>路由变更记录</CardDescription>
            </div>
            <Badge variant="outline" className="text-xs">
              2 条记录
            </Badge>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>时间</TableHead>
                  <TableHead>模型</TableHead>
                  <TableHead>切换路径</TableHead>
                  <TableHead>触发</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {recentSwitches.map((sw) => (
                  <TableRow key={sw.id}>
                    <TableCell className="text-muted-foreground text-xs">
                      {sw.time}
                    </TableCell>
                    <TableCell className="font-medium">{sw.model}</TableCell>
                    <TableCell className="text-sm">
                      <span className="text-muted-foreground">{sw.from}</span>
                      <ArrowRightIcon className="inline h-3 w-3 mx-1 text-muted-foreground" />
                      <span className="text-primary">{sw.to}</span>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={sw.reason === "自动熔断" ? "destructive" : "outline"}
                        className="text-xs"
                      >
                        {sw.reason}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>

      {/* API Key Usage */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base">API Key 调用 Top 5</CardTitle>
          <CardDescription>今日调用量排行</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {topApiKeys.map((key, index) => (
              <div key={key.name} className="flex items-center gap-4">
                <div className="w-6 text-center text-sm text-muted-foreground">
                  {index + 1}
                </div>
                <div className="flex-1">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm font-medium">{key.name}</span>
                    <span className="text-sm text-muted-foreground">
                      {key.calls.toLocaleString()} 次
                    </span>
                  </div>
                  <div className="h-2 rounded-full bg-secondary overflow-hidden">
                    <div
                      className="h-full bg-primary rounded-full transition-all"
                      style={{ width: `${key.percentage}%` }}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
