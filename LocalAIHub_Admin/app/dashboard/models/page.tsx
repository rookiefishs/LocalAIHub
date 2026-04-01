'use client'

import { useEffect, useMemo, useState } from 'react'
import { FiLayers } from 'react-icons/fi'
import { GoDotFill } from 'react-icons/go'
import { HiOutlineCubeTransparent, HiOutlineBolt } from 'react-icons/hi2'
import { PiGitBranchBold } from 'react-icons/pi'
import { RiAddLine, RiDeleteBinLine, RiEditLine } from 'react-icons/ri'
import { api } from '@/lib/api'
import { StatCard } from '@/components/stat-card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { Card, CardContent } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { PaginationBar } from '@/components/pagination-bar'
import { useToast } from '@/components/ui/toast'
import { useRefresh } from '@/components/refresh-context'

const initialModelForm = {
  model_code: '',
  display_name: '',
  protocol_family: 'openai',
  capability_flags: ['text', 'chat'],
  visible: true,
  status: 'active',
  sort_order: 0,
  description: '',
  default_params_json: {},
}

interface Binding {
  id: number
  provider_id: number
  upstream_model_name: string
  priority: number
  enabled: boolean
  is_same_name: boolean
  provider_key_id?: number
}

const initialBindingForm = {
  provider_id: '',
  provider_key_id: 0,
  upstream_model_name: '',
  priority: 1,
  enabled: true,
  is_same_name: false,
}

