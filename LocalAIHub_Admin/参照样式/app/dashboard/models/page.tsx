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
  BoxIcon,
  EditIcon,
  TrashIcon,
  EyeIcon,
  EyeOffIcon,
  GitBranchIcon,
  CircleIcon,
} from "lucide-react"

// Mock data
const virtualModels = [
  {
    id: "vm-001",
    virtualName: "gpt-4o",
    displayName: "GPT-4 Omni",
    description: "最新的 GPT-4 多模态模型",
    primaryProtocol: "openai",
    visible: true,
    status: "active",
    candidates: [
      { upstream: "OpenAI 官方", model: "gpt-4o", priority: 1, status: "active" },
      { upstream: "中转站 A", model: "gpt-4o", priority: 2, status: "active" },
    ],
    tags: ["文本生成", "多模态"],
  },
  {
    id: "vm-002",
    virtualName: "gpt-4-turbo",
    displayName: "GPT-4 Turbo",
    description: "GPT-4 Turbo 128K 上下文",
    primaryProtocol: "openai",
    visible: true,
    status: "active",
    candidates: [
      { upstream: "OpenAI 官方", model: "gpt-4-turbo-preview", priority: 1, status: "active" },
      { upstream: "中转站 A", model: "gpt-4-turbo", priority: 2, status: "active" },
    ],
    tags: ["文本生成", "长上下文"],
  },
  {
    id: "vm-003",
    virtualName: "claude-sonnet",
    displayName: "Claude 3.5 Sonnet",
    description: "Anthropic 最新的 Claude 模型",
    primaryProtocol: "anthropic",
    visible: true,
    status: "active",
    candidates: [
      { upstream: "Anthropic 官方", model: "claude-3-5-sonnet-20241022", priority: 1, status: "active" },
      { upstream: "中转站 B", model: "claude-3-5-sonnet", priority: 2, status: "breaker" },
    ],
    tags: ["文本生成", "代码"],
  },
  {
    id: "vm-004",
    virtualName: "gemini-pro",
    displayName: "Gemini Pro",
    description: "Google Gemini Pro 模型",
    primaryProtocol: "gemini",
    visible: true,
    status: "degraded",
    candidates: [
      { upstream: "Gemini 官方", model: "gemini-pro", priority: 1, status: "degraded" },
    ],
    tags: ["文本生成"],
  },
  {
    id: "vm-005",
    virtualName: "gpt-3.5-turbo",
    displayName: "GPT-3.5 Turbo",
    description: "经济实惠的 GPT-3.5 模型",
    primaryProtocol: "openai",
    visible: false,
    status: "active",
    candidates: [
      { upstream: "中转站 A", model: "gpt-3.5-turbo", priority: 1, status: "active" },
      { upstream: "中转站 B", model: "gpt-3.5-turbo", priority: 2, status: "down" },
    ],
    tags: ["文本生成", "经济"],
  },
]

const statusConfig: Record<string, { label: string; color: string }> = {
  active: { label: "正常", color: "text-green-500" },
  degraded: { label: "降级", color: "text-yellow-500" },
  down: { label: "不可用", color: "text-red-500" },
  breaker: { label: "已熔断", color: "text-orange-500" },
}

const protocolLabels: Record<string, string> = {
  openai: "OpenAI",
  anthropic: "Anthropic",
  gemini: "Gemini",
}

export default function ModelsPage() {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">虚拟模型列表</h2>
          <p className="text-sm text-muted-foreground">
            管理对外提供的统一虚拟模型名，屏蔽上游真实模型差异
          </p>
        </div>
        <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <PlusIcon className="mr-2 h-4 w-4" />
              新增虚拟模型
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>新增虚拟模型</DialogTitle>
              <DialogDescription>
                创建一个新的虚拟模型名，并配置候选上游模型
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <Field>
                  <FieldLabel>虚拟模型名</FieldLabel>
                  <Input placeholder="例如：gpt-4o" />
                </Field>
                <Field>
                  <FieldLabel>展示名称</FieldLabel>
                  <Input placeholder="例如：GPT-4 Omni" />
                </Field>
              </div>
              <Field>
                <FieldLabel>模型说明</FieldLabel>
                <Textarea placeholder="描述模型的能力和用途" rows={2} />
              </Field>
              <div className="grid grid-cols-2 gap-4">
                <Field>
                  <FieldLabel>主协议</FieldLabel>
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
                <Field>
                  <FieldLabel>能力标签</FieldLabel>
                  <Input placeholder="文本生成, 代码, 多模态" />
                </Field>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">对外可见</p>
                  <p className="text-xs text-muted-foreground">在模型列表接口中返回</p>
                </div>
                <Switch defaultChecked />
              </div>
              <div className="border-t pt-4">
                <p className="text-sm font-medium mb-3">候选上游模型</p>
                <p className="text-xs text-muted-foreground mb-4">
                  保存后可在模型详情中配置候选上游和优先级
                </p>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
                取消
              </Button>
              <Button>保存</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-primary/10">
                <BoxIcon className="h-4 w-4 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold">5</p>
                <p className="text-xs text-muted-foreground">总模型数</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green-500/10">
                <EyeIcon className="h-4 w-4 text-green-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">4</p>
                <p className="text-xs text-muted-foreground">对外可见</p>
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
                <p className="text-xs text-muted-foreground">状态正常</p>
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
      </div>

      {/* Models Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {virtualModels.map((model) => {
          const status = statusConfig[model.status]
          return (
            <Card key={model.id} className="border-border/50 bg-card/50">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <CircleIcon
                      className={`h-2.5 w-2.5 ${status.color}`}
                      fill="currentColor"
                    />
                    <CardTitle className="text-base font-mono">
                      {model.virtualName}
                    </CardTitle>
                  </div>
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
                        <GitBranchIcon className="mr-2 h-4 w-4" />
                        管理候选
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        {model.visible ? (
                          <>
                            <EyeOffIcon className="mr-2 h-4 w-4" />
                            隐藏模型
                          </>
                        ) : (
                          <>
                            <EyeIcon className="mr-2 h-4 w-4" />
                            显示模型
                          </>
                        )}
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-destructive">
                        <TrashIcon className="mr-2 h-4 w-4" />
                        删除
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
                <CardDescription className="line-clamp-1">
                  {model.displayName}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground line-clamp-2">
                  {model.description}
                </p>
                <div className="flex items-center gap-2 flex-wrap">
                  <Badge variant="outline" className="text-xs">
                    {protocolLabels[model.primaryProtocol]}
                  </Badge>
                  {!model.visible && (
                    <Badge variant="secondary" className="text-xs">
                      已隐藏
                    </Badge>
                  )}
                  {model.tags.map((tag) => (
                    <Badge key={tag} variant="secondary" className="text-xs">
                      {tag}
                    </Badge>
                  ))}
                </div>
                <div className="pt-2 border-t border-border/50">
                  <p className="text-xs text-muted-foreground mb-2">候选上游</p>
                  <div className="space-y-1.5">
                    {model.candidates.map((candidate, idx) => {
                      const candidateStatus = statusConfig[candidate.status]
                      return (
                        <div
                          key={idx}
                          className="flex items-center justify-between text-sm"
                        >
                          <div className="flex items-center gap-2">
                            <span className="text-xs text-muted-foreground w-4">
                              #{candidate.priority}
                            </span>
                            <span>{candidate.upstream}</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <CircleIcon
                              className={`h-2 w-2 ${candidateStatus.color}`}
                              fill="currentColor"
                            />
                            <span className="text-xs text-muted-foreground">
                              {candidateStatus.label}
                            </span>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}
