'use client'

import { DragEvent, useEffect, useMemo, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { FiEdit2, FiTrash2 } from 'react-icons/fi'
import { GoDotFill } from 'react-icons/go'
import { HiOutlineKey } from 'react-icons/hi2'
import { PiPlugsConnectedBold } from 'react-icons/pi'
import { RiAddLine, RiDraggable } from 'react-icons/ri'
import { api } from '@/lib/api'
import { StatCard } from '@/components/stat-card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { PaginationBar } from '@/components/pagination-bar'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useToast } from '@/components/ui/toast'
import { useRefresh } from '@/components/refresh-context'

const initialProviderForm = {
  name: '',
  provider_type: 'openai',
  service_type: 'text',
  base_url: '',
  auth_type: 'x_api_key',
  timeout_ms: 60000,
  enabled: true,
  health_status: 'healthy',
  remark: '',
  new_key: '',
}

export default function UpstreamsPage() {
  const [providers, setProviders] = useState<any[]>([])
  const [providerKeys, setProviderKeys] = useState<any[]>([])
  const [selectedProviderId, setSelectedProviderId] = useState<number | null>(null)
  const [message, setMessage] = useState('')
  const [loadingData, setLoadingData] = useState(true)
  const [editingProviderId, setEditingProviderId] = useState<number | null>(null)
  const [providerModalOpen, setProviderModalOpen] = useState(false)
  const [providerKeyModalOpen, setProviderKeyModalOpen] = useState(false)
  const [pendingDeleteProvider, setPendingDeleteProvider] = useState<any | null>(null)
  const [pendingDeleteProviderKey, setPendingDeleteProviderKey] = useState<any | null>(null)
  const [form, setForm] = useState(initialProviderForm)
  const [keyForm, setKeyForm] = useState({ secret: '', priority: 1, remark: '' })
  const [loadingStatus, setLoadingStatus] = useState<number | null>(null)
  const [loadingSubmit, setLoadingSubmit] = useState(false)
  const [loadingKeySubmit, setLoadingKeySubmit] = useState(false)
  const [testingProviders, setTestingProviders] = useState<Set<number>>(new Set())
  const [testingKeys, setTestingKeys] = useState<Set<number>>(new Set())
  const [editingKeyId, setEditingKeyId] = useState<number | null>(null)
  const [editingKeyForm, setEditingKeyForm] = useState({ secret: '', priority: 1, remark: '' })
  const [draggingKeyId, setDraggingKeyId] = useState<number | null>(null)
  const [dragOverKeyId, setDragOverKeyId] = useState<number | null>(null)
  const [savingKeyOrder, setSavingKeyOrder] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const { showSuccess, showError } = useToast()
  const { registerRefresh } = useRefresh()

  async function load() {
    setLoadingData(true)
    try {
      const data = await api.providers(`page=${page}&page_size=${pageSize}`)
      setProviders(data.items || [])
      setTotal(data.total || 0)
    } catch (err) {
      setMessage(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoadingData(false)
    }
  }

  useEffect(() => {
    registerRefresh(load)
    load().catch((err) => setMessage(err.message))
  }, [page, pageSize])

  useEffect(() => {
    setPage(1)
  }, [pageSize])

  const stats = useMemo(() => {
    const usable = providers.filter((item) => item.enabled && item.health_status === 'healthy').length
    const disabled = providers.filter((item) => !item.enabled).length
    const unavailable = providers.filter((item) => item.enabled && item.health_status === 'disabled').length
    return { total: providers.length, usable, disabled, unavailable }
  }, [providers])

  async function submitProvider() {
    if (!form.name.trim()) {
      showError('请输入名称')
      return
    }
    if (!form.base_url.trim()) {
      showError('请输入 Base URL')
      return
    }
    if (!editingProviderId && !form.new_key.trim()) {
      showError('请输入 API Key')
      return
    }
    try {
      if (editingProviderId) {
        await api.updateProvider(editingProviderId, form)
        showSuccess('保存成功')
      } else {
        const created = await api.createProvider(form)
        showSuccess('创建成功')
        if (form.new_key.trim() && created?.id) {
          setTestingProviders(prev => new Set(prev).add(created.id))
          try {
            const result = await api.testProvider(created.id)
            if (result?.success) {
              showSuccess(`创建成功，自动识别鉴权为 ${result?.auth_type || 'x_api_key'}，测试通过`)
            } else {
              showError(result?.message || '创建成功，但测试失败')
            }
          } catch (testErr) {
            showError(`创建成功，但测试失败: ${testErr instanceof Error ? testErr.message : '未知错误'}`)
          } finally {
            setTestingProviders(prev => { const next = new Set(prev); next.delete(created.id); return next })
            await load()
          }
        } else {
          await load()
        }
      }
      setForm(initialProviderForm)
      setEditingProviderId(null)
      setProviderModalOpen(false)
    } catch (err) {
      showError(err instanceof Error ? err.message : '保存失败')
    }
  }

  async function loadProviderKeys(providerId: number) {
    const data = await api.providerKeys(providerId)
    setProviderKeys(data.items || [])
    setSelectedProviderId(providerId)
  }

  async function createProviderKey() {
    if (!selectedProviderId) return
    try {
      const created = await api.createProviderKey(selectedProviderId, { ...keyForm, priority: providerKeys.length + 1 })
      if (created?.id) {
        setTestingKeys((prev) => new Set(prev).add(created.id))
        try {
          const result = await api.testProviderKey(selectedProviderId, created.id)
          if (result?.success) {
            showSuccess(`创建成功，已使用该 Key 完成真实接口调用校验，鉴权为 ${result?.auth_type || 'x_api_key'}`)
          } else {
            showError(result?.error || '创建成功，但该 Key 无法完成真实接口调用校验')
          }
        } finally {
          setTestingKeys((prev) => {
            const next = new Set(prev)
            next.delete(created.id)
            return next
          })
        }
      } else {
        showSuccess('创建成功')
      }
      setKeyForm({ secret: '', priority: 1, remark: '' })
      await loadProviderKeys(selectedProviderId)
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建失败')
    }
  }

  function reorderKeys(items: any[], fromId: number, toId: number) {
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

  async function persistKeyOrder(items: any[]) {
    if (!selectedProviderId) return
    setSavingKeyOrder(true)
    try {
      await Promise.all(items.map((item, index) => api.updateProviderKeyPriority(selectedProviderId, item.id, index + 1)))
      showSuccess('排序已更新')
      await loadProviderKeys(selectedProviderId)
    } catch (err) {
      showError(err instanceof Error ? err.message : '排序保存失败')
      await loadProviderKeys(selectedProviderId)
    } finally {
      setSavingKeyOrder(false)
    }
  }

  function handleKeyDragStart(keyId: number) {
    setDraggingKeyId(keyId)
    setDragOverKeyId(keyId)
  }

  function handleKeyDragOver(event: DragEvent<HTMLDivElement>, keyId: number) {
    event.preventDefault()
    if (draggingKeyId === null || draggingKeyId === keyId) return
    setDragOverKeyId(keyId)
    setProviderKeys((current) => reorderKeys(current, draggingKeyId, keyId))
    setDraggingKeyId(keyId)
  }

  async function handleKeyDrop() {
    setDragOverKeyId(null)
    if (draggingKeyId === null) return
    setDraggingKeyId(null)
  }

  function handleKeyDragEnd() {
    setDraggingKeyId(null)
    setDragOverKeyId(null)
  }

  function startEdit(item: any) {
    setEditingProviderId(item.id)
    setForm({
      name: item.name,
      provider_type: item.provider_type,
      service_type: item.service_type,
      base_url: item.base_url,
      auth_type: item.auth_type,
      timeout_ms: item.timeout_ms,
      enabled: item.enabled,
      health_status: item.health_status,
      remark: item.remark || '',
      new_key: '',
    })
    setProviderModalOpen(true)
  }

  function closeProviderModal() {
    setProviderModalOpen(false)
    setEditingProviderId(null)
    setForm(initialProviderForm)
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="总数" value={stats.total} icon={<PiPlugsConnectedBold className="h-4 w-4" />} />
        <StatCard title="可正常使用" value={stats.usable} icon={<GoDotFill className="h-4 w-4 text-emerald-400" />} />
        <StatCard title="已禁用" value={stats.disabled} icon={<GoDotFill className="h-4 w-4 text-slate-400" />} />
        <StatCard title="无法使用" value={stats.unavailable} icon={<GoDotFill className="h-4 w-4 text-rose-400" />} />
      </div>

      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
          <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>上游渠道列表</div>
          <div className="flex items-center gap-2">
            <Button onClick={() => { setEditingProviderId(null); setForm(initialProviderForm); setProviderModalOpen(true) }} size="sm"><RiAddLine className="h-4 w-4" /></Button>
          </div>
        </div>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>状态</TableHead>
                <TableHead>名称</TableHead>
                <TableHead>备注</TableHead>
                <TableHead>鉴权</TableHead>
                <TableHead>启用</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loadingData ? (
                <TableRow>
                  <TableCell colSpan={6} className="h-32 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>
                    <div className="flex items-center justify-center gap-2">
                      <div className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600" />
                      加载中...
                    </div>
                  </TableCell>
                </TableRow>
              ) : providers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="h-32 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</TableCell>
                </TableRow>
              ) : providers.map((item) => {
                const isUsable = item.enabled && item.health_status === 'healthy'
                const statusText = isUsable ? '可正常使用' : '无法使用'
                const statusColor = isUsable ? 'text-emerald-400' : 'text-rose-400'
                return (
                  <TableRow key={item.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <GoDotFill className={`h-3 w-3 ${statusColor}`} />
                        <span className={statusColor}>{statusText}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col items-center text-center">
                        <div className="font-medium" style={{ color: 'var(--foreground)' }}>{item.name}</div>
                        <div className="max-w-[240px] truncate text-xs text-center" style={{ color: 'var(--muted-foreground)' }}>{item.base_url}</div>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm" style={{ color: 'var(--muted-foreground)' }}>{item.remark || '-'}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1.5"><HiOutlineKey className="h-3.5 w-3.5 text-slate-500" />{item.auth_type}</div>
                    </TableCell>
                    <TableCell><span style={{ color: item.enabled ? 'var(--success)' : 'var(--danger)' }}>{item.enabled ? '是' : '否'}</span></TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-2">
                        <Button variant="secondary" size="sm" loading={testingProviders.has(item.id)} onClick={() => { setTestingProviders(prev => new Set(prev).add(item.id)); api.testProvider(item.id).then((result: any) => { if (result?.success) { showSuccess(`检测完成，鉴权为 ${result?.auth_type || 'x_api_key'}`) } else { showError(result?.message || '检测失败') } return load() }).catch((err) => showError(err.message)).finally(() => setTestingProviders(prev => { const next = new Set(prev); next.delete(item.id); return next })) }}>测试</Button>
                        <Button variant="secondary" size="sm" onClick={() => { loadProviderKeys(item.id).catch((err) => setMessage(err.message)); setProviderKeyModalOpen(true) }}>Keys</Button>
                        <Button variant="secondary" size="sm" onClick={() => startEdit(item)}>编辑</Button>
                        <Button variant="secondary" size="sm" loading={loadingStatus === item.id} onClick={() => { setLoadingStatus(item.id); api.updateProviderStatus(item.id, !item.enabled).then(() => { showSuccess('状态更新成功'); return load() }).catch((err) => showError(err.message)).finally(() => setLoadingStatus(null)) }} style={{ color: item.enabled ? 'var(--danger)' : 'var(--success)' }}>{item.enabled ? '禁用' : '启用'}</Button>
                        <Button variant="destructive" size="sm" onClick={() => setPendingDeleteProvider(item)}>删除</Button>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </CardContent>
        <PaginationBar page={page} pageSize={pageSize} total={total} onPageChange={setPage} onPageSizeChange={setPageSize} />
      </Card>

      {message ? <Card><CardContent className="p-4 text-sm" style={{ color: 'var(--muted-foreground)' }}>{message}</CardContent></Card> : null}
      
      <Modal
        open={providerModalOpen}
        title={editingProviderId ? '编辑上游' : '新增上游'}
        onClose={closeProviderModal}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={closeProviderModal}>取消</Button>
            <Button loading={loadingSubmit} onClick={async () => { setLoadingSubmit(true); try { await submitProvider() } finally { setLoadingSubmit(false) } }}>保存</Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>名称</label>
            <Input className="flex-1" placeholder="请输入名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>Base URL</label>
            <Input className="flex-1" placeholder="请输入 Base URL" value={form.base_url} onChange={(e) => setForm({ ...form, base_url: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>鉴权</label>
            <div className="flex-1 space-y-2">
              <Input value={editingProviderId ? form.auth_type || 'x_api_key' : '自动识别（默认 x_api_key，失败后尝试 bearer）'} readOnly />
              <Input type="number" placeholder="超时(ms)" value={form.timeout_ms} onChange={(e) => setForm({ ...form, timeout_ms: Number(e.target.value) })} />
            </div>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>API Key</label>
            <Input className="flex-1" placeholder="请输入 API Key（必填）" value={form.new_key} onChange={(e) => setForm({ ...form, new_key: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>备注</label>
            <Textarea className="flex-1" placeholder="请输入备注" value={form.remark} onChange={(e) => setForm({ ...form, remark: e.target.value })} />
          </div>
        </div>
      </Modal>
      <Modal
        open={providerKeyModalOpen}
        title="Provider Keys"
        maxWidthClass="max-w-5xl"
        onClose={() => { setProviderKeyModalOpen(false); setKeyForm({ secret: '', priority: 1, remark: '' }) }}
        footer={
          <div className="flex justify-between gap-2">
            <div className="flex items-center gap-2">
              {providerKeys.length > 1 && (
                <Button variant="secondary" onClick={async () => await persistKeyOrder([...providerKeys])} disabled={savingKeyOrder} loading={savingKeyOrder}>
                  保存顺序
                </Button>
              )}
            </div>
            <div className="flex gap-2">
              <Button variant="secondary" onClick={() => { setProviderKeyModalOpen(false); setKeyForm({ secret: '', priority: 1, remark: '' }) }}>关闭</Button>
              <Button disabled={!selectedProviderId || !keyForm.secret.trim() || loadingKeySubmit} loading={loadingKeySubmit} onClick={async () => { setLoadingKeySubmit(true); try { await createProviderKey() } finally { setLoadingKeySubmit(false) } }}>新增 Key</Button>
            </div>
          </div>
        }
      >
        {selectedProviderId ? (
          <div className="space-y-4">
            <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>Secret</label>
            <Input className="flex-1" placeholder="请输入 Secret" value={keyForm.secret} onChange={(e) => setKeyForm({ ...keyForm, secret: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>顺序</label>
            <Input className="flex-1" value={`自动追加到第 ${providerKeys.length + 1} 位`} readOnly />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>备注</label>
            <Textarea className="flex-1" placeholder="请输入备注" value={keyForm.remark} onChange={(e) => setKeyForm({ ...keyForm, remark: e.target.value })} />
              </div>
            </div>
            <div className="space-y-2 rounded-[10px] border p-4 text-sm" style={{ borderColor: 'var(--border)', background: 'rgba(255,255,255,0.03)' }}>
              {providerKeys.length ? <AnimatePresence initial={false}>{providerKeys.map((item) => (
                <motion.div
                  key={item.id}
                  layout
                  transition={{ type: 'spring', stiffness: 380, damping: 30 }}
                  draggable={!savingKeyOrder}
                  onDragStart={() => handleKeyDragStart(item.id)}
                  onDragOver={(event) => handleKeyDragOver(event, item.id)}
                  onDrop={() => { void handleKeyDrop() }}
                  onDragEnd={handleKeyDragEnd}
                  className="flex items-center justify-between gap-3 rounded-lg border p-3 transition-colors"
                  style={{
                    borderColor: dragOverKeyId === item.id ? 'var(--primary)' : 'var(--border)',
                    opacity: draggingKeyId === item.id ? 0.75 : 1,
                    background: dragOverKeyId === item.id ? 'rgba(59,130,246,0.08)' : 'transparent',
                  }}
                >
                  <div className="min-w-0 flex items-center gap-3 flex-1">
                    <RiDraggable className="cursor-move text-slate-400" />
                    <div className="min-w-0 flex flex-col gap-1 flex-1">
                      <div className="font-mono text-xs break-all" style={{ color: 'var(--foreground)' }}>{item.plain_key || item.key_masked}</div>
                      <div className="flex items-center gap-2 text-xs" style={{ color: 'var(--muted-foreground)' }}>
                        <span>{item.status === 'enabled' ? '正常' : '禁用'}</span>
                        {item.fail_count ? <span className="text-rose-400">/ 失败 {item.fail_count}</span> : null}
                        {item.remark && <span>/ {item.remark}</span>}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button variant="secondary" size="sm" loading={testingKeys.has(item.id)} onClick={() => { setTestingKeys(prev => new Set(prev).add(item.id)); api.testProviderKey(selectedProviderId!, item.id).then((res: any) => { if (res?.success) { showSuccess('连接成功') } else { showError(`连接失败: ${res?.error || '未知错误'}`) } return loadProviderKeys(selectedProviderId!) }).catch((err) => showError(err.message)).finally(() => setTestingKeys(prev => { const next = new Set(prev); next.delete(item.id); return next })) }}>测试</Button>
                    <Button variant="secondary" size="sm" onClick={() => { setEditingKeyId(item.id); setEditingKeyForm({ secret: item.plain_key || '', priority: item.priority, remark: item.remark || '' }) }}><FiEdit2 className="h-3 w-3" /></Button>
                    <Button variant="secondary" size="sm" onClick={() => api.updateProviderKeyStatus(selectedProviderId!, item.id, item.status === 'enabled' ? 'disabled' : 'enabled').then(() => { showSuccess('状态更新成功'); return loadProviderKeys(selectedProviderId!) }).catch((err) => showError(err.message))}>{item.status === 'enabled' ? '禁用' : '启用'}</Button>
                    <Button variant="destructive" size="sm" onClick={() => setPendingDeleteProviderKey(item)}><FiTrash2 className="h-3 w-3" /></Button>
                  </div>
                </motion.div>
              ))}</AnimatePresence> : <div style={{ color: 'var(--muted-foreground)' }}>暂无</div>}
            </div>
            {providerKeys.length > 1 ? <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>拖动左侧图标即可调整顺序，点击"保存顺序"按钮后生效。</div> : null}
          </div>
        ) : <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>请先选择上游</div>}
      </Modal>
      <Modal
        open={editingKeyId !== null}
        title="编辑 Key"
        onClose={() => { setEditingKeyId(null); setEditingKeyForm({ secret: '', priority: 1, remark: '' }) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setEditingKeyId(null); setEditingKeyForm({ secret: '', priority: 1, remark: '' }) }}>取消</Button>
            <Button loading={loadingKeySubmit} onClick={async () => { if (!selectedProviderId || !editingKeyId) return; setLoadingKeySubmit(true); try { await api.updateProviderKey(selectedProviderId, editingKeyId, editingKeyForm); setTestingKeys((prev) => new Set(prev).add(editingKeyId)); try { const result = await api.testProviderKey(selectedProviderId, editingKeyId); if (result?.success) { showSuccess(`更新成功，Key 测试通过，鉴权为 ${result?.auth_type || 'x_api_key'}`) } else { showError(result?.error || '更新成功，但 Key 测试失败') } } finally { setTestingKeys((prev) => { const next = new Set(prev); next.delete(editingKeyId); return next }) } setEditingKeyId(null); await loadProviderKeys(selectedProviderId); await load() } catch (err) { showError(err instanceof Error ? err.message : '更新失败') } finally { setLoadingKeySubmit(false) } }}>保存</Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>Secret</label>
            <Input className="flex-1" placeholder="请输入 Secret" value={editingKeyForm.secret} onChange={(e) => setEditingKeyForm({ ...editingKeyForm, secret: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>优先级</label>
            <Input type="number" className="flex-1" placeholder="优先级" value={editingKeyForm.priority} onChange={(e) => setEditingKeyForm({ ...editingKeyForm, priority: Number(e.target.value) })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>备注</label>
            <Textarea className="flex-1" placeholder="请输入备注" value={editingKeyForm.remark} onChange={(e) => setEditingKeyForm({ ...editingKeyForm, remark: e.target.value })} />
          </div>
        </div>
      </Modal>
      <ConfirmDialog
        open={Boolean(pendingDeleteProvider)}
        title="删除确认"
        description={pendingDeleteProvider ? `确认删除上游「${pendingDeleteProvider.name}」吗？` : ''}
        confirmText="确认删除"
        confirmVariant="destructive"
        onCancel={() => setPendingDeleteProvider(null)}
        onConfirm={() => {
          if (!pendingDeleteProvider) return
          api.deleteProvider(pendingDeleteProvider.id)
            .then(async () => {
              if (selectedProviderId === pendingDeleteProvider.id) {
                setSelectedProviderId(null)
                setProviderKeys([])
                setProviderKeyModalOpen(false)
              }
              setPendingDeleteProvider(null)
              showSuccess('删除成功')
              await load()
            })
            .catch((err) => showError(err.message))
        }}
      />
      <ConfirmDialog
        open={Boolean(pendingDeleteProviderKey)}
        title="删除确认"
        description={pendingDeleteProviderKey ? `确认删除 Key「${pendingDeleteProviderKey.key_masked}」吗？` : ''}
        confirmText="确认删除"
        confirmVariant="destructive"
        onCancel={() => setPendingDeleteProviderKey(null)}
        onConfirm={() => {
          if (!pendingDeleteProviderKey || !selectedProviderId) return
          api.deleteProviderKey(selectedProviderId, pendingDeleteProviderKey.id).then(async () => {
            setPendingDeleteProviderKey(null)
            showSuccess('删除成功')
            await loadProviderKeys(selectedProviderId)
          }).catch((err) => showError(err.message))
        }}
      />
    </div>
  )
}