export default function ModelsPage() {
  const [models, setModels] = useState<any[]>([])
  const [providers, setProviders] = useState<any[]>([])
  const [bindings, setBindings] = useState<Binding[]>([])
  const [selectedModelId, setSelectedModelId] = useState<number | null>(null)
  const [selectedModelName, setSelectedModelName] = useState<string>('')
  const [message, setMessage] = useState('')
  const [editingModelId, setEditingModelId] = useState<number | null>(null)
  const [modelModalOpen, setModelModalOpen] = useState(false)
  const [bindingDetailOpen, setBindingDetailOpen] = useState(false)
  const [addBindingOpen, setAddBindingOpen] = useState(false)
  const [pendingDeleteModel, setPendingDeleteModel] = useState<any | null>(null)
  const [pendingDeleteBinding, setPendingDeleteBinding] = useState<Binding | null>(null)
  const [editingBinding, setEditingBinding] = useState<Binding | null>(null)
  const [form, setForm] = useState(initialModelForm)
  const [bindingForm, setBindingForm] = useState(initialBindingForm)
  const [loadingSubmit, setLoadingSubmit] = useState(false)
  const [loadingBinding, setLoadingBinding] = useState(false)
  const [testingBindings, setTestingBindings] = useState<Set<number>>(new Set())
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)
  const { showSuccess, showError } = useToast()
  const { registerRefresh } = useRefresh()

  async function load() {
    const [modelsData, providersData] = await Promise.all([
      api.models(`page=${page}&page_size=${pageSize}`),
      api.providers()
    ])
    setModels(modelsData.items || [])
    setTotal(modelsData.total || 0)
    setProviders(providersData.items || [])
  }

  async function loadBindings(id: number) {
    const data = await api.modelBindings(id)
    const sortedBindings = (data.items || []).sort((a: Binding, b: Binding) => a.priority - b.priority)
    setBindings(sortedBindings)
    setSelectedModelId(id)
    const model = models.find(m => m.id === id)
    setSelectedModelName(model?.model_code || '')
  }

  useEffect(() => { registerRefresh(load); load().catch((err) => setMessage(err.message)) }, [page, pageSize])

  useEffect(() => {
    setPage(1)
  }, [pageSize])

  const stats = useMemo(() => ({
    total: models.length,
    visible: models.filter((item) => item.visible).length,
    active: models.filter((item) => item.status === 'active').length,
    selectedBindings: bindings.length,
  }), [models, bindings])

  async function submitModel() {
    if (!form.model_code.trim()) {
      showError('请输入模型名')
      return
    }
    try {
      if (editingModelId) {
        await api.updateModel(editingModelId, form)
        showSuccess('保存成功')
      } else {
        await api.createModel(form)
        showSuccess('创建成功')
      }
      setForm(initialModelForm)
      setEditingModelId(null)
      setModelModalOpen(false)
      await load()
    } catch (err) {
      showError(err instanceof Error ? err.message : '保存失败')
    }
  }

  function startEdit(item: any) {
    setEditingModelId(item.id)
    setForm({
      model_code: item.model_code,
      display_name: item.display_name,
      protocol_family: item.protocol_family,
      capability_flags: item.capability_flags || ['text', 'chat'],
      visible: item.visible,
      status: item.status,
      sort_order: item.sort_order || 0,
      description: item.description || '',
      default_params_json: item.default_params_json || {},
    })
    setModelModalOpen(true)
  }

  function openBindingDetail(item: any) {
    loadBindings(item.id).then(() => setBindingDetailOpen(true)).catch((err) => setMessage(err.message))
  }

  async function addBinding() {
    if (!selectedModelId || !bindingForm.provider_id || !bindingForm.upstream_model_name.trim()) {
      showError('请填写完整信息')
      return
    }
    setLoadingBinding(true)
    try {
      await api.createModelBinding(selectedModelId, {
        provider_id: Number(bindingForm.provider_id),
        provider_key_id: bindingForm.provider_key_id || null,
        upstream_model_name: bindingForm.upstream_model_name,
        priority: bindings.length + 1,
        enabled: bindingForm.enabled,
        is_same_name: bindingForm.is_same_name,
      })
      showSuccess('绑定成功')
      setBindingForm(initialBindingForm)
      setAddBindingOpen(false)
      await loadBindings(selectedModelId)
    } catch (err) {
      showError(err instanceof Error ? err.message : '绑定失败')
    } finally {
      setLoadingBinding(false)
    }
  }

  async function updateBinding(binding: Binding) {
    if (!selectedModelId) return
    setLoadingBinding(true)
    try {
      await api.updateModelBinding(selectedModelId, binding.id, {
        provider_id: binding.provider_id,
        provider_key_id: binding.provider_key_id || null,
        upstream_model_name: binding.upstream_model_name,
        priority: binding.priority,
        enabled: binding.enabled,
        is_same_name: binding.is_same_name,
      })
      showSuccess('更新成功')
      setEditingBinding(null)
      await loadBindings(selectedModelId)
    } catch (err) {
      showError(err instanceof Error ? err.message : '更新失败')
    } finally {
      setLoadingBinding(false)
    }
  }

  async function deleteBinding() {
    if (!pendingDeleteBinding || !selectedModelId) return
    try {
      await api.deleteModelBinding(selectedModelId, pendingDeleteBinding.id)
      setPendingDeleteBinding(null)
      showSuccess('删除成功')
      await loadBindings(selectedModelId)
    } catch (err) {
      showError(err instanceof Error ? err.message : '删除失败')
    }
  }

  async function handleDragStart(index: number) {
    setDraggedIndex(index)
  }

  async function handleDragOver(e: React.DragEvent, index: number) {
    e.preventDefault()
  }

  async function handleDrop(targetIndex: number) {
    if (draggedIndex === null || draggedIndex === targetIndex || !selectedModelId) return
    
    const newBindings = [...bindings]
    const [draggedItem] = newBindings.splice(draggedIndex, 1)
    newBindings.splice(targetIndex, 0, draggedItem)
    
    const updatedPriorities = newBindings.map((b, i) => ({ ...b, priority: i + 1 }))
    setBindings(updatedPriorities)
    setDraggedIndex(null)
    
    setLoadingBinding(true)
    try {
      for (let i = 0; i < updatedPriorities.length; i++) {
        const binding = updatedPriorities[i]
        await api.updateModelBinding(selectedModelId, binding.id, {
          provider_id: binding.provider_id,
          provider_key_id: binding.provider_key_id || null,
          upstream_model_name: binding.upstream_model_name,
          priority: i + 1,
          enabled: binding.enabled,
          is_same_name: binding.is_same_name,
        })
      }
      showSuccess('排序已更新')
    } catch (err) {
      showError(err instanceof Error ? err.message : '更新排序失败')
      await loadBindings(selectedModelId)
    } finally {
      setLoadingBinding(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="总数" value={stats.total} icon={<HiOutlineCubeTransparent className="h-4 w-4" />} />
        <StatCard title="可见" value={stats.visible} icon={<FiLayers className="h-4 w-4" />} />
        <StatCard title="正常" value={stats.active} icon={<GoDotFill className="h-4 w-4 text-emerald-400" />} />
        <StatCard title="绑定" value={stats.selectedBindings} icon={<PiGitBranchBold className="h-4 w-4" />} />
      </div>

      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
          <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>虚拟模型列表</div>
          <Button onClick={() => { setEditingModelId(null); setForm(initialModelForm); setModelModalOpen(true) }} size="sm"><RiAddLine className="mr-1 h-4 w-4" />新增</Button>
        </div>
        <CardContent className="p-6">
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {models.length === 0 ? (
              <div className="col-span-full h-32 flex items-center justify-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</div>
            ) : models.map((item) => (
              <Card key={item.id}>
                <CardContent className="p-5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <GoDotFill className={`h-3 w-3 ${item.status === 'active' ? 'text-emerald-400' : 'text-yellow-400'}`} />
                      <div className="font-mono text-sm" style={{ color: 'var(--foreground)' }}>{item.model_code}</div>
                    </div>
                    <div className="flex gap-2">
                      <Button variant="secondary" size="sm" onClick={() => openBindingDetail(item)}>绑定</Button>
                      <Button variant="secondary" size="sm" onClick={() => startEdit(item)}>编辑</Button>
                      <Button variant="destructive" size="sm" onClick={() => setPendingDeleteModel(item)}>删除</Button>
                    </div>
                  </div>
                  <div className="mt-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>{item.display_name}</div>
                  <div className="mt-4 flex flex-wrap gap-2 text-xs">
                    <span className="badge border" style={{ borderColor: 'var(--border)', background: 'rgba(255,255,255,0.03)' }}>{item.protocol_family}</span>
                    <span className="badge border" style={{ borderColor: 'var(--border)', background: 'rgba(255,255,255,0.03)' }}>{item.visible ? '可见' : '隐藏'}</span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
        <PaginationBar page={page} pageSize={pageSize} total={total} onPageChange={setPage} onPageSizeChange={setPageSize} />
      </Card>

      {message ? <Card><CardContent className="p-4 text-sm" style={{ color: 'var(--muted-foreground)' }}>{message}</CardContent></Card> : null}
      
      <Modal
        open={modelModalOpen}
        title={editingModelId ? '编辑模型' : '新增模型'}
        onClose={() => { setModelModalOpen(false); setEditingModelId(null); setForm(initialModelForm) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setModelModalOpen(false); setEditingModelId(null); setForm(initialModelForm) }}>取消</Button>
            <Button loading={loadingSubmit} onClick={async () => { setLoadingSubmit(true); try { await submitModel() } finally { setLoadingSubmit(false) } }}>保存</Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>模型名</label>
            <Input className="flex-1" placeholder="请输入模型名" value={form.model_code} onChange={(e) => setForm({ ...form, model_code: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>展示名</label>
            <Input className="flex-1" placeholder="请输入展示名称" value={form.display_name} onChange={(e) => setForm({ ...form, display_name: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>协议</label>
            <div className="flex-1">
              <Select value={form.protocol_family} onValueChange={(value) => setForm({ ...form, protocol_family: value })}>
                <SelectTrigger className="flex-1"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="openai">OpenAI</SelectItem>
                  <SelectItem value="anthropic">Anthropic</SelectItem>
                  <SelectItem value="gemini">Gemini</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>说明</label>
            <Textarea className="flex-1" placeholder="请输入说明" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </div>
        </div>
      </Modal>

      <Modal
        open={bindingDetailOpen}
        title={`上游绑定 - ${selectedModelName}`}
        onClose={() => { setBindingDetailOpen(false); setSelectedModelId(null); setBindings([]) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setBindingDetailOpen(false); setSelectedModelId(null); setBindings([]) }}>关闭</Button>
            <Button onClick={() => { setBindingForm(initialBindingForm); setAddBindingOpen(true) }}><RiAddLine className="mr-1 h-4 w-4" />新增绑定</Button>
          </div>
        }
      >
        <div className="space-y-3">
          {bindings.length === 0 ? (
            <div className="py-8 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>
              暂无绑定<br />
              <span className="text-xs">点击"新增绑定"添加上游</span>
            </div>
          ) : (
            <div className="space-y-2">
              <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>拖拽可调整优先级，优先级数字越小越优先</div>
              {bindings.map((binding, index) => {
                const provider = providers.find(p => p.id === binding.provider_id)
                const isEditing = editingBinding?.id === binding.id
                return (
                  <div
                    key={binding.id}
                    draggable
                    onDragStart={() => handleDragStart(index)}
                    onDragOver={(e) => handleDragOver(e, index)}
                    onDrop={() => handleDrop(index)}
                    className={`flex items-center gap-3 rounded-[10px] border p-3 ${draggedIndex === index ? 'opacity-50' : ''}`}
                    style={{ 
                      borderColor: 'var(--border)', 
                      background: isEditing ? 'rgba(255,255,255,0.05)' : 'rgba(255,255,255,0.02)',
                      cursor: 'grab'
                    }}
                  >
                    <div className="flex h-6 w-6 items-center justify-center rounded-full text-xs font-medium" style={{ background: 'var(--primary)', color: 'var(--primary-foreground)' }}>
                      {binding.priority}
                    </div>
                    {isEditing ? (
                      <div className="flex-1 space-y-2">
                        <Input
                          className="h-8 text-sm"
                          value={editingBinding.upstream_model_name}
                          onChange={(e) => setEditingBinding({ ...editingBinding, upstream_model_name: e.target.value })}
                          placeholder="上游模型名"
                        />
                        <div className="flex gap-2">
                          <Select
                            value={String(editingBinding.provider_id)}
                            onValueChange={(v) => setEditingBinding({ ...editingBinding, provider_id: Number(v) })}
                          >
                            <SelectTrigger className="h-8 flex-1 text-sm"><SelectValue /></SelectTrigger>
                            <SelectContent>
                              {providers.map(p => (
                                <SelectItem key={p.id} value={String(p.id)}>{p.name}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <Button size="sm" className="h-8" onClick={() => updateBinding(editingBinding)} loading={loadingBinding}>保存</Button>
                          <Button size="sm" variant="secondary" className="h-8" onClick={() => setEditingBinding(null)}>取消</Button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <div style={{ color: 'var(--foreground)' }}>{binding.upstream_model_name}</div>
                            <div className={`h-2 w-2 rounded-full ${binding.enabled ? 'bg-emerald-400' : 'bg-gray-400'}`} />
                          </div>
                          <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>{provider?.name || `p${binding.provider_id}`}</div>
                        </div>
                        <div className="flex items-center gap-1">
                          <Button variant="ghost" size="sm" className="h-7 w-7 p-0" loading={testingBindings.has(binding.id)} onClick={async () => {
                            if (!selectedModelId) return
                            setTestingBindings(prev => new Set(prev).add(binding.id))
                            try {
                              const result = await api.testModelBinding(selectedModelId, binding.id)
                              showSuccess(`测试成功: ${result.model || 'OK'}`)
                            } catch (err) {
                              showError(err instanceof Error ? err.message : '测试失败')
                            } finally {
                              setTestingBindings(prev => { const next = new Set(prev); next.delete(binding.id); return next })
                            }
                          }}>
                            <HiOutlineBolt className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={() => setEditingBinding(binding)}>
                            <RiEditLine className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" className="h-7 w-7 p-0 text-red-400" onClick={() => setPendingDeleteBinding(binding)}>
                            <RiDeleteBinLine className="h-4 w-4" />
                          </Button>
                        </div>
                      </>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </Modal>

      <Modal
        open={addBindingOpen}
        title="新增绑定"
        onClose={() => { setAddBindingOpen(false); setBindingForm(initialBindingForm) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setAddBindingOpen(false); setBindingForm(initialBindingForm) }}>取消</Button>
            <Button loading={loadingBinding} onClick={addBinding}>保存</Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>上游</label>
            <div className="flex-1">
              <Select value={bindingForm.provider_id} onValueChange={(value) => setBindingForm({ ...bindingForm, provider_id: value })}>
                <SelectTrigger><SelectValue placeholder="选择上游" /></SelectTrigger>
                <SelectContent>
                  {providers.map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>模型名</label>
            <Input 
              className="flex-1"
              placeholder="上游模型名（如 gpt-4o）" 
              value={bindingForm.upstream_model_name} 
              onChange={(e) => setBindingForm({ ...bindingForm, upstream_model_name: e.target.value })} 
            />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>启用</label>
            <input
              type="checkbox"
              checked={bindingForm.enabled}
              onChange={(e) => setBindingForm({ ...bindingForm, enabled: e.target.checked })}
              className="h-4 w-4"
            />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>同名</label>
            <input
              type="checkbox"
              checked={bindingForm.is_same_name}
              onChange={(e) => setBindingForm({ ...bindingForm, is_same_name: e.target.checked })}
              className="h-4 w-4"
            />
          </div>
        </div>
      </Modal>

      <ConfirmDialog
        open={Boolean(pendingDeleteModel)}
        title="删除确认"
        description={pendingDeleteModel ? `确认删除模型「${pendingDeleteModel.model_code}」吗？` : ''}
        confirmText="确认删除"
        confirmVariant="destructive"
        onCancel={() => setPendingDeleteModel(null)}
        onConfirm={() => {
          if (!pendingDeleteModel) return
          api.deleteModel(pendingDeleteModel.id).then(async () => {
            if (selectedModelId === pendingDeleteModel.id) {
              setSelectedModelId(null)
              setBindings([])
              setBindingDetailOpen(false)
            }
            setPendingDeleteModel(null)
            showSuccess('删除成功')
            await load()
          }).catch((err) => showError(err.message))
        }}
      />
      <ConfirmDialog
        open={Boolean(pendingDeleteBinding)}
        title="删除确认"
        description={pendingDeleteBinding ? `确认删除绑定「${pendingDeleteBinding.upstream_model_name}」吗？` : ''}
        confirmText="确认删除"
        confirmVariant="destructive"
        onCancel={() => setPendingDeleteBinding(null)}
        onConfirm={deleteBinding}
      />
    </div>
  )
}
