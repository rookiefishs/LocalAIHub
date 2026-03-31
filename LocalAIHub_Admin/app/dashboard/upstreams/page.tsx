'use client'

import { useEffect, useMemo, useState } from 'react'
import { FiEdit2, FiTrash2 } from 'react-icons/fi'
import { GoDotFill } from 'react-icons/go'
import { HiOutlineBolt, HiOutlineKey } from 'react-icons/hi2'
import { PiPlugsConnectedBold } from 'react-icons/pi'
import { RiAddLine } from 'react-icons/ri'
import { api } from '@/lib/api'
import { StatCard } from '@/components/stat-card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { PaginationBar } from '@/components/pagination-bar'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useToast } from '@/components/ui/toast'
import { useRefresh } from '@/components/refresh-context'

const initialProviderForm = {
  name: '',
  provider_type: 'proxy',
  service_type: 'proxy',
  base_url: '',
  auth_type: 'x_api_key',
  timeout_ms: 60000,
  enabled: true,
  health_status: 'healthy',
  remark: '',
  newKey: '',
}

export default function UpstreamsPage() {
  const [providers, setProviders] = useState<any[]>([])
  const [providerKeys, setProviderKeys] = useState<any[]>([])
  const [selectedProviderId, setSelectedProviderId] = useState<number | null>(null)
  const [message, setMessage] = useState('')
  const [editingProviderId, setEditingProviderId] = useState<number | null>(null)
  const [providerModalOpen, setProviderModalOpen] = useState(false)
  const [providerKeyModalOpen, setProviderKeyModalOpen] = useState(false)
  const [pendingDeleteProvider, setPendingDeleteProvider] = useState<any | null>(null)
  const [pendingDeleteProviderKey, setPendingDeleteProviderKey] = useState<any | null>(null)
  const [form, setForm] = useState(initialProviderForm)
  const [keyForm, setKeyForm] = useState({ secret: '', priority: 1, remark: '' })
  const [loadingTest, setLoadingTest] = useState<number | null>(null)
  const [loadingStatus, setLoadingStatus] = useState<number | null>(null)
  const [loadingSubmit, setLoadingSubmit] = useState(false)
  const [loadingKeySubmit, setLoadingKeySubmit] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const { showSuccess, showError } = useToast()
  const { registerRefresh } = useRefresh()

  async function load() {
    const data = await api.providers(`page=${page}&page_size=${pageSize}`)
    setProviders(data.items || [])
    setTotal(data.total || 0)
  }

  useEffect(() => {
    registerRefresh(load)
    load().catch((err) => setMessage(err.message))
  }, [page, pageSize])

  useEffect(() => {
    setPage(1)
  }, [pageSize])

  const stats = useMemo(() => {
    const healthy = providers.filter((item) => item.health_status === 'healthy').length
    const degraded = providers.filter((item) => item.health_status === 'degraded').length
    const disabled = providers.filter((item) => !item.enabled).length
    return { total: providers.length, healthy, degraded, disabled }
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
    if (!editingProviderId && !form.newKey.trim()) {
      showError('请输入 API Key')
      return
    }
    try {
      if (editingProviderId) {
        await api.updateProvider(editingProviderId, form)
        showSuccess('保存成功')
      } else {
        await api.createProvider(form)
        showSuccess('创建成功')
      }
      setForm(initialProviderForm)
      setEditingProviderId(null)
      setProviderModalOpen(false)
      await load()
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
      await api.createProviderKey(selectedProviderId, keyForm)
      setKeyForm({ secret: '', priority: 1, remark: '' })
      setProviderKeyModalOpen(false)
      showSuccess('创建成功')
      await loadProviderKeys(selectedProviderId)
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建失败')
    }
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
      newKey: '',
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
        <StatCard title="正常" value={stats.healthy} icon={<GoDotFill className="h-4 w-4 text-emerald-400" />} />
        <StatCard title="降级" value={stats.degraded} icon={<GoDotFill className="h-4 w-4 text-yellow-400" />} />
        <StatCard title="停用" value={stats.disabled} icon={<GoDotFill className="h-4 w-4 text-rose-400" />} />
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
              {providers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-32 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</TableCell>
                </TableRow>
              ) : providers.map((item) => {
                const statusColor = item.health_status === 'healthy' ? 'text-emerald-400' : item.health_status === 'degraded' ? 'text-yellow-400' : 'text-rose-400'
                return (
                  <TableRow key={item.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <GoDotFill className={`h-3 w-3 ${statusColor}`} />
                        <span>{item.health_status}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="font-medium" style={{ color: 'var(--foreground)' }}>{item.name}</div>
                      <div className="max-w-[240px] truncate text-xs" style={{ color: 'var(--muted-foreground)' }}>{item.base_url}</div>
                    </TableCell>
                    <TableCell className="text-sm" style={{ color: 'var(--muted-foreground)' }}>{item.remark || '-'}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1.5"><HiOutlineKey className="h-3.5 w-3.5 text-slate-500" />{item.auth_type}</div>
                    </TableCell>
                    <TableCell><span style={{ color: item.enabled ? 'var(--success)' : 'var(--danger)' }}>{item.enabled ? '是' : '否'}</span></TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-2">
                        <Button variant="secondary" size="sm" loading={loadingTest === item.id} onClick={() => { setLoadingTest(item.id); api.testProvider(item.id).then(() => { showSuccess('检测完成'); return load() }).catch((err) => showError(err.message)).finally(() => setLoadingTest(null)) }}>测试</Button>
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
            <div className="flex flex-1 gap-3">
              <Select value={form.auth_type} onValueChange={(value) => setForm({ ...form, auth_type: value })}>
                <SelectTrigger className="flex-1"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="bearer">bearer</SelectItem>
                  <SelectItem value="x_api_key">x_api_key</SelectItem>
                </SelectContent>
              </Select>
              <Input type="number" className="flex-1" placeholder="超时(ms)" value={form.timeout_ms} onChange={(e) => setForm({ ...form, timeout_ms: Number(e.target.value) })} />
            </div>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>API Key</label>
            <Input className="flex-1" placeholder="请输入 API Key（必填）" value={form.newKey} onChange={(e) => setForm({ ...form, newKey: e.target.value })} />
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
        onClose={() => { setProviderKeyModalOpen(false); setKeyForm({ secret: '', priority: 1, remark: '' }) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setProviderKeyModalOpen(false); setKeyForm({ secret: '', priority: 1, remark: '' }) }}>关闭</Button>
            <Button disabled={!selectedProviderId || !keyForm.secret.trim() || loadingKeySubmit} loading={loadingKeySubmit} onClick={async () => { setLoadingKeySubmit(true); try { await createProviderKey() } finally { setLoadingKeySubmit(false) } }}>新增 Key</Button>
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
                <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>优先级</label>
                <Input type="number" className="flex-1" placeholder="优先级" value={keyForm.priority} onChange={(e) => setKeyForm({ ...keyForm, priority: Number(e.target.value) })} />
              </div>
              <div className="flex items-center gap-4">
                <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>备注</label>
                <Textarea className="flex-1" placeholder="请输入备注" value={keyForm.remark} onChange={(e) => setKeyForm({ ...keyForm, remark: e.target.value })} />
              </div>
            </div>
            <div className="space-y-2 rounded-2xl border p-4 text-sm" style={{ borderColor: 'var(--border)', background: 'rgba(255,255,255,0.03)' }}>
              {providerKeys.length ? providerKeys.map((item) => (
                <div key={item.id} className="flex items-center justify-between gap-3">
                  <div>{item.key_masked}</div>
                  <div className="flex items-center gap-2">
                    <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>
                      {item.status} / p{item.priority}
                      {item.fail_count ? ` / fail ${item.fail_count}` : ''}
                    </div>
                    <Button variant="secondary" size="sm" onClick={() => api.updateProviderKeyStatus(selectedProviderId, item.id, item.status === 'enabled' ? 'disabled' : 'enabled').then(() => { showSuccess('状态更新成功'); return loadProviderKeys(selectedProviderId) }).catch((err) => showError(err.message))}>{item.status === 'enabled' ? '禁用' : '启用'}</Button>
                    <Button variant="destructive" size="sm" onClick={() => setPendingDeleteProviderKey(item)}>删除</Button>
                  </div>
                </div>
              )) : <div style={{ color: 'var(--muted-foreground)' }}>暂无</div>}
            </div>
          </div>
        ) : <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>请先选择上游</div>}
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
