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
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Field, FieldLabel } from "@/components/ui/field"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import {
  PlusIcon,
  MoreHorizontalIcon,
  RefreshCwIcon,
  CircleIcon,
  EditIcon,
  TrashIcon,
  ZapIcon,
  ServerIcon,
  KeyIcon,
} from "lucide-react"

// Mock data
const upstreams = [
  {
    id: "up-001",
    name: "OpenAI 官方",
    type: "openai",
    providerType: "official",
    baseUrl: "https://api.openai.com/v1",
    keyMode: "single",
    keyCount: 1,
    timeout: 30000,
    retryCount: 3,
    enabled: true,
    healthStatus: "healthy",
    lastCheckAt: "2 分钟前",
    linkedModels: 5,
  },
  {
    id: "up-002",
    name: "Anthropic 官方",
    type: "anthropic",
    providerType: "official",
    baseUrl: "https://api.anthropic.com/v1",
    keyMode: "single",
    keyCount: 1,
    timeout: 30000,
    retryCount: 3,
    enabled: true,
    healthStatus: "healthy",
    lastCheckAt: "1 分钟前",
    linkedModels: 3,
  },
  {
    id: "up-003",
    name: "Gemini 官方",
    type: "gemini",
    providerType: "official",
    baseUrl: "https://generativelanguage.googleapis.com/v1",
    keyMode: "single",
    keyCount: 1,
    timeout: 30000,
    retryCount: 3,
    enabled: true,
    healthStatus: "degraded",
    lastCheckAt: "刚刚",
    linkedModels: 2,
  },
  {
    id: "up-004",
    name: "中转站 A",
    type: "openai",
    providerType: "proxy",
    baseUrl: "https://proxy-a.example.com/v1",
    keyMode: "multi",
    keyCount: 5,
    timeout: 20000,
    retryCount: 2,
    enabled: true,
    healthStatus: "healthy",
    lastCheckAt: "3 分钟前",
    linkedModels: 8,
  },
  {
    id: "up-005",
    name: "中转站 B",
    type: "openai",
    providerType: "proxy",
    baseUrl: "https://proxy-b.example.com/v1",
    keyMode: "multi",
    keyCount: 3,
    timeout: 20000,
    retryCount: 2,
    enabled: false,
    healthStatus: "down",
    lastCheckAt: "10 分钟前",
    linkedModels: 4,
  },
]

const statusConfig: Record<string, { label: string; color: string }> = {
  healthy: { label: "正常", color: "text-green-500" },
  degraded: { label: "降级", color: "text-yellow-500" },
  down: { label: "离线", color: "text-red-500" },
}

const typeLabels: Record<string, string> = {
  openai: "OpenAI",
  anthropic: "Anthropic",
  gemini: "Gemini",
}

