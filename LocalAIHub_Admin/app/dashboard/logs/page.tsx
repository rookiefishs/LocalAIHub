'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { FiDownload, FiExternalLink } from 'react-icons/fi'
import { LuFileText } from 'react-icons/lu'
import { MdOutlineHistory } from 'react-icons/md'
import { PiClockCountdownBold } from 'react-icons/pi'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { PaginationBar } from '@/components/pagination-bar'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useToast } from '@/components/ui/toast'
import { useRefresh } from '@/components/refresh-context'

interface ClientKey {
  id: number
  name: string
  key_prefix: string
  status: string
}

interface ModelItem {
  id: number
  model_code: string
  display_name: string
}

const defaultRequestFilters = { client_id: '', virtual_model_code: '', success: '', time_range: '' }
const defaultAuditFilters = { admin_user_id: '', action: '', target_type: '', target_id: '', keyword: '', start_time: '', end_time: '' }

function queryString(filters: Record<string, string>) {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([key, value]) => {
    if (value) params.set(key, value)
  })
  return params.toString()
}

function formatDateTime(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  const beijingTime = new Date(date.getTime() + 8 * 60 * 60 * 1000)
  return beijingTime.toLocaleString('zh-CN', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit', 
    hour: '2-digit', 
    minute: '2-digit', 
    second: '2-digit' 
  })
}

function parseRequestContent(requestSummary?: string): string {
  if (!requestSummary) return '-'
  try {
    const summary = JSON.parse(requestSummary)
    if (summary.messages && Array.isArray(summary.messages)) {
      return summary.messages.join('; ').substring(0, 150)
    }
    if (summary.path) {
      return summary.path
    }
    return requestSummary.substring(0, 100)
  } catch {
    return requestSummary.substring(0, 100) || '-'
  }
}

