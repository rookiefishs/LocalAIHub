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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Input } from "@/components/ui/input"
import { Field, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import {
  PlusIcon,
  MoreHorizontalIcon,
  KeyIcon,
  CopyIcon,
  EditIcon,
  TrashIcon,
  CheckIcon,
  EyeIcon,
  EyeOffIcon,
  CircleIcon,
  ActivityIcon,
} from "lucide-react"

// Mock data
const apiKeys = [
  {
    id: "key-001",
    name: "生产环境主Key",
    keyPrefix: "sk-prod-",
    keySuffix: "...a3f7",
    status: "active",
    createdAt: "2024-01-15",
    lastUsedAt: "刚刚",
    callsToday: 45231,
    callsTotal: 1234567,
    remark: "主要生产环境使用",
  },
  {
    id: "key-002",
    name: "测试环境Key",
    keyPrefix: "sk-test-",
    keySuffix: "...b2e8",
    status: "active",
    createdAt: "2024-02-01",
    lastUsedAt: "5 分钟前",
    callsToday: 23145,
    callsTotal: 456789,
    remark: "测试环境专用",
  },
  {
    id: "key-003",
    name: "内部调试Key",
    keyPrefix: "sk-debug-",
    keySuffix: "...c1d9",
    status: "active",
    createdAt: "2024-02-15",
    lastUsedAt: "1 小时前",
    callsToday: 18432,
    callsTotal: 234567,
    remark: "内部开发调试",
  },
  {
    id: "key-004",
    name: "合作方A",
    keyPrefix: "sk-partner-",
    keySuffix: "...d4e2",
    status: "active",
    createdAt: "2024-03-01",
    lastUsedAt: "3 小时前",
    callsToday: 12876,
    callsTotal: 98765,
    remark: "合作方 A 公司接入",
  },
  {
    id: "key-005",
    name: "合作方B",
    keyPrefix: "sk-partner-",
    keySuffix: "...e5f3",
    status: "disabled",
    createdAt: "2024-03-10",
    lastUsedAt: "3 天前",
    callsToday: 0,
    callsTotal: 45678,
    remark: "合作方 B - 已暂停合作",
  },
  {
    id: "key-006",
    name: "临时测试Key",
    keyPrefix: "sk-temp-",
    keySuffix: "...f6g4",
    status: "expired",
    createdAt: "2024-03-20",
    lastUsedAt: "7 天前",
    callsToday: 0,
    callsTotal: 1234,
    remark: "临时测试使用，已过期",
  },
]

const statusConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  active: { label: "启用", color: "text-green-500", bgColor: "bg-green-500/10" },
  disabled: { label: "禁用", color: "text-red-500", bgColor: "bg-red-500/10" },
  expired: { label: "已过期", color: "text-gray-500", bgColor: "bg-gray-500/10" },
}

export default function KeysPage() {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
  const [isKeyCreatedDialogOpen, setIsKeyCreatedDialogOpen] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [newKey, setNewKey] = useState("")
  const [copied, setCopied] = useState(false)

  const handleCreateKey = () => {
    // Simulate key creation
    setNewKey("sk-prod-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz")
    setIsAddDialogOpen(false)
    setIsKeyCreatedDialogOpen(true)
  }

  const handleCopyKey = async () => {
    await navigator.clipboard.writeText(newKey)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">API Key 管理</h2>
          <p className="text-sm text-muted-foreground">
            管理客户端接入使用的 API Key，支持创建、禁用和监控
          </p>
        </div>
        <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <PlusIcon className="mr-2 h-4 w-4" />
              创建 Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>创建 API Key</DialogTitle>
              <DialogDescription>
                创建一个新的 API Key 供客户端接入使用
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <Field>
                <FieldLabel>名称</FieldLabel>
                <Input placeholder="例如：生产环境主Key" />
              </Field>
              <Field>
                <FieldLabel>备注</FieldLabel>
                <Textarea placeholder="可选，添加备注信息便于识别" rows={2} />
              </Field>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">设置过期时间</p>
                  <p className="text-xs text-muted-foreground">可选，留空则永不过期</p>
                </div>
                <Switch />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
                取消
              </Button>
              <Button onClick={handleCreateKey}>创建</Button>
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
                <KeyIcon className="h-4 w-4 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold">6</p>
                <p className="text-xs text-muted-foreground">总 Key 数</p>
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
                <p className="text-2xl font-bold">4</p>
                <p className="text-xs text-muted-foreground">启用中</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue-500/10">
                <ActivityIcon className="h-4 w-4 text-blue-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">99,684</p>
                <p className="text-xs text-muted-foreground">今日调用</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-border/50 bg-card/50">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-primary/10">
                <ActivityIcon className="h-4 w-4 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold">2.07M</p>
                <p className="text-xs text-muted-foreground">总调用量</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Keys Table */}
      <Card className="border-border/50 bg-card/50">
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>状态</TableHead>
                <TableHead>名称</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>今日调用</TableHead>
                <TableHead>总调用</TableHead>
                <TableHead>最后使用</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="w-12"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {apiKeys.map((key) => {
                const status = statusConfig[key.status]
                return (
                  <TableRow key={key.id}>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={`${status.bgColor} ${status.color} border-0 text-xs`}
                      >
                        {status.label}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div>
                        <p className="font-medium">{key.name}</p>
                        <p className="text-xs text-muted-foreground truncate max-w-[150px]">
                          {key.remark}
                        </p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <code className="text-sm font-mono text-muted-foreground">
                        {key.keyPrefix}{key.keySuffix}
                      </code>
                    </TableCell>
                    <TableCell>
                      <span className="font-mono">{key.callsToday.toLocaleString()}</span>
                    </TableCell>
                    <TableCell>
                      <span className="font-mono text-muted-foreground">
                        {key.callsTotal.toLocaleString()}
                      </span>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {key.lastUsedAt}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {key.createdAt}
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
                            <ActivityIcon className="mr-2 h-4 w-4" />
                            查看调用详情
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem>
                            {key.status === "active" ? (
                              <>
                                <EyeOffIcon className="mr-2 h-4 w-4" />
                                禁用
                              </>
                            ) : (
                              <>
                                <EyeIcon className="mr-2 h-4 w-4" />
                                启用
                              </>
                            )}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => setIsDeleteDialogOpen(true)}
                          >
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

      {/* Key Created Dialog */}
      <Dialog open={isKeyCreatedDialogOpen} onOpenChange={setIsKeyCreatedDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>API Key 创建成功</DialogTitle>
            <DialogDescription>
              请立即复制保存您的 API Key，关闭后将无法再次查看完整 Key
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <div className="p-4 rounded-lg bg-secondary/50 border border-border">
              <div className="flex items-center justify-between gap-4">
                <code className="text-sm font-mono break-all flex-1">{newKey}</code>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={handleCopyKey}
                  className="shrink-0"
                >
                  {copied ? (
                    <CheckIcon className="h-4 w-4 text-green-500" />
                  ) : (
                    <CopyIcon className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <p className="text-sm text-destructive mt-4">
              警告：这是您唯一一次看到完整 Key 的机会，请妥善保存！
            </p>
          </div>
          <DialogFooter>
            <Button onClick={() => setIsKeyCreatedDialogOpen(false)}>
              我已保存，关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              删除后该 API Key 将立即失效，所有使用该 Key 的客户端将无法继续访问。此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