export default function UpstreamsPage() {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">上游渠道列表</h2>
          <p className="text-sm text-muted-foreground">
            管理所有上游 API 服务商和中转站配置
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm">
            <RefreshCwIcon className="mr-2 h-4 w-4" />
            刷新状态
          </Button>
          <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
            <DialogTrigger asChild>
              <Button size="sm">
                <PlusIcon className="mr-2 h-4 w-4" />
                新增上游
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>新增上游渠道</DialogTitle>
                <DialogDescription>
                  添加新的上游 API 服务商或中转站
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>上游名称</FieldLabel>
                    <Input placeholder="例如：OpenAI 官方" />
                  </Field>
                  <Field>
                    <FieldLabel>协议类型</FieldLabel>
                    <Select defaultValue="openai">
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="openai">OpenAI</SelectItem>
                        <SelectItem value="anthropic">Anthropic</SelectItem>
                        <SelectItem value="gemini">Gemini</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>服务商类型</FieldLabel>
                    <Select defaultValue="official">
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="official">官方服务</SelectItem>
                        <SelectItem value="proxy">第三方中转</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                  <Field>
                    <FieldLabel>Key 模式</FieldLabel>
                    <Select defaultValue="single">
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="single">单 Key</SelectItem>
                        <SelectItem value="multi">多 Key</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                </div>
                <Field>
                  <FieldLabel>Base URL</FieldLabel>
                  <Input placeholder="https://api.openai.com/v1" />
                </Field>
                <Field>
                  <FieldLabel>API Key / Token</FieldLabel>
                  <Input type="password" placeholder="sk-..." />
                </Field>
                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>超时时间 (ms)</FieldLabel>
                    <Input type="number" defaultValue="30000" />
                  </Field>
                  <Field>
                    <FieldLabel>重试次数</FieldLabel>
                    <Input type="number" defaultValue="3" />
                  </Field>
                </div>
                <Field>
                  <FieldLabel>备注</FieldLabel>
                  <Textarea placeholder="可选，添加备注信息" rows={2} />
                </Field>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">启用健康检查</p>
                    <p className="text-xs text-muted-foreground">定期检测上游可用性</p>
                  </div>
                  <Switch defaultChecked />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
                  取消
                </Button>
                <Button variant="outline">
                  <ZapIcon className="mr-2 h-4 w-4" />
                  测试连接
                </Button>
                <Button>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-primary/10">
                <ServerIcon className="h-4 w-4 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold">5</p>
                <p className="text-xs text-muted-foreground">总上游数</p>
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
                <p className="text-2xl font-bold">3</p>
                <p className="text-xs text-muted-foreground">正常运行</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-yellow-500/10">
                <CircleIcon className="h-4 w-4 text-yellow-500" fill="currentColor" />
              </div>
              <div>
                <p className="text-2xl font-bold">1</p>
                <p className="text-xs text-muted-foreground">性能降级</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-red-500/10">
                <CircleIcon className="h-4 w-4 text-red-500" fill="currentColor" />
              </div>
              <div>
                <p className="text-2xl font-bold">1</p>
                <p className="text-xs text-muted-foreground">离线</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Upstreams Table */}
      <Card className="border-border/50 bg-card/50">
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>状态</TableHead>
                <TableHead>名称</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>服务商</TableHead>
                <TableHead>Key 模式</TableHead>
                <TableHead>关联模型</TableHead>
                <TableHead>最后检测</TableHead>
                <TableHead>启用</TableHead>
                <TableHead className="w-12"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {upstreams.map((upstream) => {
                const status = statusConfig[upstream.healthStatus]
                return (
                  <TableRow key={upstream.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <CircleIcon
                          className={`h-2.5 w-2.5 ${status.color}`}
                          fill="currentColor"
                        />
                        <span className="text-sm">{status.label}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div>
                        <p className="font-medium">{upstream.name}</p>
                        <p className="text-xs text-muted-foreground truncate max-w-[200px]">
                          {upstream.baseUrl}
                        </p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {typeLabels[upstream.type]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={upstream.providerType === "official" ? "default" : "secondary"}
                        className="text-xs"
                      >
                        {upstream.providerType === "official" ? "官方" : "中转"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <KeyIcon className="h-3 w-3 text-muted-foreground" />
                        <span className="text-sm">
                          {upstream.keyMode === "single" ? "单 Key" : `${upstream.keyCount} Key`}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm">{upstream.linkedModels} 个</span>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {upstream.lastCheckAt}
                      </span>
                    </TableCell>
                    <TableCell>
                      <Switch checked={upstream.enabled} />
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <MoreHorizontalIcon className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>
                            <EditIcon className="mr-2 h-4 w-4" />
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <ZapIcon className="mr-2 h-4 w-4" />
                            测试连接
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <KeyIcon className="mr-2 h-4 w-4" />
                            管理 Key
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem className="text-destructive">
                            <TrashIcon className="mr-2 h-4 w-4" />
                            删除
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
    </div>
  )
}
