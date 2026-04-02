'use client'

import { useEffect, useState } from 'react'
import { LuSend, LuLoader, LuCopy, LuCheck } from 'react-icons/lu'
import { FiCode } from 'react-icons/fi'
import { api } from '@/lib/api'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useToast } from '@/components/ui/toast'
import { Textarea } from '@/components/ui/textarea'

const defaultBody = {
  model: '',
  messages: [
    { role: 'user', content: '1' }
  ],
  temperature: 0.7,
  max_tokens: 500,
  stream: false
}

export default function TestPage() {
  const [keys, setKeys] = useState<any[]>([])
  const [models, setModels] = useState<any[]>([])
  const [selectedKey, setSelectedKey] = useState('')
  const [selectedModel, setSelectedModel] = useState('')
  const [bodyText, setBodyText] = useState(JSON.stringify(defaultBody, null, 2))
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)
  const [copied, setCopied] = useState(false)
  const { showSuccess, showError } = useToast()

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    try {
      const [keysData, modelsData] = await Promise.all([
        api.clientKeys('page=1&page_size=100'),
        api.models()
      ])
      setKeys(keysData.items || [])
      setModels(modelsData.items || [])
    } catch (err) {
      console.error(err)
    }
  }

  function handleModelChange(model: string) {
    setSelectedModel(model)
    try {
      const body = JSON.parse(bodyText)
      body.model = model
      setBodyText(JSON.stringify(body, null, 2))
    } catch {}
  }

  async function sendRequest() {
    if (!selectedKey) {
      showError('请选择 API Key')
      return
    }
    if (!selectedModel) {
      showError('请选择模型')
      return
    }

    let body
    try {
      body = JSON.parse(bodyText)
    } catch {
      showError('请求体 JSON 格式错误')
      return
    }

    setLoading(true)
    setResult(null)

    try {
      const data = await api.testRequest({
        api_key: selectedKey,
        model: body.model || selectedModel,
        messages: body.messages || defaultBody.messages,
        stream: body.stream || false,
        temperature: body.temperature,
        max_tokens: body.max_tokens
      })
      setResult(data)
      if (data.key_status) {
        setKeys((prev) => prev.map((key) => key.plain_key === selectedKey ? { ...key, status: data.key_status } : key))
      }
      if (data.success) {
        showSuccess(`请求成功，耗时 ${data.latency_ms}ms`)
      } else {
        showError(data.error || '请求失败')
      }
    } catch (err) {
      showError(err instanceof Error ? err.message : '请求失败')
    } finally {
      setLoading(false)
    }
  }

  function copyResult() {
    if (!result) return
    navigator.clipboard.writeText(JSON.stringify(result, null, 2))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  function fillTemplate() {
      const template = {
        model: selectedModel || '',
        messages: [
        { role: 'user', content: '1' }
        ],
      temperature: 0.7,
      max_tokens: 500,
      stream: false
    }
    setBodyText(JSON.stringify(template, null, 2))
  }

  return (
    <div className="space-y-4">
      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
          <div className="flex items-center gap-2">
            <FiCode className="h-4 w-4" style={{ color: 'var(--muted-foreground)' }} />
            <span className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>API 测试工具</span>
          </div>
        </div>
        <CardContent className="p-4 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm mb-1 block" style={{ color: 'var(--foreground)' }}>API Key</label>
              <Select value={selectedKey} onValueChange={setSelectedKey}>
                <SelectTrigger>
                  <SelectValue placeholder="选择 API Key" />
                </SelectTrigger>
                <SelectContent>
                  {keys.map((key) => (
                    <SelectItem key={key.id} value={key.plain_key || ''}>
                      {key.name} ({key.key_prefix}****) {key.status === 'active' ? '[启用]' : '[禁用]'}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-sm mb-1 block" style={{ color: 'var(--foreground)' }}>模型</label>
              <Select value={selectedModel} onValueChange={handleModelChange}>
                <SelectTrigger>
                  <SelectValue placeholder="选择模型" />
                </SelectTrigger>
                <SelectContent>
                  {models.map((model) => (
                    <SelectItem key={model.id} value={model.model_code}>
                      {model.model_code}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-sm" style={{ color: 'var(--foreground)' }}>请求体</label>
              <Button variant="secondary" size="sm" onClick={fillTemplate}>填充模板</Button>
            </div>
            <Textarea
              className="font-mono text-sm min-h-[180px]"
              value={bodyText}
              onChange={(e) => setBodyText(e.target.value)}
            />
          </div>

          <div className="flex justify-end">
            <Button onClick={sendRequest} loading={loading} disabled={!selectedKey || !selectedModel}>
              <LuSend className="h-4 w-4 mr-1" />
              发送请求
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
          <span className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>响应结果</span>
          {result && (
            <Button variant="secondary" size="sm" onClick={copyResult}>
              {copied ? <LuCheck className="h-3 w-3 mr-1" /> : <LuCopy className="h-3 w-3 mr-1" />}
              {copied ? '已复制' : '复制'}
            </Button>
          )}
        </div>
        <CardContent className="p-4">
          {loading ? (
            <div className="flex items-center justify-center py-12 text-sm" style={{ color: 'var(--muted-foreground)' }}>
              <LuLoader className="h-4 w-4 mr-2 animate-spin" />
              请求中...
            </div>
          ) : result ? (
            <div className="space-y-3">
              <div className="flex gap-4 text-sm">
                <span style={{ color: result.success ? 'var(--success)' : 'var(--danger)' }}>
                  状态: {result.success ? '成功' : '失败'}
                </span>
                {result.latency_ms && (
                  <span style={{ color: 'var(--muted-foreground)' }}>
                    耗时: {result.latency_ms}ms
                  </span>
                )}
                {result.total_tokens !== undefined && (
                  <span style={{ color: 'var(--muted-foreground)' }}>
                    Tokens: {result.total_tokens}
                  </span>
                )}
                {result.prompt_tokens !== undefined && (
                  <span style={{ color: 'var(--muted-foreground)' }}>
                    输入: {result.prompt_tokens}
                  </span>
                )}
                {result.completion_tokens !== undefined && (
                  <span style={{ color: 'var(--muted-foreground)' }}>
                    输出: {result.completion_tokens}
                  </span>
                )}
              </div>
              {result.error && (
                <div className="text-sm p-3 rounded" style={{ background: 'rgba(239,68,68,0.1)', color: 'var(--danger)' }}>
                  {result.error}
                </div>
              )}
              <pre className="overflow-auto max-h-[500px] rounded-xl border p-4 text-sm font-mono shadow-inner" style={{ background: 'color-mix(in srgb, var(--card) 82%, black 18%)', borderColor: 'var(--border)', color: 'var(--foreground)' }}>
                {result.response ? JSON.stringify(result.response, null, 2) : JSON.stringify(result, null, 2)}
              </pre>
            </div>
          ) : (
            <div className="py-12 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>
              发送请求后显示结果
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
