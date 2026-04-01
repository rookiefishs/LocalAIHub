'use client'

import { useEffect, useState } from 'react'
import { FiCheck, FiChevronLeft, FiChevronRight, FiPlus } from 'react-icons/fi'
import { GoDotFill } from 'react-icons/go'
import { RiAddLine } from 'react-icons/ri'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { useToast } from '@/components/ui/toast'

interface Step {
  title: string
  description: string
}

const steps: Step[] = [
  { title: '选择上游', description: '选择或添加 AI 服务商' },
  { title: '创建模型', description: '定义虚拟模型' },
  { title: '添加绑定', description: '绑定上游模型' },
  { title: '创建密钥', description: '生成客户端 API Key' },
  { title: '完成', description: '获取调用方式' },
]

interface Provider {
  id: number
  name: string
  base_url: string
  health_status: string
  enabled: boolean
  auth_type: string
}

interface Model {
  id: number
  model_code: string
  display_name: string
  status: string
}

interface ModelBinding {
  id: number
  provider_id: number
  upstream_model_name: string
}

interface WizardState {
  selectedProvider: Provider | null
  selectedModelId: number | null
  upstreamModelName: string
  keyName: string
  keyExpiresAt: string
  keyAllowedModels: number[]
}

const initialState: WizardState = {
  selectedProvider: null,
  selectedModelId: null,
  upstreamModelName: '',
  keyName: '',
  keyExpiresAt: '',
  keyAllowedModels: [],
}

interface ModelBinding {
  id: number
  provider_id: number
  upstream_model_name: string
}

