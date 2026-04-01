'use client'

import { useEffect, useState } from 'react'
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

interface ClientKey {
  id: number
  name: string
  key_prefix: string
  status: string
}

export default function DashboardPage() {
  const [data, setData] = useState<any>(null)
  const [error, setError] = useState('')
  const [timeRange, setTimeRange] = useState('1d')
  const [selectedKey, setSelectedKey] = useState<string>('all')
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
    try {
      const hours = timeRange === '1d' ? 24 : timeRange === '3d' ? 72 : 168
      const query = `?hours=${hours}${selectedKey !== 'all' ? `&client_id=${selectedKey}` : ''}`
      const result = await api.dashboard(query)
      setData(result)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    }
  }

  useEffect(() => {
    loadClientKeys()
  }, [])

  useEffect(() => {
    registerRefresh(loadDashboard)
    loadDashboard()
  }, [timeRange, selectedKey])

  useEffect(() => {
    const timer = setInterval(() => {
      loadDashboard()
    }, 30000)
    return () => clearInterval(timer)
  }, [timeRange, selectedKey])

  const selectedKeyName = selectedKey === 'all' 
    ? '全部' 
    : clientKeys.find(k => k.id.toString() === selectedKey)?.name || '未知'

  const keyStatsData = selectedKey === 'all' 
    ? (data?.key_stats || []).map((item: any) => ({
        name: item.key_name || '未知',
        requests: item.request_count || 0,
        tokens: item.total_tokens || 0,
        success_rate: item.success_rate || 0,
      }))
    : []

  const keyColors = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899', '#84cc16', '#14b8a6', '#f97316']

  const keyTrendData = selectedKey === 'all' && data?.key_trend ? data.key_trend : []
  const keyNames = [...new Set(keyTrendData.map((t: any) => t.key_name))]
  const allTimes = [...new Set(keyTrendData.map((t: any) => t.hour?.slice(11, 16)).filter(Boolean))]
  
  const unifiedTrendData = allTimes.map(time => {
    const entry: any = { time }
    keyNames.forEach((keyName: any) => {
      const dataPoint = keyTrendData.find((t: any) => t.key_name === keyName && t.hour?.slice(11, 16) === time)
      entry[keyName + '_requests'] = dataPoint ? (dataPoint.count || 0) : 0
      entry[keyName + '_tokens'] = dataPoint ? (dataPoint.tokens || 0) : 0
    })
    return entry
  })

  const keyModelData = selectedKey === 'all' && data?.key_model_distribution ? data.key_model_distribution : []
  const modelNames = [...new Set(keyModelData.map((t: any) => t.model_code))]
  const keyModelNames = [...new Set(keyModelData.map((t: any) => t.key_name))]
  
  const modelChartData = modelNames.map((modelCode: any) => {
    const entry: any = { modelCode }
    keyModelNames.forEach((keyName: any) => {
      const item = keyModelData.find((t: any) => t.model_code === modelCode && t.key_name === keyName)
      entry[keyName] = item ? item.count : 0
    })
    return entry
  })

  const chartData = data?.request_trend?.map((item: any) => ({
    time: item.hour?.slice(11, 16) || '',
    requests: item.count,
    success: item.success,
    tokens: (item.total_tokens || 0),
  })) || []

  const timeRangeLabel = timeRange === '1d' ? '24h' : timeRange === '3d' ? '3天' : '7天'

  const modelData = (data?.model_distribution || []).map((item: any) => ({
    name: item.model_code,
    value: item.count,
  }))
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
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
        <StatCard title={`请求数 (${timeRangeLabel})`} value={data?.request_count ?? '-'} icon={<LuActivity className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
        <StatCard title={`成功率 (${timeRangeLabel})`} value={data ? `${Math.round((data.success_rate || 0) * 100)}%` : '-'} icon={<HiOutlineServerStack className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
        <StatCard title="平均延迟" value={data?.avg_latency_ms ? `${data.avg_latency_ms}ms` : '-'} icon={<TbClockHour4 className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
        <StatCard title={`Token (${timeRangeLabel})`} value={data?.total_tokens ? `${(data.total_tokens / 1000).toFixed(1)}k` : '-'} icon={<LuActivity className="h-4 w-4 text-slate-400" />} href="/dashboard/logs" />
        <StatCard title={`上游 (${timeRangeLabel})`} value={data?.active_upstream_count ?? '-'} subValue={data?.active_upstream_count ? "启用" : undefined} icon={<HiOutlineGlobeAlt className="h-4 w-4 text-slate-400" />} href="/dashboard/upstreams" />
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
                    {keyNames.map((_, idx) => (
                      <linearGradient key={idx} id={`colorKey${idx}`} x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor={keyColors[idx % keyColors.length]} stopOpacity={0.3}/>
                        <stop offset="95%" stopColor={keyColors[idx % keyColors.length]} stopOpacity={0}/>
                      </linearGradient>
                    ))}
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                  />
                  {keyNames.map((keyName: any, idx: number) => (
                    <Area key={idx} type="monotone" dataKey={keyName + '_requests'} stackId="request" stroke={keyColors[idx % keyColors.length]} strokeWidth={2} fillOpacity={1} fill={`url(#colorKey${idx})`} name={keyName} />
                  ))}
                </AreaChart>
              </ResponsiveContainer>
            ) : chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                  />
                  <Area type="monotone" dataKey="requests" stroke="#3b82f6" strokeWidth={2} fillOpacity={1} fill="url(#colorRequests)" name="请求数" />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                暂无趋势数据
              </div>
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
                  <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                    formatter={(value: number, name: string) => [value.toLocaleString(), name.replace('_tokens', '')]}
                  />
                  {keyNames.map((keyName: any, idx: number) => (
                    <Bar key={idx} dataKey={keyName + '_tokens'} stackId="token" fill="#8B5CF6" name={keyName} radius={[4, 4, 0, 0]} />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            ) : chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="time" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ 
                      background: 'var(--card)', 
                      border: '1px solid var(--border)',
                      borderRadius: '8px'
                    }}
                    labelStyle={{ color: 'var(--foreground)' }}
                    formatter={(value: number) => [value.toLocaleString(), 'Token']}
                  />
                  <Bar dataKey="tokens" fill="#8b5cf6" name="Token" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                暂无 Token 数据
              </div>
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
                  {keyModelNames.map((keyName: any, idx: number) => (
                    <Bar key={idx} dataKey={keyName} stackId="model" fill={keyColors[idx % keyColors.length]} name={keyName} radius={[4, 4, 0, 0]} />
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
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                暂无模型数据
              </div>
            )}
          </CardContent>
        </Card>

        {selectedKey === 'all' && keyStatsData.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>各 Key 使用情况</CardTitle>
            </CardHeader>
            <CardContent>
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
            </CardContent>
          </Card>
        )}

        {selectedKey === 'all' && keyStatsData.length === 0 && (
          <Card>
            <CardHeader>
              <CardTitle>各 Key 使用情况</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                暂无 Key 使用数据
              </div>
            </CardContent>
          </Card>
        )}

        {selectedKey !== 'all' && (
          <Card>
            <CardHeader>
              <CardTitle>请求状态 {selectedKey !== 'all' && <span className="text-sm font-normal text-muted-foreground">- {selectedKeyName}</span>}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-6">
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(59,130,246,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>总请求</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#3b82f6' }}>{data?.request_count ?? 0}</div>
                </div>
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(16,185,129,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>成功</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#10b981' }}>{data?.success_rate ? Math.round(data.request_count * data.success_rate) : 0}</div>
                </div>
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(239,68,68,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>失败</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#ef4444' }}>{data?.request_count ? data.request_count - Math.round(data.request_count * (data.success_rate || 0)) : 0}</div>
                </div>
                <div className="rounded-[10px] border p-4" style={{ borderColor: 'var(--border)', background: 'rgba(139,92,246,0.05)' }}>
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>平均延迟</div>
                  <div className="text-2xl font-bold mt-1" style={{ color: '#8b5cf6' }}>{data?.avg_latency_ms ?? 0}<span className="text-sm font-normal">ms</span></div>
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