export default function LogsPage() {
  const [tab, setTab] = useState<'requests' | 'audit'>('requests')
  const [requestLogs, setRequestLogs] = useState<any[]>([])
  const [auditLogs, setAuditLogs] = useState<any[]>([])
  const [detail, setDetail] = useState<any>(null)
  const [detailType, setDetailType] = useState<'request' | 'audit'>('request')
  const [detailModalOpen, setDetailModalOpen] = useState(false)
  const [requestDraftFilters, setRequestDraftFilters] = useState(defaultRequestFilters)
  const [requestAppliedFilters, setRequestAppliedFilters] = useState(defaultRequestFilters)
  const [auditDraftFilters, setAuditDraftFilters] = useState(defaultAuditFilters)
  const [auditAppliedFilters, setAuditAppliedFilters] = useState(defaultAuditFilters)
  const [message, setMessage] = useState('')
  const [requestPage, setRequestPage] = useState(1)
  const [requestPageSize, setRequestPageSize] = useState(10)
  const [auditPage, setAuditPage] = useState(1)
  const [auditPageSize, setAuditPageSize] = useState(10)
  const [requestTotal, setRequestTotal] = useState(0)
  const [auditTotal, setAuditTotal] = useState(0)
  const [exportingAudit, setExportingAudit] = useState(false)
  const [clientKeys, setClientKeys] = useState<ClientKey[]>([])
  const [models, setModels] = useState<ModelItem[]>([])
  const { showSuccess, showError } = useToast()
  const { registerRefresh } = useRefresh()

  const requestQuery = useMemo(() => queryString(requestAppliedFilters), [requestAppliedFilters])
  const auditQuery = useMemo(() => queryString(auditAppliedFilters), [auditAppliedFilters])

  const loadRequests = useCallback(async function loadRequests(filters = requestAppliedFilters, page = requestPage, pageSize = requestPageSize) {
    const filterQuery = queryString(filters)
    const data = await api.requestLogs(`${filterQuery}${filterQuery ? '&' : ''}page_size=${pageSize}&page=${page}`)
    setRequestLogs(data.items || [])
    setRequestTotal(data.total || 0)
  }, [requestAppliedFilters, requestPage, requestPageSize])

  const loadAudit = useCallback(async function loadAudit(filters = auditAppliedFilters, page = auditPage, pageSize = auditPageSize) {
    const filterQuery = queryString(filters)
    const data = await api.auditLogs(`${filterQuery}${filterQuery ? '&' : ''}page_size=${pageSize}&page=${page}`)
    setAuditLogs(data.items || [])
    setAuditTotal(data.total || 0)
  }, [auditAppliedFilters, auditPage, auditPageSize])

  useEffect(() => {
    loadRequests(requestAppliedFilters, requestPage, requestPageSize).catch((err) => setMessage(err.message))
  }, [loadRequests, requestAppliedFilters, requestPage, requestPageSize])

  useEffect(() => {
    loadAudit(auditAppliedFilters, auditPage, auditPageSize).catch((err) => setMessage(err.message))
  }, [loadAudit, auditAppliedFilters, auditPage, auditPageSize])

  const handleRefresh = useCallback(() => {
    loadRequests().catch(() => {})
    loadAudit().catch(() => {})
  }, [loadRequests, loadAudit])

  useEffect(() => {
    registerRefresh(handleRefresh)
  }, [registerRefresh, handleRefresh])

  useEffect(() => {
    Promise.all([
      api.clientKeys('?page=1&page_size=100'),
      api.models('?page=1&page_size=100')
    ]).then(([keysRes, modelsRes]) => {
      setClientKeys(keysRes.items || [])
      setModels(modelsRes.items || [])
    }).catch(console.error)
  }, [])

  useEffect(() => {
    setRequestPage(1)
  }, [requestPageSize])

  useEffect(() => {
    setAuditPage(1)
  }, [auditPageSize])

  async function showRequestDetail(id: number) {
    try {
      const data = await api.requestLogDetail(id)
      setDetailType('request')
      setDetail(data)
      setDetailModalOpen(true)
    } catch (err) {
      showError(err instanceof Error ? err.message : '加载详情失败')
    }
  }

  async function showAuditDetail(id: number) {
    try {
      const data = await api.auditLogDetail(id)
      setDetailType('audit')
      setDetail(data)
      setDetailModalOpen(true)
    } catch (err) {
      showError(err instanceof Error ? err.message : '加载详情失败')
    }
  }

  function handleSearch() {
    if (tab === 'requests') {
      setRequestAppliedFilters(requestDraftFilters)
      setRequestPage(1)
      return
    }
    setAuditAppliedFilters(auditDraftFilters)
    setAuditPage(1)
  }

  function handleClear() {
    if (tab === 'requests') {
      setRequestDraftFilters({ client_id: '', virtual_model_code: '', success: '', time_range: '1d' })
      setRequestAppliedFilters({ client_id: '', virtual_model_code: '', success: '', time_range: '1d' })
      setRequestPage(1)
      return
    }
    setAuditDraftFilters(defaultAuditFilters)
    setAuditAppliedFilters(defaultAuditFilters)
    setAuditPage(1)
  }

  async function exportAuditLogs() {
    setExportingAudit(true)
    try {
      await api.downloadAuditLogs(queryString(auditAppliedFilters))
      showSuccess('审计日志导出成功')
    } catch (err) {
      showError(err instanceof Error ? err.message : '导出失败')
    } finally {
      setExportingAudit(false)
    }
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="p-4">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-1 relative">
              <Button variant={tab === 'requests' ? 'default' : 'secondary'} onClick={() => setTab('requests')} size="lg" className={`flex-1 font-semibold`}><LuFileText className="mr-2 h-5 w-5" />请求日志</Button>
              <Button variant={tab === 'audit' ? 'default' : 'secondary'} onClick={() => setTab('audit')} size="lg" className={`flex-1 font-semibold`}><MdOutlineHistory className="mr-2 h-5 w-5" />审计日志</Button>
            </div>
            <div className="flex items-center gap-3">
              {tab === 'audit' ? <Button variant="secondary" onClick={exportAuditLogs} size="lg" loading={exportingAudit}><FiDownload className="mr-2 h-5 w-5" />导出 CSV</Button> : null}
              <Button variant="secondary" onClick={handleSearch} size="lg" className="min-w-[80px]">搜索</Button>
              <Button variant="secondary" onClick={handleClear} size="lg" className="min-w-[80px]">清空</Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {tab === 'requests' ? (
        <div className="space-y-4">
          <Card>
            <CardContent className="p-4">
              <div className="grid gap-3 xl:grid-cols-4">
                <div className="flex items-center gap-3"><label className="w-20 text-sm">API Key</label>
                  <Select value={requestDraftFilters.client_id || 'all'} onValueChange={(value) => setRequestDraftFilters({ ...requestDraftFilters, client_id: value === 'all' ? '' : value })}>
                    <SelectTrigger className="flex-1"><SelectValue placeholder="全部" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">全部</SelectItem>
                      {clientKeys.map((key) => (
                        <SelectItem key={key.id} value={key.id.toString()}>{key.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">模型</label>
                  <Select value={requestDraftFilters.virtual_model_code || 'all'} onValueChange={(value) => setRequestDraftFilters({ ...requestDraftFilters, virtual_model_code: value === 'all' ? '' : value })}>
                    <SelectTrigger className="flex-1"><SelectValue placeholder="全部" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">全部</SelectItem>
                      {models.map((model) => (
                        <SelectItem key={model.id} value={model.model_code}>{model.display_name || model.model_code}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">状态</label><Select value={requestDraftFilters.success || 'all'} onValueChange={(value) => setRequestDraftFilters({ ...requestDraftFilters, success: value === 'all' ? '' : value })}>
                  <SelectTrigger className="flex-1"><SelectValue placeholder="状态" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部</SelectItem>
                    <SelectItem value="true">成功</SelectItem>
                    <SelectItem value="false">失败</SelectItem>
                  </SelectContent>
                </Select></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">时间</label>
                  <Select value={requestDraftFilters.time_range || '1d'} onValueChange={(value) => setRequestDraftFilters({ ...requestDraftFilters, time_range: value })}>
                    <SelectTrigger className="flex-1"><SelectValue placeholder="时间" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1h">最近 1 小时</SelectItem>
                      <SelectItem value="6h">最近 6 小时</SelectItem>
                      <SelectItem value="1d">最近 1 天</SelectItem>
                      <SelectItem value="3d">最近 3 天</SelectItem>
                      <SelectItem value="7d">最近 7 天</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="overflow-hidden mb-4">
            <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
              <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>请求日志</div>
              <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>{requestTotal} 条记录</div>
            </div>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>状态</TableHead>
                    <TableHead>时间</TableHead>
                    <TableHead>Key</TableHead>
                    <TableHead>请求内容</TableHead>
                    <TableHead>模型</TableHead>
                    <TableHead>Token</TableHead>
                    <TableHead>延迟</TableHead>
                    <TableHead>操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {requestLogs.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={8} className="h-32 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</TableCell>
                    </TableRow>
                  ) : requestLogs.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell><span style={{ color: item.success ? 'var(--success)' : 'var(--danger)' }}>{item.success ? '成功' : '失败'}</span></TableCell>
                      <TableCell>{formatDateTime(item.created_at)}</TableCell>
                      <TableCell className="font-medium">{item.key_name || '-'}</TableCell>
                      <TableCell className="max-w-[300px] truncate" title={item.request_summary}>
                        {parseRequestContent(item.request_summary)}
                      </TableCell>
                      <TableCell>{item.virtual_model_code || '-'}</TableCell>
                      <TableCell>{item.total_tokens ? item.total_tokens.toLocaleString() : '-'}</TableCell>
                      <TableCell><div className="flex items-center gap-1"><PiClockCountdownBold className="h-3.5 w-3.5 text-slate-500" />{item.latency_ms ?? '-'}ms</div></TableCell>
                      <TableCell><Button variant="secondary" size="sm" onClick={() => showRequestDetail(item.id)}><FiExternalLink className="h-4 w-4" /></Button></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
            <PaginationBar page={requestPage} pageSize={requestPageSize} total={requestTotal} onPageChange={setRequestPage} onPageSizeChange={setRequestPageSize} />
          </Card>
        </div>
      ) : (
        <div className="space-y-4">
          <Card>
            <CardContent className="p-4">
              <div className="grid gap-3 xl:grid-cols-4">
                <div className="flex items-center gap-3"><label className="w-20 text-sm">管理员</label><Input className="flex-1" value={auditDraftFilters.admin_user_id} onChange={(e) => setAuditDraftFilters({ ...auditDraftFilters, admin_user_id: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">动作</label><Input className="flex-1" value={auditDraftFilters.action} onChange={(e) => setAuditDraftFilters({ ...auditDraftFilters, action: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">对象类型</label><Input className="flex-1" value={auditDraftFilters.target_type} onChange={(e) => setAuditDraftFilters({ ...auditDraftFilters, target_type: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">对象 ID</label><Input className="flex-1" value={auditDraftFilters.target_id} onChange={(e) => setAuditDraftFilters({ ...auditDraftFilters, target_id: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">关键词</label><Input className="flex-1" value={auditDraftFilters.keyword} onChange={(e) => setAuditDraftFilters({ ...auditDraftFilters, keyword: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">开始时间</label><Input className="flex-1" type="datetime-local" value={auditDraftFilters.start_time} onChange={(e) => setAuditDraftFilters({ ...auditDraftFilters, start_time: e.target.value ? new Date(e.target.value).toISOString() : '' })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">结束时间</label><Input className="flex-1" type="datetime-local" value={auditDraftFilters.end_time} onChange={(e) => setAuditDraftFilters({ ...auditDraftFilters, end_time: e.target.value ? new Date(e.target.value).toISOString() : '' })} /></div>
              </div>
            </CardContent>
          </Card>

          <Card className="overflow-hidden mb-4">
            <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
              <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>审计日志</div>
              <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>{auditTotal} 条记录</div>
            </div>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>时间</TableHead>
                    <TableHead>操作人</TableHead>
                    <TableHead>动作</TableHead>
                    <TableHead>对象</TableHead>
                    <TableHead>摘要</TableHead>
                    <TableHead>操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {auditLogs.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="h-32 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</TableCell>
                    </TableRow>
                  ) : auditLogs.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>{formatDateTime(item.created_at)}</TableCell>
                      <TableCell>{item.admin_username || `admin#${item.admin_user_id}`}</TableCell>
                      <TableCell>{item.action}</TableCell>
                      <TableCell>{item.target_name ? `${item.target_type} / ${item.target_name}` : `${item.target_type}#${item.target_id ?? '-'}`}</TableCell>
                      <TableCell className="max-w-[320px] truncate" title={item.change_summary}>{item.change_summary || '-'}</TableCell>
                      <TableCell><Button variant="secondary" size="sm" onClick={() => showAuditDetail(item.id)}><FiExternalLink className="h-4 w-4" /></Button></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
            <PaginationBar page={auditPage} pageSize={auditPageSize} total={auditTotal} onPageChange={setAuditPage} onPageSizeChange={setAuditPageSize} />
          </Card>
        </div>
      )}

      {message ? <Card><CardContent className="p-4 text-sm" style={{ color: 'var(--muted-foreground)' }}>{message}</CardContent></Card> : null}

      <Modal
        open={detailModalOpen}
        title={detailType === 'request' ? '请求详情' : '审计详情'}
        onClose={() => setDetailModalOpen(false)}
        footer={
          <div className="flex justify-end">
            <Button variant="secondary" onClick={() => setDetailModalOpen(false)}>关闭</Button>
          </div>
        }
      >
        {detail ? (
          <div className="space-y-4">
            {detailType === 'audit' ? (
              <div className="grid gap-3 md:grid-cols-2 text-sm">
                <div><span style={{ color: 'var(--muted-foreground)' }}>操作人：</span>{detail.admin_username || `admin#${detail.admin_user_id}`}</div>
                <div><span style={{ color: 'var(--muted-foreground)' }}>动作：</span>{detail.action}</div>
                <div><span style={{ color: 'var(--muted-foreground)' }}>对象：</span>{detail.target_name ? `${detail.target_type} / ${detail.target_name}` : `${detail.target_type}#${detail.target_id ?? '-'}`}</div>
                <div><span style={{ color: 'var(--muted-foreground)' }}>时间：</span>{formatDateTime(detail.created_at)}</div>
                <div><span style={{ color: 'var(--muted-foreground)' }}>请求 ID：</span>{detail.request_id || '-'}</div>
                <div><span style={{ color: 'var(--muted-foreground)' }}>IP：</span>{detail.ip_address || '-'}</div>
                <div className="md:col-span-2"><span style={{ color: 'var(--muted-foreground)' }}>摘要：</span>{detail.change_summary || '-'}</div>
              </div>
            ) : null}
            <pre className="overflow-auto rounded-2xl border p-4 text-xs max-h-[60vh]" style={{ borderColor: 'var(--border)', background: 'rgba(255,255,255,0.03)' }}>{JSON.stringify(detail, null, 2)}</pre>
          </div>
        ) : (
          <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>加载中...</div>
        )}
      </Modal>
    </div>
  )
}
