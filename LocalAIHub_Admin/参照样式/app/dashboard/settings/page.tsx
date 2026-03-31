"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
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
  SettingsIcon,
  ShieldIcon,
  ClockIcon,
  DatabaseIcon,
  AlertTriangleIcon,
  SaveIcon,
  KeyIcon,
} from "lucide-react"

export default function SettingsPage() {
  const [isPasswordDialogOpen, setIsPasswordDialogOpen] = useState(false)

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Header */}
      <div>
        <h2 className="text-lg font-semibold">系统设置</h2>
        <p className="text-sm text-muted-foreground">
          配置系统参数、熔断阈值和安全选项
        </p>
      </div>

      {/* Account Settings */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <ShieldIcon className="h-4 w-4" />
            账号安全
          </CardTitle>
          <CardDescription>
            管理员账号和密码设置
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <FieldGroup>
            <Field>
              <FieldLabel>当前账号</FieldLabel>
              <Input value="admin" disabled className="bg-secondary/50" />
            </Field>
          </FieldGroup>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">修改密码</p>
              <p className="text-xs text-muted-foreground">建议定期更换密码以保证账号安全</p>
            </div>
            <Dialog open={isPasswordDialogOpen} onOpenChange={setIsPasswordDialogOpen}>
              <DialogTrigger asChild>
                <Button variant="outline" size="sm">
                  <KeyIcon className="mr-2 h-4 w-4" />
                  修改密码
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>修改密码</DialogTitle>
                  <DialogDescription>
                    请输入当前密码和新密码
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <Field>
                    <FieldLabel>当前密码</FieldLabel>
                    <Input type="password" placeholder="请输入当前密码" />
                  </Field>
                  <Field>
                    <FieldLabel>新密码</FieldLabel>
                    <Input type="password" placeholder="请输入新密码" />
                  </Field>
                  <Field>
                    <FieldLabel>确认新密码</FieldLabel>
                    <Input type="password" placeholder="请再次输入新密码" />
                  </Field>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsPasswordDialogOpen(false)}>
                    取消
                  </Button>
                  <Button>确认修改</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
        </CardContent>
      </Card>

      {/* Breaker Settings */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <AlertTriangleIcon className="h-4 w-4" />
            熔断阈值
          </CardTitle>
          <CardDescription>
            配置自动熔断的触发条件
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-4 md:grid-cols-2">
            <Field>
              <FieldLabel>连续失败次数阈值</FieldLabel>
              <Input type="number" defaultValue="5" />
              <p className="text-xs text-muted-foreground mt-1">
                连续失败达到该次数时触发熔断
              </p>
            </Field>
            <Field>
              <FieldLabel>时间窗口失败率阈值 (%)</FieldLabel>
              <Input type="number" defaultValue="50" />
              <p className="text-xs text-muted-foreground mt-1">
                在时间窗口内失败率达到该比例时触发熔断
              </p>
            </Field>
            <Field>
              <FieldLabel>时间窗口大小 (秒)</FieldLabel>
              <Input type="number" defaultValue="60" />
              <p className="text-xs text-muted-foreground mt-1">
                统计失败率的时间窗口
              </p>
            </Field>
            <Field>
              <FieldLabel>超时次数阈值</FieldLabel>
              <Input type="number" defaultValue="3" />
              <p className="text-xs text-muted-foreground mt-1">
                连续超时达到该次数时触发熔断
              </p>
            </Field>
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">自动恢复</p>
              <p className="text-xs text-muted-foreground">
                熔断后自动尝试恢复健康的上游
              </p>
            </div>
            <Switch defaultChecked />
          </div>
          <Field>
            <FieldLabel>自动恢复探测间隔 (秒)</FieldLabel>
            <Input type="number" defaultValue="30" className="max-w-[200px]" />
          </Field>
        </CardContent>
      </Card>

      {/* Timeout Settings */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <ClockIcon className="h-4 w-4" />
            超时配置
          </CardTitle>
          <CardDescription>
            配置请求超时和重试策略
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Field>
              <FieldLabel>默认请求超时 (ms)</FieldLabel>
              <Input type="number" defaultValue="30000" />
            </Field>
            <Field>
              <FieldLabel>默认重试次数</FieldLabel>
              <Input type="number" defaultValue="3" />
            </Field>
            <Field>
              <FieldLabel>连接超时 (ms)</FieldLabel>
              <Input type="number" defaultValue="5000" />
            </Field>
            <Field>
              <FieldLabel>重试间隔 (ms)</FieldLabel>
              <Input type="number" defaultValue="1000" />
            </Field>
          </div>
        </CardContent>
      </Card>

      {/* Log Settings */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <DatabaseIcon className="h-4 w-4" />
            日志保留策略
          </CardTitle>
          <CardDescription>
            配置各类日志的保留时间
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            <Field>
              <FieldLabel>请求日志保留 (天)</FieldLabel>
              <Input type="number" defaultValue="30" />
            </Field>
            <Field>
              <FieldLabel>调试日志保留 (天)</FieldLabel>
              <Input type="number" defaultValue="7" />
            </Field>
            <Field>
              <FieldLabel>审计日志保留 (天)</FieldLabel>
              <Input type="number" defaultValue="90" />
            </Field>
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">日志自动清理</p>
              <p className="text-xs text-muted-foreground">
                自动清理过期日志以释放存储空间
              </p>
            </div>
            <Switch defaultChecked />
          </div>
        </CardContent>
      </Card>

      {/* System Info */}
      <Card className="border-border/50 bg-card/50">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <SettingsIcon className="h-4 w-4" />
            系统信息
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 text-sm">
            <div>
              <p className="text-muted-foreground">系统版本</p>
              <p className="font-mono">v1.0.0</p>
            </div>
            <div>
              <p className="text-muted-foreground">运行时间</p>
              <p className="font-mono">3 天 12 小时 45 分钟</p>
            </div>
            <div>
              <p className="text-muted-foreground">数据库状态</p>
              <p className="text-green-500">正常</p>
            </div>
            <div>
              <p className="text-muted-foreground">存储使用</p>
              <p className="font-mono">2.3 GB / 10 GB</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Save Button */}
      <div className="flex justify-end">
        <Button>
          <SaveIcon className="mr-2 h-4 w-4" />
          保存设置
        </Button>
      </div>
    </div>
  )
}
