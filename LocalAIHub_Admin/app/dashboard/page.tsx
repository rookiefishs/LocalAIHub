'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { HiOutlineServerStack } from 'react-icons/hi2'
import { LuActivity } from 'react-icons/lu'
import { TbClockHour4 } from 'react-icons/tb'
import { HiOutlineGlobeAlt } from 'react-icons/hi2'
import { StatCard } from '@/components/stat-card'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api } from '@/lib/api'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, PieChart, Pie, Cell } from 'recharts'
import { useRefresh } from '@/components/refresh-context'
import type { ClientKey, DashboardData } from '@/lib/types'

function formatTrendLabel(hour: string, timeRange: string) {
  if (!hour) return ''
  const [datePart, timePart = ''] = hour.split(' ')
  if (!datePart) return hour
  if (timeRange === '1d') {
    return timePart || hour
  }
  const [year, month, day] = datePart.split('-')
  if (!year || !month || !day) return hour
  return `${month}-${day} ${timePart}`.trim()
}

function getMinVisibleBarSize(value: number | null | undefined) {
  return typeof value === 'number' && value > 0 ? 6 : 0
}

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('1d')
  const [selectedKey, setSelectedKey] = useState<string>('all')
  const [keySortBy, setKeySortBy] = useState<'tokens' | 'requests' | 'success_rate'>('tokens')
  const [clientKeys, setClientKeys] = useState<ClientKey[]>([])
  const { registerRefresh } = useRefresh()

  async function loadClientKeys() {
    try {
      const res = await api.clientKeys('?page=1&page_size=100')
      setClientKeys(res.items || [])
    } catch (err) {
      console.error('failed to load client keys:', err)
    }
  }

  async function loadDashboard() {
    setLoading(true)
    try {
      const hours = timeRange === '1d' ? 24 : timeRange === '3d' ? 72 : timeRange === '7d' ? 168 : 720
      const query = `hours=${hours}${selectedKey !== 'all' ? `&client_id=${selectedKey}` : ''}`
      const result = await api.dashboard(query)
      setData(result)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadClientKeys()
  }, [])

  useEffect(() => {
    registerRefresh(loadDashboard)
  }, [registerRefresh])

  useEffect(() => {
    loadDashboard()
  }, [timeRange, selectedKey])

  useEffect(() => {
    const timer = setInterval(() => {
      loadDashboard()
    }, 30000)
    return () => clearInterval(timer)
  }, [])

  const selectedKeyName = selectedKey === 'all' 
    ? '全部' 
    : clientKeys.find(k => k.id.toString() === selectedKey)?.name || '未知'

  const effectiveData = data

  const keyStatsData = selectedKey === 'all' 
    ? (effectiveData?.key_stats || []).map((item: any) => ({
        name: item.key_name || '未知',
        requests: item.request_count || 0,
        tokens: item.total_tokens || 0,
        success_rate: item.success_rate || 0,
      })).sort((a: any, b: any) => {
        if (keySortBy === 'tokens') return b.tokens - a.tokens
        if (keySortBy === 'requests') return b.requests - a.requests
        return b.success_rate - a.success_rate
      })
    : []

  const keyColors = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899', '#84cc16', '#14b8a6', '#f97316']

  const keyTrendData = selectedKey === 'all' && effectiveData?.key_trend ? effectiveData.key_trend : []
  const keyModelData = selectedKey === 'all' && effectiveData?.key_model_distribution ? effectiveData.key_model_distribution : []

  const activeKeyNames = new Set(clientKeys.map(k => k.name))
  const filteredKeyTrendData = keyTrendData.filter((t: any) => activeKeyNames.has(t.key_name))
  const filteredKeyModelData = keyModelData.filter((t: any) => activeKeyNames.has(t.key_name))

  const allKeyNames = [
    ...new Set([
      ...keyStatsData.map((item: any) => item.name).filter((value: any): value is string => Boolean(value)),
      ...filteredKeyTrendData.map((t: any) => t.key_name).filter((value: any): value is string => Boolean(value)),
      ...filteredKeyModelData.map((t: any) => t.key_name).filter((value: any): value is string => Boolean(value)),
    ]),
  ]
  const keyColorMap = new Map(allKeyNames.map((keyName, idx) => [keyName, keyColors[idx % keyColors.length]]))

  const keyNames = allKeyNames.filter((keyName) => filteredKeyTrendData.some((t: any) => t.key_name === keyName))
  const keyTrendHours: string[] = [...new Set(filteredKeyTrendData.map((t: any) => t.hour).filter((value: any): value is string => Boolean(value)))].sort() as string[]
  const keyTrendBuckets = keyTrendHours.map((hour: string) => ({
    hour,
    label: formatTrendLabel(hour, timeRange),
  }))

  const unifiedTrendData = keyTrendBuckets.map(({ hour, label }) => {
    const entry: any = { hour, time: label }
    keyNames.forEach((keyName: any) => {
      const dataPoint = filteredKeyTrendData.find((t: any) => t.key_name === keyName && t.hour === hour)
      entry[keyName + '_requests'] = dataPoint ? (dataPoint.count || 0) : 0
      entry[keyName + '_tokens'] = dataPoint ? (dataPoint.tokens || 0) : 0
    })
    return entry
  })

  const modelNames = [...new Set(filteredKeyModelData.map((t: any) => t.model_code))]
  const keyModelNames = allKeyNames.filter((keyName) => filteredKeyModelData.some((t: any) => t.key_name === keyName))

  const modelChartData = modelNames.map((modelCode: any) => {
    const entry: any = { modelCode }
    keyModelNames.forEach((keyName: any) => {
      const item = filteredKeyModelData.find((t: any) => t.model_code === modelCode && t.key_name === keyName)
      entry[keyName] = item ? item.count : 0
    })
    return entry
  })

  const chartData = useMemo(() => (effectiveData?.request_trend?.map((item: any) => ({
    hour: item.hour || '',
    time: formatTrendLabel(item.hour || '', timeRange),
    requests: item.count || 0,
    success: item.success || 0,
    tokens: item.total_tokens || 0,
  })) || []), [effectiveData?.request_trend, timeRange])

  const emptyTrendData = useMemo(() => {
    if (timeRange === '1d') {
      return [
        { hour: '', time: '00:00', requests: 0, success: 0, tokens: 0 },
        { hour: '', time: '12:00', requests: 0, success: 0, tokens: 0 },
        { hour: '', time: '23:59', requests: 0, success: 0, tokens: 0 },
      ]
    }

    if (timeRange === '30d') {
      return [
        { hour: '', time: '起始', requests: 0, success: 0, tokens: 0 },
        { hour: '', time: '中间', requests: 0, success: 0, tokens: 0 },
        { hour: '', time: '当前', requests: 0, success: 0, tokens: 0 },
      ]
    }

    return [
      { hour: '', time: '开始', requests: 0, success: 0, tokens: 0 },
      { hour: '', time: '中间', requests: 0, success: 0, tokens: 0 },
      { hour: '', time: '当前', requests: 0, success: 0, tokens: 0 },
    ]
  }, [timeRange])

  const requestChartSummary = useMemo(() => chartData.reduce((acc: { requests: number; tokens: number }, item: any) => {
    acc.requests += item.requests || 0
    acc.tokens += item.tokens || 0
    return acc
  }, { requests: 0, tokens: 0 }), [chartData])

  const successCount = effectiveData?.success_count ?? Math.round((effectiveData?.request_count ?? requestChartSummary.requests) * (effectiveData?.success_rate || 0))
  const failureCount = effectiveData?.failure_count ?? Math.max((effectiveData?.request_count ?? requestChartSummary.requests) - successCount, 0)

  const timeRangeLabel = timeRange === '1d' ? '24h' : timeRange === '3d' ? '3天' : timeRange === '7d' ? '7天' : '30天'

  const modelData = (effectiveData?.model_distribution || []).map((item: any) => ({
    name: item.model_code,
    value: item.count,
  }))
  const emptyModelChartData = [{ modelCode: '暂无记录', value: 0 }]
  const COLORS = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899', '#84cc16', '#14b8a6', '#f97316']

  return (
    <div className="space-y-4">
      <div className="flex justify-end gap-2">
        <Select value={selectedKey} onValueChange={setSelectedKey}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="选择 API Key" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部 Key</SelectItem>
            {clientKeys.map((key) => (
              <SelectItem key={key.id} value={key.id.toString()}>
                {key.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={timeRange} onValueChange={setTimeRange}>
          <SelectTrigger className="w-28">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1d">1 天</SelectItem>
            <SelectItem value="3d">3 天</SelectItem>
            <SelectItem value="7d">7 天</SelectItem>
            <SelectItem value="30d">30 天</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
        {loading ? Array.from({ length: 5 }).map((_, index) => (
          <Card key={index}>
            <CardContent className="flex h-[92px] items-center justify-center">
              <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600" />
                加载中...
              </div>
            </CardContent>
          </Card>
        )) : (
          <>
            <StatCard title={`请求数 (${timeRangeLabel})`} value={effectiveData?.request_count ?? '-'} icon={<LuActivity className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
            <StatCard title={`成功率 (${timeRangeLabel})`} value={effectiveData ? `${Math.round((effectiveData.success_rate || 0) * 100)}%` : '-'} icon={<HiOutlineServerStack className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
            <StatCard title="平均延迟" value={effectiveData?.avg_latency_ms ? `${effectiveData.avg_latency_ms}ms` : '-'} icon={<TbClockHour4 className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
            <StatCard title={`Token (${timeRangeLabel})`} value={effectiveData?.total_tokens ? `${(effectiveData.total_tokens / 1000).toFixed(1)}k` : '-'} icon={<LuActivity className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
            <StatCard title={`上游 (${timeRangeLabel})`} value={effectiveData?.active_upstream_count ?? '-'} subValue={effectiveData?.active_upstream_count ? "启用" : undefined} icon={<HiOutlineGlobeAlt className="h-4 w-4 text-slate-400" />} href="/dashboard/upstreams" />
          </>
        )}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>请求趋势 {selectedKey !== 'all' && <span className="text-sm font-normal text-muted-foreground">- {selectedKeyName}</span>}</CardTitle>
          </CardHeader>
          <CardContent>
            {selectedKey === 'all' && unifiedTrendData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <AreaChart data={unifiedTrendData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <defs>
                    {keyNames.map((keyName) => (
                      <linearGradient key={keyName} id={`colorKey-${keyName}`} x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor={keyColorMap.get(keyName) || keyColors[0]} stopOpacity={0.3}/>
                        <stop offset="95%" stopColor={keyColorMap.get(keyName) || keyColors[0]} stopOpacity={0}/>
                      </linearGradient>
                    ))}
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                   <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} minTickGap={timeRange === '1d' ? 24 : timeRange === '30d' ? 200 : 80} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                     labelFormatter={(_, payload) => payload?.[0]?.payload?.hour || ''}
                     formatter={(value: number, name: string) => [value.toLocaleString(), name.replace('_requests', '')]}
                  />
                  {keyNames.map((keyName: any) => (
                    <Area key={keyName} type="monotone" dataKey={keyName + '_requests'} stroke={keyColorMap.get(keyName) || keyColors[0]} strokeWidth={2} fillOpacity={1} fill={`url(#colorKey-${keyName})`} name={keyName} />
                  ))}
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <ResponsiveContainer width="100%" height={280}>
                <AreaChart data={chartData.length > 0 ? chartData : emptyTrendData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                   <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} minTickGap={timeRange === '1d' ? 24 : timeRange === '30d' ? 200 : 80} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                    labelFormatter={(_, payload) => payload?.[0]?.payload?.hour || ''}
                  />
                  <Area type="monotone" dataKey="requests" stroke="#3b82f6" strokeWidth={2} fillOpacity={1} fill="url(#colorRequests)" name="请求数" />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Token 趋势 {selectedKey !== 'all' && <span className="text-sm font-normal text-muted-foreground">- {selectedKeyName}</span>}</CardTitle>
          </CardHeader>
          <CardContent>
            {selectedKey === 'all' && unifiedTrendData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={unifiedTrendData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                   <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} minTickGap={timeRange === '1d' ? 24 : timeRange === '30d' ? 200 : 80} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                     labelFormatter={(_, payload) => payload?.[0]?.payload?.hour || ''}
                     formatter={(value: number, name: string) => [value.toLocaleString(), name.replace('_tokens', '')]}
                  />
                  {keyNames.map((keyName: any) => (
                    <Bar key={keyName} dataKey={keyName + '_tokens'} stackId="token" fill={keyColorMap.get(keyName) || keyColors[0]} name={keyName} radius={[4, 4, 0, 0]} minPointSize={getMinVisibleBarSize} />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={chartData.length > 0 ? chartData : emptyTrendData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                   <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} minTickGap={timeRange === '1d' ? 24 : timeRange === '30d' ? 200 : 80} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                     labelFormatter={(_, payload) => payload?.[0]?.payload?.hour || ''}
                     formatter={(value: number) => [value.toLocaleString(), 'Token']}
                   />
                    <Bar dataKey="tokens" fill="#8b5cf6" name="Token" radius={[4, 4, 0, 0]} minPointSize={getMinVisibleBarSize} />
                  </BarChart>
                </ResponsiveContainer>
              )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>模型分布 {selectedKey === 'all' && <span className="text-sm font-normal text-muted-foreground">- 按 Key</span>}</CardTitle>
          </CardHeader>
          <CardContent>
            {selectedKey === 'all' && modelNames.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={modelChartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="modelCode" stroke="var(--muted-foreground)" fontSize={11} tickLine={false} angle={-45} textAnchor="end" height={60} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                  />
                  {keyModelNames.map((keyName: any) => (
                    <Bar key={keyName} dataKey={keyName} stackId="model" fill={keyColorMap.get(keyName) || keyColors[0]} name={keyName} radius={[4, 4, 0, 0]} />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            ) : modelData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <PieChart>
                  <Pie
                    data={modelData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={2}
                    dataKey="value"
                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  >
                    {modelData.map((entry: any, index: number) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip 
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={emptyModelChartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="modelCode" stroke="var(--muted-foreground)" fontSize={11} tickLine={false} angle={-45} textAnchor="end" height={60} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} domain={[0, 1]} />
                  <Tooltip
                    contentStyle={{
                      background: 'var(--card)',
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                    formatter={() => ['0', '请求数']}
                  />
                  <Bar dataKey="value" fill="#cbd5e1" name="请求数" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>各 Key 使用情况</CardTitle>
                <Select value={keySortBy} onValueChange={(v: any) => setKeySortBy(v)}>
                  <SelectTrigger className="w-32 h-8 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="tokens">Token 排序</SelectItem>
                    <SelectItem value="requests">请求次数</SelectItem>
                    <SelectItem value="success_rate">成功率</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardHeader>
            <CardContent>
              {keyStatsData.length > 0 ? (
                <div className="space-y-3 max-h-[280px] overflow-y-auto">
                  {keyStatsData.map((item: any, index: number) => (
                  <div key={index} className="flex items-center justify-between p-3 rounded-[10px] border" style={{ borderColor: 'var(--border)' }}>
                    <div>
                      <div className="font-medium text-sm" style={{ color: 'var(--foreground)' }}>{item.name}</div>
                      <div className="text-xs mt-1" style={{ color: 'var(--muted-foreground)' }}>
                        请求: {item.requests.toLocaleString()} | Token: {(item.tokens / 1000).toFixed(1)}k
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-lg font-bold" style={{ color: item.success_rate >= 0.9 ? '#10b981' : item.success_rate >= 0.5 ? '#f59e0b' : '#ef4444' }}>
                        {Math.round(item.success_rate * 100)}%
                      </div>
                      <div className="text-xs" style={{ color: 'var(--muted-foreground)' }}>成功率</div>
                    </div>
                  </div>
                  ))}
                </div>
              ) : (
                <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                  暂无 Key 使用数据
                </div>
              )}
            </CardContent>
          </Card>

        {selectedKey !== 'all' && (
          <Card>
            <CardHeader>
              <CardTitle>请求状态 {selectedKey !== 'all' && <span className="text-sm font-normal text-muted-foreground">- {selectedKeyName}</span>}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-6">
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(59,130,246,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>总请求</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#3b82f6' }}>{effectiveData?.request_count ?? requestChartSummary.requests}</div>
                </div>
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(16,185,129,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>成功</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#10b981' }}>{successCount}</div>
                </div>
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(239,68,68,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>失败</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#ef4444' }}>{failureCount}</div>
                </div>
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(139,92,246,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>平均延迟</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#8b5cf6' }}>{effectiveData?.avg_latency_ms ?? 0}<span className="text-sm font-normal">ms</span></div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {error ? <div className="rounded-[10px] border px-4 py-3 text-sm" style={{ borderColor: 'rgba(239,95,114,0.25)', background: 'rgba(239,95,114,0.08)', color: '#ffb4bd' }}>{error}</div> : null}
    </div>
  )
}
