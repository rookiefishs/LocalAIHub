'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { FiServer, FiKey, FiBox, FiActivity, FiArrowRight, FiCheckCircle } from 'react-icons/fi'
import Link from 'next/link'

const steps = [
  {
    icon: FiServer,
    title: '添加上游',
    desc: '在「上游管理」中添加 AI 提供商配置，设置 Base URL 和 API Key',
    href: '/dashboard/upstreams',
    color: 'text-blue-400'
  },
  {
    icon: FiBox,
    title: '创建模型',
    desc: '在「虚拟模型」中创建要暴露给客户端的模型定义',
    href: '/dashboard/models',
    color: 'text-purple-400'
  },
  {
    icon: FiArrowRight,
    title: '绑定上游',
    desc: '为模型绑定上游提供商、模型名称和优先级',
    href: '/dashboard/routes',
    color: 'text-yellow-400'
  },
  {
    icon: FiKey,
    title: '分发 Key',
    desc: '在「API Key」中创建客户端密钥，设置有效期和模型权限',
    href: '/dashboard/keys',
    color: 'text-emerald-400'
  }
]

const examples = [
  {
    label: 'cURL',
    code: `curl -X POST your-domain/proxy/openai/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "你好"}]
  }'`
  },
  {
    label: 'Python',
    code: `import requests

response = requests.post(
    "your-domain/proxy/openai/v1/chat/completions",
    headers={
        "Content-Type": "application/json",
        "Authorization": "Bearer YOUR_API_KEY"
    },
    json={
        "model": "gpt-4o-mini",
        "messages": [{"role": "user", "content": "你好"}]
    }
)
print(response.json())`
  },
  {
    label: 'JavaScript',
    code: `fetch("your-domain/proxy/openai/v1/chat/completions", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_API_KEY"
  },
  body: JSON.stringify({
    model: "gpt-4o-mini",
    messages: [{role: "user", content: "你好"}]
  })
}).then(r => r.json()).then(console.log)`
  }
]

export default function HelpPage() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {steps.map((step, i) => (
          <Link key={i} href={step.href}>
            <Card className="h-full cursor-pointer hover:border-primary/50 transition-colors">
              <CardContent className="p-5">
                <div className="flex items-start justify-between">
                  <div className={`p-2 rounded-[10px] bg-primary/10 ${step.color}`}>
                    <step.icon className="h-5 w-5" />
                  </div>
                  <span className="text-xs font-mono" style={{ color: 'var(--muted-foreground)' }}>{String(i + 1).padStart(2, '0')}</span>
                </div>
                <div className="mt-4 font-medium">{step.title}</div>
                <div className="mt-1 text-xs" style={{ color: 'var(--muted-foreground)' }}>{step.desc}</div>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">调用示例</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="text-sm font-medium mb-2">cURL</div>
              <pre className="p-3 rounded-[10px] border overflow-x-auto text-xs font-mono leading-relaxed" style={{ background: 'rgba(0,0,0,0.15)', borderColor: 'var(--border)' }}>
{`curl -X POST your-domain/proxy/openai/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "你好"}]}'`}
              </pre>
            </div>
            <div>
              <div className="text-sm font-medium mb-2">Python</div>
              <pre className="p-3 rounded-[10px] border overflow-x-auto text-xs font-mono leading-relaxed" style={{ background: 'rgba(0,0,0,0.15)', borderColor: 'var(--border)' }}>
{`import requests
r = requests.post("your-domain/proxy/openai/v1/chat/completions",
  headers={"Authorization": "Bearer YOUR_API_KEY"},
  json={"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "你好"}]})
print(r.json())`}
              </pre>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">注意事项</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {[
              'API Key 首次创建时会自动检测连通性，失败会自动禁用',
              '模型优先级数字越小越优先，当前一个失败会自动切换',
              '可在「日志中心」查看详细请求和调用日志',
              '建议为不同用途创建独立的 API Key 便于管理'
            ].map((item, i) => (
              <div key={i} className="flex items-start gap-2">
                <FiCheckCircle className="h-4 w-4 text-emerald-400 mt-0.5 flex-shrink-0" />
                <span className="text-sm" style={{ color: 'var(--foreground)' }}>{item}</span>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base flex items-center gap-2">
            <FiActivity className="h-4 w-4" />
            快速跳转
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          <Link href="/dashboard/upstreams"><Button variant="secondary" size="sm">上游管理</Button></Link>
          <Link href="/dashboard/models"><Button variant="secondary" size="sm">虚拟模型</Button></Link>
          <Link href="/dashboard/routes"><Button variant="secondary" size="sm">路由管理</Button></Link>
          <Link href="/dashboard/keys"><Button variant="secondary" size="sm">API Key</Button></Link>
          <Link href="/dashboard/logs"><Button variant="secondary" size="sm">日志中心</Button></Link>
        </CardContent>
      </Card>
    </div>
  )
}
