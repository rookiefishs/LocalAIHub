'use client'

import { DragEvent, useEffect, useMemo, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { GoDotFill } from 'react-icons/go'
import { PiGitBranchBold } from 'react-icons/pi'
import { RiAddLine, RiDeleteBinLine, RiDraggable } from 'react-icons/ri'
import { api } from '@/lib/api'
import { StatCard } from '@/components/stat-card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { Card, CardContent } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { PaginationBar } from '@/components/pagination-bar'
import { useToast } from '@/components/ui/toast'
import { useRefresh } from '@/components/refresh-context'
import { describeTextMeaning } from '@/lib/utils'

export default function RoutesPage() {
  const [routes, setRoutes] = useState<any[]>([])
  const [providers, setProviders] = useState<any[]>([])
  const [bindings, setBindings] = useState<any[]>([])
  const [message, setMessage] = useState('')
  const [selectedRoute, setSelectedRoute] = useState<any | null>(null)
  const [detailModalOpen, setDetailModalOpen] = useState(false)
  const [bindingModalOpen, setBindingModalOpen] = useState(false)
  const [pendingDeleteBinding, setPendingDeleteBinding] = useState<any | null>(null)
  const [pendingDeleteRoute, setPendingDeleteRoute] = useState<any | null>(null)
  const [routeModalOpen, setRouteModalOpen] = useState(false)
  const [routeForm, setRouteForm] = useState({ model_code: '', model_name: '' })
  const [loadingCreateRoute, setLoadingCreateRoute] = useState(false)
  const [loadingData, setLoadingData] = useState(true)
  const [loadingLock, setLoadingLock] = useState<number | null>(null)
  const [loadingBinding, setLoadingBinding] = useState(false)
  const [savingBindingOrder, setSavingBindingOrder] = useState(false)
  const [draggingBindingId, setDraggingBindingId] = useState<number | null>(null)
  const [dragOverBindingId, setDragOverBindingId] = useState<number | null>(null)
  const [bindingForm, setBindingForm] = useState({ provider_id: '', provider_key_id: 0, upstream_model_name: '', priority: 1, enabled: true, is_same_name: false })
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const { showSuccess, showError } = useToast()
  const { registerRefresh } = useRefresh()

  async function load() {
    setLoadingData(true)
    try {
      const [routesData, providersData] = await Promise.all([
        api.routes(`page=${page}&page_size=${pageSize}`),
        api.providers()
      ])
      setRoutes(routesData.items || [])
      setTotal(routesData.total || 0)
      setProviders(providersData.items || [])
    } finally {
      setLoadingData(false)
    }
  }

  async function loadBindings(modelId: number) {
    const data = await api.modelBindings(modelId)
    const sortedBindings = (data.items || []).sort((a: any, b: any) => a.priority - b.priority)
    setBindings(sortedBindings)
  }

  useEffect(() => {
    registerRefresh(load)
    load().catch((err) => setMessage(err.message))
  }, [page, pageSize])

  useEffect(() => {
    setPage(1)
  }, [pageSize])

  const stats = useMemo(() => {
    const normal = routes.filter((item) => !item.manual_locked && item.route_status === 'normal').length
    const locked = routes.filter((item) => item.manual_locked).length
    const error = routes.filter((item) => item.route_status === 'fallback' || item.route_status === 'switched').length
    return { total: routes.length, normal, locked, error }
  }, [routes])

  async function toggleLock(item: any) {
    try {
      if (item.manual_locked) {
        await api.unlockRoute(item.virtual_model_id)
        showSuccess('解锁成功')
      } else {
        await api.switchRoute(item.virtual_model_id, {
          target_binding_id: item.current_binding_id,
          manual_lock: true,
          reason: '手动锁定',
        })
        showSuccess('锁定成功')
      }
      await load()
    } catch (err) {
      showError(err instanceof Error ? err.message : '操作失败')
    }
  }

  async function deleteBinding() {
    if (!pendingDeleteBinding) return
    try {
      await api.deleteModelBinding(pendingDeleteBinding.virtual_model_id, pendingDeleteBinding.id)
      showSuccess('删除成功')
      await loadBindings(pendingDeleteBinding.virtual_model_id)
    } catch (err) {
      showError(err instanceof Error ? err.message : '删除失败')
    }
    setPendingDeleteBinding(null)
  }

  async function deleteRoute() {
    if (!pendingDeleteRoute) return
    try {
      await api.deleteRoute(pendingDeleteRoute.virtual_model_id)
      showSuccess('删除成功')
      await load()
    } catch (err) {
      showError(err instanceof Error ? err.message : '删除失败')
    }
    setPendingDeleteRoute(null)
    setDetailModalOpen(false)
  }

  async function createRoute() {
    if (!routeForm.model_code.trim()) {
      showError('请填写模型标识')
      return
    }
    try {
      setLoadingCreateRoute(true)
      await api.createModel({
        model_code: routeForm.model_code.trim(),
        display_name: routeForm.model_name.trim() || routeForm.model_code.trim()
      })
      showSuccess('创建成功')
      setRouteModalOpen(false)
      setRouteForm({ model_code: '', model_name: '' })
      await load()
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建失败')
    } finally {
      setLoadingCreateRoute(false)
    }
  }

  async function submitBinding() {
    if (!selectedRoute || !bindingForm.provider_id || !bindingForm.upstream_model_name.trim()) {
      showError('请填写完整信息')
      return
    }
    try {
      setLoadingBinding(true)
      const nextPriority = bindings.length + 1
      await api.createModelBinding(selectedRoute.virtual_model_id, {
        ...bindingForm,
        priority: nextPriority,
        provider_id: Number(bindingForm.provider_id),
        provider_key_id: bindingForm.provider_key_id || undefined
      })
      showSuccess('添加成功')
      setBindingModalOpen(false)
      setBindingForm({ provider_id: '', provider_key_id: 0, upstream_model_name: '', priority: 1, enabled: true, is_same_name: false })
      await loadBindings(selectedRoute.virtual_model_id)
    } catch (err) {
      showError(err instanceof Error ? err.message : '添加失败')
    } finally {
      setLoadingBinding(false)
    }
  }

  function reorderBindings(items: any[], fromId: number, toId: number) {
    const next = [...items]
    const fromIndex = next.findIndex((item) => item.id === fromId)
    const toIndex = next.findIndex((item) => item.id === toId)
    if (fromIndex === -1 || toIndex === -1 || fromIndex === toIndex) {
      return items
    }
    const [moved] = next.splice(fromIndex, 1)
    next.splice(toIndex, 0, moved)
    return next.map((item, index) => ({ ...item, priority: index + 1 }))
  }

  async function persistBindingOrder(items: any[]) {
    if (!selectedRoute) return
    setSavingBindingOrder(true)
    try {
      await Promise.all(items.map((binding, index) => api.updateModelBinding(selectedRoute.virtual_model_id, binding.id, {
        provider_id: binding.provider_id,
        provider_key_id: binding.provider_key_id || null,
        upstream_model_name: binding.upstream_model_name,
        priority: index + 1,
        enabled: binding.enabled,
        is_same_name: binding.is_same_name,
      })))
      const refreshedSelectedRoute = (await api.routes(`page=${page}&page_size=${pageSize}`)).items?.find((item: any) => item.virtual_model_id === selectedRoute.virtual_model_id) || null
      showSuccess('绑定顺序已更新')
      await loadBindings(selectedRoute.virtual_model_id)
      if (refreshedSelectedRoute) {
        setSelectedRoute(refreshedSelectedRoute)
      }
      await load()
    } catch (err) {
      showError(err instanceof Error ? err.message : '绑定顺序保存失败')
      await loadBindings(selectedRoute.virtual_model_id)
    } finally {
      setSavingBindingOrder(false)
    }
  }

  function handleBindingDragStart(bindingId: number) {
    setDraggingBindingId(bindingId)
    setDragOverBindingId(bindingId)
  }

  function handleBindingDragOver(event: DragEvent<HTMLDivElement>, bindingId: number) {
    event.preventDefault()
    if (draggingBindingId === null || draggingBindingId === bindingId) return
    setDragOverBindingId(bindingId)
    setBindings((current) => reorderBindings(current, draggingBindingId, bindingId))
    setDraggingBindingId(bindingId)
  }

  async function handleBindingDrop() {
    setDragOverBindingId(null)
    if (draggingBindingId === null) return
    setDraggingBindingId(null)
  }

  function handleBindingDragEnd() {
    setDraggingBindingId(null)
    setDragOverBindingId(null)
  }

  function openDetail(item: any) {
    setSelectedRoute(item)
    loadBindings(item.virtual_model_id)
    setDetailModalOpen(true)
  }

  function openBindingModal() {
    setBindingForm({ provider_id: '', provider_key_id: 0, upstream_model_name: '', priority: 1, enabled: true, is_same_name: false })
    setBindingModalOpen(true)
  }

  return (
    <div className="space-y-4">
      <div className="grid w-full gap-4 md:grid-cols-4">
        <StatCard title="总数" value={stats.total} icon={<PiGitBranchBold className="h-4 w-4" />} />
        <StatCard title="正常" value={stats.normal} icon={<GoDotFill className="h-4 w-4 text-emerald-400" />} />
        <StatCard title="锁定" value={stats.locked} icon={<GoDotFill className="h-4 w-4 text-blue-400" />} />
        <StatCard title="异常" value={stats.error} icon={<GoDotFill className="h-4 w-4 text-yellow-400" />} />
      </div>

      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
          <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>路由列表</div>
          <Button onClick={() => setRouteModalOpen(true)} size="sm" title="添加路由"><RiAddLine className="h-4 w-4" /></Button>
        </div>
        <CardContent className="p-4">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {loadingData ? (
              <div className="col-span-full h-32 flex items-center justify-center text-sm" style={{ color: 'var(--muted-foreground)' }}>
                <div className="flex items-center justify-center gap-2">
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600" />
                  加载中...
                </div>
              </div>
            ) : routes.length === 0 ? (
              <div className="col-span-full h-32 flex items-center justify-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</div>
            ) : routes.map((item) => {
              const statusColor = item.manual_locked ? 'text-blue-400' : item.route_status === 'normal' ? 'text-emerald-400' : 'text-yellow-400'
              const currentBindingText = item.current_binding_name ? `${item.current_binding_name} (#${item.current_binding_id})` : (item.current_binding_id ?? '-')
              const switchReasonText = describeTextMeaning(item.last_switch_reason || (item.route_status === 'switched' ? 'auto_switch' : ''))
              return (
                <div key={item.virtual_model_id} className="rounded-[10px] border p-4 cursor-pointer hover:border-[var(--accent)] transition-colors" style={{ borderColor: 'var(--border)' }} onClick={() => openDetail(item)}>
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-1.5">
                      <GoDotFill className={`h-3 w-3 ${statusColor}`} />
                      <span className="font-medium font-mono text-sm" style={{ color: 'var(--foreground)' }}>{item.model_code}</span>
                    </div>
                    <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                      <Button variant="secondary" size="sm" loading={loadingLock === item.virtual_model_id} onClick={() => { setLoadingLock(item.virtual_model_id); toggleLock(item).finally(() => setLoadingLock(null)) }} className="h-7 text-xs">
                        {item.manual_locked ? '解锁' : '锁定'}
                      </Button>
                      <Button variant="ghost" size="sm" className="h-7 text-xs px-1" onClick={() => setPendingDeleteRoute(item)}>
                        <RiDeleteBinLine className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </div>
                  <div className="mt-2 space-y-1 text-xs" style={{ color: 'var(--muted-foreground)' }}>
                    <div>状态: {item.manual_locked ? '已锁定' : item.route_status === 'normal' ? '正常' : '降级中'}</div>
                    <div>当前绑定: {currentBindingText}</div>
                    {item.route_status === 'switched' ? <div>切换原因: {switchReasonText}</div> : null}
                  </div>
                </div>
              )
            })}
          </div>
        </CardContent>
        <PaginationBar page={page} pageSize={pageSize} total={total} onPageChange={setPage} onPageSizeChange={setPageSize} />
      </Card>

      {message ? <Card><CardContent className="p-4 text-sm">{message}</CardContent></Card> : null}

      <Modal
        open={detailModalOpen}
        title="路由详情"
        onClose={() => setDetailModalOpen(false)}
        footer={
          <div className="flex justify-between gap-2">
            <div className="flex items-center gap-2">
              {selectedRoute && bindings.length > 1 && (
                <Button variant="secondary" onClick={async () => await persistBindingOrder([...bindings])} disabled={savingBindingOrder} loading={savingBindingOrder}>
                  保存顺序
                </Button>
              )}
              {selectedRoute && (
                <Button variant="destructive" onClick={() => setPendingDeleteRoute(selectedRoute)}>
                  <RiDeleteBinLine className="mr-1 h-4 w-4" /> 删除路由
                </Button>
              )}
            </div>
            <div className="flex gap-2">
              <Button variant="secondary" onClick={() => setDetailModalOpen(false)}>关闭</Button>
              {selectedRoute && (
                <Button variant={selectedRoute.manual_locked ? 'default' : 'destructive'} onClick={() => { toggleLock(selectedRoute); setDetailModalOpen(false) }}>
                  {selectedRoute.manual_locked ? '解锁' : '锁定'}
                </Button>
              )}
            </div>
          </div>
        }
      >
        {selectedRoute && (
          <div className="space-y-4">
            <div className="flex items-center gap-4">
              <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>模型</label>
              <Input className="flex-1" value={selectedRoute.model_code} readOnly />
            </div>
            <div className="flex items-center gap-4">
              <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>当前绑定</label>
              <Input className="flex-1" value={selectedRoute.current_binding_name ? `${selectedRoute.current_binding_name} (#${selectedRoute.current_binding_id})` : (selectedRoute.current_binding_id ?? '-')} readOnly />
            </div>
            <div className="flex items-center gap-4">
              <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>状态</label>
              <Input className="flex-1" value={selectedRoute.manual_locked ? '已锁定' : selectedRoute.route_status === 'normal' ? '正常' : '降级'} readOnly />
            </div>
            <div className="flex items-center gap-4">
              <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>切换原因</label>
              <Input className="flex-1" value={describeTextMeaning(selectedRoute.last_switch_reason || (selectedRoute.route_status === 'switched' ? 'auto_switch' : ''))} readOnly />
            </div>
            <div className="flex items-center gap-4">
              <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>最近切换</label>
              <Input className="flex-1" value={selectedRoute.last_switch_at ? new Date(selectedRoute.last_switch_at).toLocaleString() : '-'} readOnly />
            </div>

            <div className="mt-6">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>绑定列表</span>
                <Button size="sm" onClick={openBindingModal}>
                  <RiAddLine className="mr-1 h-4 w-4" /> 添加绑定
                </Button>
              </div>
              {bindings.length === 0 ? (
                <div className="text-sm py-4 text-center" style={{ color: 'var(--muted-foreground)' }}>暂无绑定</div>
              ) : (
                <div className="space-y-2 rounded-[10px] border p-4 text-sm" style={{ borderColor: 'var(--border)', background: 'rgba(255,255,255,0.03)' }}>
                  <AnimatePresence initial={false}>{bindings.map((binding: any) => {
                     const provider = providers.find((p: any) => p.id === binding.provider_id)
                     return (
                    <motion.div
                      key={binding.id}
                      layout
                      transition={{ type: 'spring', stiffness: 380, damping: 30 }}
                      draggable={!savingBindingOrder}
                      onDragStart={() => handleBindingDragStart(binding.id)}
                      onDragOver={(event) => handleBindingDragOver(event, binding.id)}
                      onDrop={() => { void handleBindingDrop() }}
                      onDragEnd={handleBindingDragEnd}
                      className="flex items-center justify-between gap-3 rounded-lg border p-3 transition-colors"
                      style={{
                        borderColor: dragOverBindingId === binding.id ? 'var(--primary)' : 'var(--border)',
                        opacity: draggingBindingId === binding.id ? 0.75 : 1,
                        background: dragOverBindingId === binding.id ? 'rgba(59,130,246,0.08)' : 'transparent',
                      }}
                    >
                      <div className="min-w-0 flex flex-1 items-center gap-3 text-sm">
                        <RiDraggable className="h-4 w-4 shrink-0 cursor-move text-slate-400" />
                        <div className="min-w-0 flex-1">
                          <div className="truncate" style={{ color: 'var(--foreground)' }}>{provider?.name || '未知供应商'} / {binding.upstream_model_name}</div>
                        </div>
                      </div>
                      <Button variant="ghost" size="sm" onClick={() => setPendingDeleteBinding({ ...binding, virtual_model_id: selectedRoute.virtual_model_id })}>
                        删除
                      </Button>
                    </motion.div>
                  )})}</AnimatePresence>
                </div>
              )}
              {bindings.length > 1 ? <div className="mt-2 text-xs" style={{ color: 'var(--muted-foreground)' }}>拖动左侧图标即可调整绑定顺序，点击"保存顺序"按钮后生效。</div> : null}
            </div>
          </div>
        )}
      </Modal>

      <Modal
        open={bindingModalOpen}
        title="添加绑定"
        onClose={() => setBindingModalOpen(false)}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setBindingModalOpen(false)}>取消</Button>
            <Button onClick={submitBinding} disabled={!bindingForm.provider_id || !bindingForm.upstream_model_name.trim() || loadingBinding} loading={loadingBinding}>
              添加
            </Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>供应商</label>
            <Select value={bindingForm.provider_id} onValueChange={(v) => setBindingForm({ ...bindingForm, provider_id: v })}>
              <SelectTrigger className="flex-1">
                <SelectValue placeholder="选择供应商" />
              </SelectTrigger>
              <SelectContent>
                {providers.map((p: any) => (
                  <SelectItem key={p.id} value={String(p.id)}>{p.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>上游模型</label>
            <Input className="flex-1" value={bindingForm.upstream_model_name} onChange={(e) => setBindingForm({ ...bindingForm, upstream_model_name: e.target.value })} placeholder="gpt-4o" />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>顺序</label>
            <Input className="flex-1" value={`自动追加到第 ${bindings.length + 1} 位`} readOnly />
          </div>
        </div>
      </Modal>

      <ConfirmDialog
        open={!!pendingDeleteBinding}
        title="删除绑定"
        description={`确定要删除绑定 "${pendingDeleteBinding?.upstream_model_name}" 吗？此操作不可恢复。`}
        confirmText="删除"
        confirmVariant="destructive"
        onConfirm={deleteBinding}
        onCancel={() => setPendingDeleteBinding(null)}
      />

      <ConfirmDialog
        open={!!pendingDeleteRoute}
        title="删除路由"
        description={`确定要删除路由 "${pendingDeleteRoute?.model_code}" 吗？此操作不可恢复。`}
        confirmText="删除"
        confirmVariant="destructive"
        onConfirm={deleteRoute}
        onCancel={() => setPendingDeleteRoute(null)}
      />

      <Modal
        open={routeModalOpen}
        title="添加路由"
        onClose={() => setRouteModalOpen(false)}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setRouteModalOpen(false)}>取消</Button>
            <Button onClick={createRoute} disabled={!routeForm.model_code.trim() || loadingCreateRoute} loading={loadingCreateRoute}>
              创建
            </Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>模型标识</label>
            <Input className="flex-1" value={routeForm.model_code} onChange={(e) => setRouteForm({ ...routeForm, model_code: e.target.value })} placeholder="gpt-4o-mini" />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-20 text-sm" style={{ color: 'var(--foreground)' }}>显示名称</label>
            <Input className="flex-1" value={routeForm.model_name} onChange={(e) => setRouteForm({ ...routeForm, model_name: e.target.value })} placeholder="GPT-4o Mini (可选)" />
          </div>
        </div>
      </Modal>

      <ConfirmDialog
        open={!!pendingDeleteRoute}
        title="删除路由"
        description={`确定要删除路由 "${pendingDeleteRoute?.model_code}" 吗？此操作不可恢复。`}
        confirmText="删除"
        confirmVariant="destructive"
        onConfirm={deleteRoute}
        onCancel={() => setPendingDeleteRoute(null)}
      />
    </div>
  )
}