export default function WizardPage() {
  const [currentStep, setCurrentStep] = useState(0)
  const [providers, setProviders] = useState<Provider[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [modelBindings, setModelBindings] = useState<Record<number, ModelBinding[]>>({})
  const [state, setState] = useState<WizardState>(initialState)
  const [loading, setLoading] = useState(false)
  const [providerModalOpen, setProviderModalOpen] = useState(false)
  const [modelModalOpen, setModelModalOpen] = useState(false)
  const [createdApiKey, setCreatedApiKey] = useState<string>('')
  const { showSuccess, showError } = useToast()

  const providerForm = {
    name: '',
    base_url: '',
    auth_type: 'x_api_key',
    newKey: '',
    remark: '',
  }
  const [newProvider, setNewProvider] = useState(providerForm)

  const modelForm = {
    model_code: '',
    display_name: '',
    description: '',
  }
  const [newModel, setNewModel] = useState(modelForm)

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    try {
      const [providersData, modelsData] = await Promise.all([
        api.providers(),
        api.models()
      ])
      setProviders(providersData.items || [])
      const modelList = modelsData.items || []
      setModels(modelList)
      
      const bindingsMap: Record<number, ModelBinding[]> = {}
      for (const model of modelList) {
        try {
          const bindingData = await api.modelBindings(model.id)
          bindingsMap[model.id] = bindingData.items || []
        } catch {
          bindingsMap[model.id] = []
        }
      }
      setModelBindings(bindingsMap)
    } catch (err) {
      console.error('加载数据失败', err)
    }
  }

  async function createProvider() {
    if (!newProvider.name.trim() || !newProvider.base_url.trim() || !newProvider.newKey.trim()) {
      showError('请填写完整信息')
      return
    }
    try {
      const created = await api.createProvider({
        name: newProvider.name,
        base_url: newProvider.base_url,
        auth_type: newProvider.auth_type,
        newKey: newProvider.newKey,
        remark: newProvider.remark,
        enabled: true,
        provider_type: 'proxy',
        service_type: 'proxy',
        timeout_ms: 60000,
        health_status: 'healthy',
      })
      setProviderModalOpen(false)
      setNewProvider(providerForm)
      await loadData()
      setState(prev => ({ ...prev, selectedProvider: created }))
      showSuccess('上游创建成功')
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建上游失败')
    }
  }

  async function createModel() {
    if (!newModel.model_code.trim()) {
      showError('请输入模型标识')
      return
    }
    try {
      const created = await api.createModel({
        model_code: newModel.model_code,
        display_name: newModel.display_name || newModel.model_code,
        protocol_family: 'openai',
        capability_flags: ['text', 'chat'],
        visible: true,
        status: 'active',
        sort_order: 0,
        description: newModel.description,
        default_params_json: {},
      })
      setModelModalOpen(false)
      setNewModel(modelForm)
      await loadData()
      const createdModel = models.find(m => m.model_code === newModel.model_code)
      if (createdModel) {
        setState(prev => ({ ...prev, selectedModelId: createdModel.id }))
      }
      showSuccess('模型创建成功')
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建模型失败')
    }
  }

  async function goNext() {
    if (currentStep === 0) {
      if (!state.selectedProvider) {
        showError('请选择上游')
        return
      }
    } else if (currentStep === 1) {
      if (!state.selectedModelId) {
        showError('请选择模型')
        return
      }
    } else if (currentStep === 2) {
      if (!state.selectedModelId || !state.upstreamModelName.trim()) {
        showError('请填写上游模型名')
        return
      }
      
      const existingBindings = modelBindings[state.selectedModelId] || []
      const existingBindingToProvider = existingBindings.find(b => b.provider_id === state.selectedProvider!.id)
      
      if (existingBindingToProvider) {
        setState(prev => ({ ...prev, upstreamModelName: existingBindingToProvider.upstream_model_name }))
        showSuccess('已使用现有绑定')
        setCurrentStep(prev => prev + 1)
        return
      }
      
      setLoading(true)
      try {
        await api.createModelBinding(state.selectedModelId, {
          provider_id: state.selectedProvider!.id,
          upstream_model_name: state.upstreamModelName,
          priority: existingBindings.length + 1,
          enabled: true,
          is_same_name: false,
        })
        showSuccess('绑定成功')
      } catch (err) {
        showError(err instanceof Error ? err.message : '绑定失败')
        setLoading(false)
        return
      }
      setLoading(false)
    } else if (currentStep === 3) {
      if (!state.keyName.trim()) {
        showError('请输入密钥名称')
        return
      }
      setLoading(true)
      try {
        const today = new Date().toISOString().split('T')[0]
        const created = await api.createClientKey({
          name: state.keyName,
          remark: '快捷流程创建',
          expires_at: state.keyExpiresAt === today ? '' : state.keyExpiresAt,
          allowed_models: state.keyAllowedModels.length ? state.keyAllowedModels : [],
        })
        setCreatedApiKey(created.plain_key || '')
        await loadData()
        showSuccess('API Key 创建成功')
      } catch (err) {
        showError(err instanceof Error ? err.message : '创建密钥失败')
        setLoading(false)
        return
      }
      setLoading(false)
    }
    setCurrentStep(prev => prev + 1)
  }

  function goBack() {
    if (currentStep > 0) {
      setCurrentStep(prev => prev - 1)
    }
  }

  const canProceed = () => {
    switch (currentStep) {
      case 0: return !!state.selectedProvider
      case 1: return !!state.selectedModelId
      case 2: return !!state.upstreamModelName.trim()
      case 3: return !!state.keyName.trim()
      case 4: return true
      default: return false
    }
  }

  const getCurrentModelBindings = () => {
    if (!state.selectedModelId) return []
    return modelBindings[state.selectedModelId] || []
  }

  const getBindingToSelectedProvider = () => {
    const bindings = getCurrentModelBindings()
    return bindings.find(b => b.provider_id === state.selectedProvider?.id)
  }

  const existingBinding = getBindingToSelectedProvider()

  const getProxyURL = () => {
    if (typeof window === 'undefined') return 'http://127.0.0.1:8080'
    const envBase = process.env.NEXT_PUBLIC_API_BASE_URL
    if (envBase && (envBase.startsWith('http://') || envBase.startsWith('https://'))) {
      return envBase.replace('/admin', '')
    }
    return window.location.origin
  }

  return (
    <div className="mx-auto max-w-5xl p-6">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold" style={{ color: 'var(--foreground)' }}>快捷流程</h1>
        <p className="mt-1 text-sm" style={{ color: 'var(--muted-foreground)' }}>通过简单几步快速配置完整 AI 接入</p>
      </div>

      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex items-center justify-center gap-2">
            {steps.map((step, index) => (
              <div key={index} className="flex items-center">
                <div className="flex flex-col items-center">
                  <div
                    className="flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium transition-colors"
                    style={{
                      background: index < currentStep ? 'var(--success)' : index === currentStep ? 'var(--primary)' : 'var(--muted)',
                      color: index <= currentStep ? 'white' : 'var(--muted-foreground)',
                    }}
                  >
                    {index < currentStep ? <FiCheck className="h-4 w-4" /> : index + 1}
                  </div>
                  <div className="mt-2 text-center">
                    <div className="text-xs font-medium whitespace-nowrap" style={{ color: index === currentStep ? 'var(--foreground)' : 'var(--muted-foreground)' }}>{step.title}</div>
                  </div>
                </div>
                {index < steps.length - 1 && (
                  <div
                    className="mx-3 h-0.5 min-w-[40px] transition-colors"
                    style={{ background: index < currentStep ? 'var(--success)' : 'var(--border)' }}
                  />
                )}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-6">
          {currentStep === 0 && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium" style={{ color: 'var(--foreground)' }}>选择上游</h2>
              <p className="text-sm" style={{ color: 'var(--muted-foreground)' }}>选择已有的上游服务或添加新的上游</p>
              
              <div className="grid gap-3 md:grid-cols-2">
                {providers.map((provider) => (
                  <div
                    key={provider.id}
                    className={`cursor-pointer rounded-[10px] border p-4 transition-all ${state.selectedProvider?.id === provider.id ? 'border-[var(--primary)] bg-[var(--primary)]/10' : 'hover:border-[var(--primary)]/50'}`}
                    style={{ borderColor: state.selectedProvider?.id === provider.id ? 'var(--primary)' : 'var(--border)' }}
                    onClick={() => setState(prev => ({ ...prev, selectedProvider: provider }))}
                  >
                    <div className="flex items-center justify-between">
                      <div className="font-medium" style={{ color: 'var(--foreground)' }}>{provider.name}</div>
                      <GoDotFill className={`h-3 w-3 ${provider.health_status === 'healthy' ? 'text-emerald-400' : 'text-yellow-400'}`} />
                    </div>
                    <div className="mt-1 text-xs truncate" style={{ color: 'var(--muted-foreground)' }}>{provider.base_url}</div>
                  </div>
                ))}
              </div>

              <Button variant="secondary" onClick={() => setProviderModalOpen(true)}>
                <FiPlus className="mr-1 h-4 w-4" />添加上游
              </Button>
            </div>
          )}

          {currentStep === 1 && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium" style={{ color: 'var(--foreground)' }}>选择模型</h2>
              <p className="text-sm" style={{ color: 'var(--muted-foreground)' }}>选择已有模型进行绑定</p>
              
              <div className="space-y-3">
                <div className="grid gap-3 md:grid-cols-2">
                  {models.map((model) => (
                    <div
                      key={model.id}
                      className={`cursor-pointer rounded-[10px] border p-4 transition-all ${state.selectedModelId === model.id ? 'border-[var(--primary)] bg-[var(--primary)]/10' : 'hover:border-[var(--primary)]/50'}`}
                      style={{ borderColor: state.selectedModelId === model.id ? 'var(--primary)' : 'var(--border)' }}
                      onClick={() => setState(prev => ({ ...prev, selectedModelId: model.id }))}
                    >
                      <div className="flex items-center justify-between">
                        <div className="font-medium" style={{ color: 'var(--foreground)' }}>{model.model_code}</div>
                        <GoDotFill className={`h-3 w-3 ${model.status === 'active' ? 'text-emerald-400' : 'text-yellow-400'}`} />
                      </div>
                      <div className="mt-1 text-xs" style={{ color: 'var(--muted-foreground)' }}>{model.display_name || '-'}</div>
                    </div>
                  ))}
                </div>
                <Button variant="secondary" onClick={() => setModelModalOpen(true)}>
                  <FiPlus className="mr-1 h-4 w-4" />添加模型
                </Button>
              </div>
            </div>
          )}

          {currentStep === 2 && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium" style={{ color: 'var(--foreground)' }}>添加绑定</h2>
              <p className="text-sm" style={{ color: 'var(--muted-foreground)' }}>将虚拟模型绑定到上游模型</p>
              
              {existingBinding ? (
                <div className="rounded-[10px] border border-emerald-500/50 bg-emerald-500/10 p-4">
                  <div className="flex items-center gap-2 text-emerald-500">
                    <FiCheck className="h-4 w-4" />
                    <span className="text-sm font-medium">已存在绑定</span>
                  </div>
                  <div className="mt-2 text-sm" style={{ color: 'var(--foreground)' }}>
                    上游模型：{existingBinding.upstream_model_name}
                  </div>
                </div>
              ) : (
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)' }}>
                  <div className="mb-3 flex items-center justify-between">
                    <span className="text-sm" style={{ color: 'var(--foreground)' }}>上游</span>
                    <span className="text-sm font-medium" style={{ color: 'var(--primary)' }}>{state.selectedProvider?.name}</span>
                  </div>
                  <div className="flex items-center gap-4">
                    <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>上游模型</label>
                    <Input
                      className="flex-1"
                      placeholder="如 gpt-4o-mini (上游实际的模型名)"
                      value={state.upstreamModelName}
                      onChange={(e) => setState(prev => ({ ...prev, upstreamModelName: e.target.value }))}
                    />
                  </div>
                </div>
              )}

              <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>
                虚拟模型：{models.find(m => m.id === state.selectedModelId)?.model_code || '-'}
              </div>
            </div>
          )}

          {currentStep === 3 && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium" style={{ color: 'var(--foreground)' }}>创建密钥</h2>
              <p className="text-sm" style={{ color: 'var(--muted-foreground)' }}>生成客户端 API Key 用于调用</p>
              
              <div className="space-y-4">
                <div className="flex items-center gap-4">
                  <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>名称</label>
                  <Input
                    className="flex-1"
                    placeholder="如 测试密钥"
                    value={state.keyName}
                    onChange={(e) => setState(prev => ({ ...prev, keyName: e.target.value }))}
                  />
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>过期时间</label>
                  <div className="flex flex-1 gap-2">
                    <Input
                      type="date"
                      className="flex-1"
                      value={state.keyExpiresAt}
                      onChange={(e) => setState(prev => ({ ...prev, keyExpiresAt: e.target.value }))}
                    />
                    <Select value="" onValueChange={(value) => {
                      if (!value) return
                      const now = new Date()
                      if (value === 'permanent') {
                        setState(prev => ({ ...prev, keyExpiresAt: '' }))
                      } else {
                        now.setDate(now.getDate() + parseInt(value))
                        setState(prev => ({ ...prev, keyExpiresAt: now.toISOString().split('T')[0] }))
                      }
                    }}>
                      <SelectTrigger className="w-32"><SelectValue placeholder="快捷选择" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="permanent">永久</SelectItem>
                        <SelectItem value="7">7天</SelectItem>
                        <SelectItem value="30">30天</SelectItem>
                        <SelectItem value="90">90天</SelectItem>
                        <SelectItem value="365">1年</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>可用模型</label>
                  <div className="flex-1">
                    <Select value={state.keyAllowedModels[0] ? String(state.keyAllowedModels[0]) : 'all'} onValueChange={(value) => setState(prev => ({ ...prev, keyAllowedModels: value === 'all' ? [] : [Number(value)] }))}>
                      <SelectTrigger className="flex-1"><SelectValue placeholder="选择可用模型" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">所有模型</SelectItem>
                        {models.map((item) => (
                          <SelectItem key={item.id} value={String(item.id)}>{item.model_code}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </div>
            </div>
          )}

          {currentStep === 4 && (
            <div className="space-y-6">
              <div className="text-center">
                <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/20">
                  <FiCheck className="h-8 w-8 text-emerald-500" />
                </div>
                <h2 className="text-xl font-semibold" style={{ color: 'var(--foreground)' }}>配置完成！</h2>
                <p className="mt-1 text-sm" style={{ color: 'var(--muted-foreground)' }}>您的 AI 接入已准备就绪</p>
              </div>

              <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(0,0,0,0.2)' }}>
                <div className="mb-3 flex items-center justify-between">
                  <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>配置摘要</div>
                </div>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span style={{ color: 'var(--muted-foreground)' }}>上游</span>
                    <span style={{ color: 'var(--foreground)' }}>{state.selectedProvider?.name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span style={{ color: 'var(--muted-foreground)' }}>模型</span>
                    <span style={{ color: 'var(--foreground)' }}>{models.find(m => m.id === state.selectedModelId)?.model_code || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span style={{ color: 'var(--muted-foreground)' }}>绑定上游</span>
                    <span style={{ color: 'var(--foreground)' }}>{state.upstreamModelName}</span>
                  </div>
                </div>
              </div>

              <div>
                <div className="mb-2 text-sm font-medium" style={{ color: 'var(--foreground)' }}>调用示例</div>
                <div className="relative">
                  <pre className="overflow-auto rounded-xl border p-4 text-sm font-mono whitespace-pre-wrap" style={{ background: 'rgba(0,0,0,0.3)', borderColor: 'var(--border)', color: 'var(--foreground)' }}>
{`curl ${getProxyURL()}/proxy/openai/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer ${createdApiKey}" \\
  --data-raw "{\\"model\\":\\"${models.find(m => m.id === state.selectedModelId)?.model_code || ''}\\",\\"messages\\":[{\\"role\\":\\"user\\",\\"content\\":\\"你好\\"}],\\"max_tokens\\":5}"`}
                  </pre>
                  <Button
                    variant="secondary"
                    size="sm"
                    className="absolute top-2 right-2"
                    onClick={() => {
                      navigator.clipboard.writeText(`curl ${getProxyURL()}/proxy/openai/v1/chat/completions -H "Content-Type: application/json" -H "Authorization: Bearer ${createdApiKey}" --data-raw "{\\"model\\":\\"${models.find(m => m.id === state.selectedModelId)?.model_code || ''}\\",\\"messages\\":[{\\"role\\":\\"user\\",\\"content\\":\\"你好\\"}],\\"max_tokens\\":5}"`)
                      showSuccess('已复制')
                    }}
                  >
                    复制
                  </Button>
                </div>
              </div>
            </div>
          )}
        </CardContent>

        <div className="flex items-center justify-between border-t p-4" style={{ borderColor: 'var(--border)' }}>
          <Button variant="secondary" disabled={currentStep === 0} onClick={goBack}>
            <FiChevronLeft className="mr-1 h-4 w-4" />上一步
          </Button>
          {currentStep < 4 ? (
            <Button onClick={goNext} disabled={!canProceed()} loading={loading}>
              下一步<FiChevronRight className="ml-1 h-4 w-4" />
            </Button>
          ) : (
            <Button onClick={() => { setCurrentStep(0); setState(initialState) }}>
              <RiAddLine className="mr-1 h-4 w-4" />再次创建
            </Button>
          )}
        </div>
      </Card>

      <Modal
        open={providerModalOpen}
        title="新增上游"
        onClose={() => { setProviderModalOpen(false); setNewProvider(providerForm) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setProviderModalOpen(false); setNewProvider(providerForm) }}>取消</Button>
            <Button onClick={createProvider}>创建</Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>名称</label>
            <Input className="flex-1" placeholder="请输入名称" value={newProvider.name} onChange={(e) => setNewProvider({ ...newProvider, name: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>Base URL</label>
            <Input className="flex-1" placeholder="如 https://api.openai.com/v1" value={newProvider.base_url} onChange={(e) => setNewProvider({ ...newProvider, base_url: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>鉴权</label>
            <Select value={newProvider.auth_type} onValueChange={(value) => setNewProvider({ ...newProvider, auth_type: value })}>
              <SelectTrigger className="flex-1"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="bearer">bearer</SelectItem>
                <SelectItem value="x_api_key">x_api_key</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>API Key</label>
            <Input className="flex-1" placeholder="请输入 API Key" value={newProvider.newKey} onChange={(e) => setNewProvider({ ...newProvider, newKey: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>备注</label>
            <Textarea className="flex-1" placeholder="可选" value={newProvider.remark} onChange={(e) => setNewProvider({ ...newProvider, remark: e.target.value })} />
          </div>
        </div>
      </Modal>

      <Modal
        open={modelModalOpen}
        title="新增模型"
        onClose={() => { setModelModalOpen(false); setNewModel(modelForm) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setModelModalOpen(false); setNewModel(modelForm) }}>取消</Button>
            <Button onClick={createModel}>创建</Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>模型标识</label>
            <Input className="flex-1" placeholder="如 gpt-4o-mini" value={newModel.model_code} onChange={(e) => setNewModel({ ...newModel, model_code: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>展示名称</label>
            <Input className="flex-1" placeholder="可选" value={newModel.display_name} onChange={(e) => setNewModel({ ...newModel, display_name: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>描述</label>
            <Textarea className="flex-1" placeholder="可选" value={newModel.description} onChange={(e) => setNewModel({ ...newModel, description: e.target.value })} />
          </div>
        </div>
      </Modal>
    </div>
  )
}