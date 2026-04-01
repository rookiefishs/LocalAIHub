'use client'

import { useEffect, useMemo, useState } from 'react'
import { GoDotFill } from 'react-icons/go'
import { HiOutlineKey } from 'react-icons/hi2'
import { LuActivity } from 'react-icons/lu'
import { RiAddLine, RiClipboardLine, RiEditLine, RiEyeLine, RiEyeOffLine } from 'react-icons/ri'
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
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useToast } from '@/components/ui/toast'
import { useRefresh } from '@/components/refresh-context'

export default function KeysPage() {
  const [items, setItems] = useState<any[]>([])
  const [message, setMessage] = useState('')
  const [keyModalOpen, setKeyModalOpen] = useState(false)
  const [editingKey, setEditingKey] = useState<any | null>(null)
  const [pendingDeleteKey, setPendingDeleteKey] = useState<any | null>(null)
  const today = new Date().toISOString().split('T')[0]
  const defaultForm = { name: '', remark: '', expires_at: today, allowed_models: [] as number[] }
  const [form, setForm] = useState(defaultForm)
  const [showKeyId, setShowKeyId] = useState<number | null>(null)
  const [fullKeyMap, setFullKeyMap] = useState<Record<number, string>>({})
  const [loadingStatus, setLoadingStatus] = useState<number | null>(null)
  const [loadingTest, setLoadingTest] = useState<number | null>(null)
  const [loadingSubmit, setLoadingSubmit] = useState(false)
  const [testingKeys, setTestingKeys] = useState<Set<number>>(new Set())
  const [useKeyModalOpen, setUseKeyModalOpen] = useState(false)
  const [selectedKeyForUse, setSelectedKeyForUse] = useState<any | null>(null)
  const [models, setModels] = useState<any[]>([])
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const { showSuccess, showError } = useToast()
  const { registerRefresh } = useRefresh()

  async function load() {
    const [keysData, modelsData] = await Promise.all([
      api.clientKeys(`page=${page}&page_size=${pageSize}`),
      api.models()
    ])
    setItems(keysData.items || [])
    setTotal(keysData.total || 0)
    setModels(modelsData.items || [])
  }

  useEffect(() => { registerRefresh(load); load().catch((err) => setMessage(err.message)) }, [page, pageSize])

  const stats = useMemo(() => ({
    total: items.length,
    active: items.filter((item) => item.status === 'active').length,
    disabled: items.filter((item) => item.status !== 'active').length,
  }), [items])

  async function createKey() {
    if (!form.name.trim()) {
      showError('请输入名称')
      return
    }
    try {
      const created = await api.createClientKey({
        name: form.name,
        remark: form.remark,
        expires_at: form.expires_at === today ? '' : form.expires_at,
        allowed_models: form.allowed_models
      })
      if (created.status === 'disabled') {
        showError('创建成功，但联通性检测失败，已自动禁用')
      } else {
        showSuccess('创建成功')
      }
      setForm(defaultForm)
      setKeyModalOpen(false)
      await load()
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建失败')
    }
  }

  async function updateKey() {
    if (!form.name.trim() || !editingKey) {
      showError('请输入名称')
      return
    }
    try {
      await api.updateClientKey(editingKey.id, form)
      showSuccess('保存成功')
      setEditingKey(null)
      setForm(defaultForm)
      setKeyModalOpen(false)
      await load()
    } catch (err) {
      showError(err instanceof Error ? err.message : '保存失败')
    }
  }

  async function testClientKey(id: number, currentStatus: string) {
    try {
      const result = await api.testClientKey(id)
      showSuccess(`测试成功：${result.model}`)
    } catch (err) {
      showError(err instanceof Error ? err.message : '测试失败')
      if (currentStatus === 'active') {
        try {
          await api.updateClientKeyStatus(id, 'disabled')
          showSuccess('测试失败，已自动禁用')
          await load()
        } catch (updateErr) {
          console.error('自动禁用失败', updateErr)
        }
      }
    }
  }

  function openCreateModal() {
    setEditingKey(null)
    setForm(defaultForm)
    setKeyModalOpen(true)
    setShowKeyId(null)
    setFullKeyMap({})
  }

  async function openEditModal(item: any) {
    setLoadingSubmit(true)
    try {
      const keyData = await api.getClientKey(item.id)
      setEditingKey(keyData)
      setForm({ 
        name: keyData.name, 
        remark: keyData.remark || '', 
        expires_at: keyData.expires_at ? keyData.expires_at.split('T')[0] : '', 
        allowed_models: keyData.allowed_models || [] 
      })
    } catch (err) {
      setEditingKey(item)
      setForm({ 
        name: item.name, 
        remark: item.remark || '', 
        expires_at: item.expires_at ? item.expires_at.split('T')[0] : '', 
        allowed_models: [] 
      })
    } finally {
      setLoadingSubmit(false)
    }
    setKeyModalOpen(true)
    setShowKeyId(null)
    setFullKeyMap({})
  }

  function closeModal() {
    setKeyModalOpen(false)
    setEditingKey(null)
    setForm(defaultForm)
    setShowKeyId(null)
    setFullKeyMap({})
  }

  async function copyFullKey(item: any) {
    if (fullKeyMap[item.id]) {
      navigator.clipboard.writeText(fullKeyMap[item.id])
      showSuccess('已复制')
      return
    }
    try {
      const data = await api.getClientKey(item.id)
      const key = data.plain_key || item.key_prefix + '****'
      setFullKeyMap(prev => ({ ...prev, [item.id]: key }))
      navigator.clipboard.writeText(key)
      showSuccess('已复制')
    } catch (err) {
      const key = item.key_prefix + '****'
      navigator.clipboard.writeText(key)
      showSuccess('已复制')
    }
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="总数" value={stats.total} icon={<HiOutlineKey className="h-4 w-4" />} />
        <StatCard title="启用" value={stats.active} icon={<GoDotFill className="h-4 w-4 text-emerald-400" />} />
        <StatCard title="禁用" value={stats.disabled} icon={<GoDotFill className="h-4 w-4 text-rose-400" />} />
        <StatCard title="可用" value={stats.active} icon={<LuActivity className="h-4 w-4" />} />
      </div>

      <Card className="overflow-hidden">
          <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
            <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>API Key 列表</div>
            <Button onClick={openCreateModal} size="sm"><RiAddLine className="h-4 w-4" /></Button>
          </div>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>API Key</TableHead>
                <TableHead>请求次数</TableHead>
                <TableHead>Token 消耗</TableHead>
                <TableHead>备注</TableHead>
                <TableHead>到期时间</TableHead>
                <TableHead>最后使用</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="h-32 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</TableCell>
                </TableRow>
              ) : items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <div className="font-medium" style={{ color: 'var(--foreground)' }}>{item.name}</div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs" style={{ color: 'var(--muted-foreground)' }}>
                        {showKeyId === item.id && fullKeyMap[item.id] ? fullKeyMap[item.id] : item.key_prefix + '****'}
                      </span>
                      <Button variant="secondary" size="sm" className="h-6 w-6 p-0" onClick={() => copyFullKey(item)}>
                        <RiClipboardLine className="h-3 w-3" />
                      </Button>
                      <Button variant="secondary" size="sm" className="h-6 w-6 p-0" onClick={() => {
                        if (showKeyId === item.id) {
                          setShowKeyId(null)
                        } else {
                          copyFullKey(item)
                          setShowKeyId(item.id)
                        }
                      }}>
                        {showKeyId === item.id ? <RiEyeOffLine className="h-3 w-3" /> : <RiEyeLine className="h-3 w-3" />}
                      </Button>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm" style={{ color: 'var(--foreground)' }}>{item.request_count || 0}</TableCell>
                  <TableCell className="text-sm" style={{ color: 'var(--foreground)' }}>{(item.total_tokens || 0).toLocaleString()}</TableCell>
                  <TableCell className="text-sm" style={{ color: 'var(--muted-foreground)' }}>{item.remark || '-'}</TableCell>
                  <TableCell className="text-sm" style={{ color: item.expires_at && new Date(item.expires_at) < new Date() ? 'var(--danger)' : 'var(--muted-foreground)' }}>
                    {item.expires_at ? new Date(item.expires_at).toLocaleDateString() : '永久'}
                  </TableCell>
                  <TableCell>{item.last_used_at ? new Date(item.last_used_at).toLocaleString() : '-'}</TableCell>
                  <TableCell><span style={{ color: item.status === 'active' ? 'var(--success)' : 'var(--danger)' }}>{item.status === 'active' ? '启用' : '禁用'}</span></TableCell>
                    <TableCell>
                    <div className="flex gap-2">
                      <Button variant="secondary" size="sm" loading={testingKeys.has(item.id)} onClick={() => { setTestingKeys(prev => new Set(prev).add(item.id)); testClientKey(item.id, item.status).finally(() => setTestingKeys(prev => { const next = new Set(prev); next.delete(item.id); return next })) }}>测试</Button>
                      <Button variant="secondary" size="sm" onClick={async () => { const keyData = await api.getClientKey(item.id); setSelectedKeyForUse(keyData); setUseKeyModalOpen(true) }}>使用密钥</Button>
                      <Button variant="secondary" size="sm" onClick={() => openEditModal(item)}>编辑</Button>
                      <Button variant="secondary" size="sm" loading={loadingStatus === item.id} onClick={() => { setLoadingStatus(item.id); api.updateClientKeyStatus(item.id, item.status === 'active' ? 'disabled' : 'active').then(() => { showSuccess('状态更新成功'); return load() }).catch((err) => showError(err.message)).finally(() => setLoadingStatus(null)) }} style={{ color: item.status === 'active' ? 'var(--danger)' : 'var(--success)' }}>{item.status === 'active' ? '禁用' : '启用'}</Button>
                      <Button variant="destructive" size="sm" onClick={() => setPendingDeleteKey(item)}>删除</Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
        <PaginationBar page={page} pageSize={pageSize} total={total} onPageChange={setPage} onPageSizeChange={setPageSize} />
      </Card>

      {message ? <Card><CardContent className="p-4 text-sm" style={{ color: 'var(--muted-foreground)' }}>{message}</CardContent></Card> : null}
      
      <Modal
        open={keyModalOpen}
        title={editingKey ? '编辑 API Key' : '创建 API Key'}
        onClose={closeModal}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={closeModal}>{editingKey ? '关闭' : '取消'}</Button>
            <Button loading={loadingSubmit} onClick={async () => { setLoadingSubmit(true); try { if (editingKey) { await updateKey() } else { await createKey() } } finally { setLoadingSubmit(false) } }}>{editingKey ? '保存' : '创建'}</Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>名称</label>
            <Input className="flex-1" placeholder="请输入名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>过期时间</label>
            <div className="flex-1 flex gap-2">
              <Input type="date" className="flex-1" value={form.expires_at} onChange={(e) => setForm({ ...form, expires_at: e.target.value })} />
              <Select value="" onValueChange={(value) => {
                if (!value) return
                const now = new Date()
                if (value === 'permanent') {
                  setForm({ ...form, expires_at: '' })
                } else {
                  const days = parseInt(value)
                  now.setDate(now.getDate() + days)
                  setForm({ ...form, expires_at: now.toISOString().split('T')[0] })
                }
              }}>
                <SelectTrigger className="w-28"><SelectValue placeholder="快捷选择" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="permanent">永久</SelectItem>
                  <SelectItem value="1">1天</SelectItem>
                  <SelectItem value="7">7天</SelectItem>
                  <SelectItem value="30">1个月</SelectItem>
                  <SelectItem value="90">3个月</SelectItem>
                  <SelectItem value="180">6个月</SelectItem>
                  <SelectItem value="365">1年</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>可用模型</label>
            <div className="flex-1">
              <Select value={form.allowed_models[0] ? String(form.allowed_models[0]) : 'all'} onValueChange={(value) => setForm({ ...form, allowed_models: value === 'all' ? [] : [Number(value)] })}>
                <SelectTrigger className="flex-1"><SelectValue placeholder="选择可用模型（单选）" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">所有模型</SelectItem>
                  {models.map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>{item.model_code}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>备注</label>
            <Textarea className="flex-1" placeholder="请输入备注" value={form.remark} onChange={(e) => setForm({ ...form, remark: e.target.value })} />
          </div>
        </div>
      </Modal>
      <Modal
        open={useKeyModalOpen}
        title="使用密钥"
        onClose={() => { setUseKeyModalOpen(false); setSelectedKeyForUse(null) }}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setUseKeyModalOpen(false); setSelectedKeyForUse(null) }}>关闭</Button>
          </div>
        }
      >
        {selectedKeyForUse && (() => {
          const getProxyURL = () => {
            if (typeof window === 'undefined') return 'http://127.0.0.1:3334'
            const envBase = process.env.NEXT_PUBLIC_API_BASE_URL
            if (envBase && (envBase.startsWith('http://') || envBase.startsWith('https://'))) {
              return envBase.replace('/admin', '')
            }
            if (envBase === '/localaihub-api') {
              return window.location.origin + '/localaihub-api'
            }
            return window.location.origin
          }
          const baseURL = getProxyURL()
          const apiKey = selectedKeyForUse.plain_key || ''
          const allowedModelList = selectedKeyForUse.allowed_models?.length ? models.filter(m => selectedKeyForUse.allowed_models.includes(m.id)) : models
          const modelName = allowedModelList[0]?.model_code || 'gpt-5.4'
          const requestURL = `${baseURL}/proxy/openai/v1/chat/completions`
          const configText = `Base URL: ${requestURL}\nAPI Key: ${apiKey}\nModel: ${modelName}`
          const curlText = `curl ${requestURL} -H "Content-Type: application/json" -H "Authorization: Bearer ${apiKey}" --data-raw "{\\"model\\":\\"${modelName}\\",\\"messages\\":[{\\"role\\":\\"user\\",\\"content\\":\\"hi\\"}],\\"max_tokens\\":5}"`
          
          return (
            <div className="space-y-4">
              <div>
                <div className="text-sm mb-1" style={{ color: 'var(--muted-foreground)' }}>配置内容</div>
                <div className="relative">
                  <pre className="p-3 rounded-[10px] border overflow-auto max-h-60 text-sm font-mono" style={{ background: 'rgba(0,0,0,0.2)', borderColor: 'var(--border)', color: 'var(--foreground)' }}>{configText}</pre>
                  <Button variant="secondary" size="sm" className="absolute top-2 right-2" onClick={() => { navigator.clipboard.writeText(configText); showSuccess('已复制') }}>复制</Button>
                </div>
              </div>
              <div>
                <div className="text-sm mb-1" style={{ color: 'var(--muted-foreground)' }}>Windows CMD 示例</div>
                <div className="relative">
                  <pre className="p-3 rounded-[10px] border overflow-auto max-h-60 text-sm font-mono whitespace-pre-wrap break-all" style={{ background: 'rgba(0,0,0,0.2)', borderColor: 'var(--border)', color: 'var(--foreground)' }}>{curlText}</pre>
                  <Button variant="secondary" size="sm" className="absolute top-2 right-2" onClick={() => { navigator.clipboard.writeText(curlText); showSuccess('已复制命令') }}>复制命令</Button>
                </div>
              </div>
            </div>
          )
        })()}
      </Modal>
      <ConfirmDialog
        open={Boolean(pendingDeleteKey)}
        title="删除确认"
        description={pendingDeleteKey ? `确认删除 Key「${pendingDeleteKey.name}」吗？` : ''}
        confirmText="确认删除"
        confirmVariant="destructive"
        onCancel={() => setPendingDeleteKey(null)}
        onConfirm={() => {
          if (!pendingDeleteKey) return
          api.deleteClientKey(pendingDeleteKey.id).then(async () => {
            setPendingDeleteKey(null)
            showSuccess('删除成功')
            await load()
          }).catch((err) => showError(err.message))
        }}
      />
    </div>
  )
}
