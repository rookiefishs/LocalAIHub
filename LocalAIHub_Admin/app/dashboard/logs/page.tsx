'use client'

import { useEffect, useMemo, useState } from 'react'
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

const defaultRequestFilters = { trace_id: '', client_id: '', provider_id: '', virtual_model_code: '', success: '', start_time: '', end_time: '' }
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
  return new Date(value).toLocaleString()
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
  const { showSuccess, showError } = useToast()
  const { registerRefresh } = useRefresh()

  const requestQuery = useMemo(() => queryString(requestAppliedFilters), [requestAppliedFilters])
  const auditQuery = useMemo(() => queryString(auditAppliedFilters), [auditAppliedFilters])

  async function loadRequests(filters = requestAppliedFilters, page = requestPage, pageSize = requestPageSize) {
    const filterQuery = queryString(filters)
    const data = await api.requestLogs(`${filterQuery}${filterQuery ? '&' : ''}page_size=${pageSize}&page=${page}`)
    setRequestLogs(data.items || [])
    setRequestTotal(data.total || 0)
  }

  async function loadAudit(filters = auditAppliedFilters, page = auditPage, pageSize = auditPageSize) {
    const filterQuery = queryString(filters)
    const data = await api.auditLogs(`${filterQuery}${filterQuery ? '&' : ''}page_size=${pageSize}&page=${page}`)
    setAuditLogs(data.items || [])
    setAuditTotal(data.total || 0)
  }

  useEffect(() => {
    loadRequests(requestAppliedFilters, requestPage, requestPageSize).catch((err) => setMessage(err.message))
  }, [requestQuery, requestPage, requestPageSize])

  useEffect(() => {
    loadAudit(auditAppliedFilters, auditPage, auditPageSize).catch((err) => setMessage(err.message))
  }, [auditQuery, auditPage, auditPageSize])

  useEffect(() => {
    registerRefresh(() => {
      loadRequests().catch(() => {})
      loadAudit().catch(() => {})
    })
  }, [registerRefresh, requestAppliedFilters, requestPage, requestPageSize, auditAppliedFilters, auditPage, auditPageSize])

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
      setRequestDraftFilters(defaultRequestFilters)
      setRequestAppliedFilters(defaultRequestFilters)
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
              <Button variant={tab === 'requests' ? 'default' : 'secondary'} onClick={() => setTab('requests')} size="sm" className={tab === 'requests' ? 'font-semibold' : ''}><LuFileText className="mr-1 h-4 w-4" />请求日志</Button>
              <Button variant={tab === 'audit' ? 'default' : 'secondary'} onClick={() => setTab('audit')} size="sm" className={tab === 'audit' ? 'font-semibold' : ''}><MdOutlineHistory className="mr-1 h-4 w-4" />审计日志</Button>
              <div className="absolute bottom-0 left-0 h-0.5 bg-primary transition-all duration-200" style={{ width: tab === 'requests' ? 'calc(50% - 4px)' : 'calc(50% + 4px)', transform: tab === 'audit' ? 'translateX(100%)' : 'translateX(0)' }} />
            </div>
            <div className="flex items-center gap-2">
              {tab === 'audit' ? <Button variant="secondary" onClick={exportAuditLogs} size="sm" loading={exportingAudit}><FiDownload className="mr-1 h-4 w-4" />导出 CSV</Button> : null}
              <Button variant="secondary" onClick={handleSearch} size="sm">搜索</Button>
              <Button variant="secondary" onClick={handleClear} size="sm">清空</Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {tab === 'requests' ? (
        <div className="space-y-4">
          <Card>
            <CardContent className="p-4">
              <div className="grid gap-3 xl:grid-cols-4">
                <div className="flex items-center gap-3"><label className="w-20 text-sm">Trace ID</label><Input className="flex-1" value={requestDraftFilters.trace_id} onChange={(e) => setRequestDraftFilters({ ...requestDraftFilters, trace_id: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">Client ID</label><Input className="flex-1" value={requestDraftFilters.client_id} onChange={(e) => setRequestDraftFilters({ ...requestDraftFilters, client_id: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">Provider ID</label><Input className="flex-1" value={requestDraftFilters.provider_id} onChange={(e) => setRequestDraftFilters({ ...requestDraftFilters, provider_id: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">模型</label><Input className="flex-1" value={requestDraftFilters.virtual_model_code} onChange={(e) => setRequestDraftFilters({ ...requestDraftFilters, virtual_model_code: e.target.value })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">状态</label><Select value={requestDraftFilters.success || 'all'} onValueChange={(value) => setRequestDraftFilters({ ...requestDraftFilters, success: value === 'all' ? '' : value })}>
                  <SelectTrigger className="flex-1"><SelectValue placeholder="状态" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部</SelectItem>
                    <SelectItem value="true">成功</SelectItem>
                    <SelectItem value="false">失败</SelectItem>
                  </SelectContent>
                </Select></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">开始时间</label><Input className="flex-1" type="datetime-local" value={requestDraftFilters.start_time} onChange={(e) => setRequestDraftFilters({ ...requestDraftFilters, start_time: e.target.value ? new Date(e.target.value).toISOString() : '' })} /></div>
                <div className="flex items-center gap-3"><label className="w-20 text-sm">结束时间</label><Input className="flex-1" type="datetime-local" value={requestDraftFilters.end_time} onChange={(e) => setRequestDraftFilters({ ...requestDraftFilters, end_time: e.target.value ? new Date(e.target.value).toISOString() : '' })} /></div>
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
                    <TableHead>Trace ID</TableHead>
                    <TableHead>模型</TableHead>
                    <TableHead>延迟</TableHead>
                    <TableHead>操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {requestLogs.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="h-32 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>暂无数据</TableCell>
                    </TableRow>
                  ) : requestLogs.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell><span style={{ color: item.success ? 'var(--success)' : 'var(--danger)' }}>{item.success ? '成功' : '失败'}</span></TableCell>
                      <TableCell>{formatDateTime(item.created_at)}</TableCell>
                      <TableCell className="font-mono text-xs" style={{ color: 'var(--muted-foreground)' }}>{item.trace_id}</TableCell>
                      <TableCell>{item.virtual_model_code || '-'}</TableCell>
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
