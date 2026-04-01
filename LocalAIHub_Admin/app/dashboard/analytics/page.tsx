'use client'

import { useEffect, useState } from 'react'
import { LuActivity } from 'react-icons/lu'
import { HiOutlineCurrencyDollar } from 'react-icons/hi2'
import { TbClockHour4 } from 'react-icons/tb'
import { StatCard } from '@/components/stat-card'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api } from '@/lib/api'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, PieChart, Pie, Cell } from 'recharts'
import { useRefresh } from '@/components/refresh-context'
import Link from 'next/link'

interface ComparisonData {
  current: number
  previous: number
  change: number
  change_rate: number
  direction: string
}

export default function AnalyticsPage() {
  const [timeRange, setTimeRange] = useState('1d')
  const [costData, setCostData] = useState<any>(null)
  const [tokenData, setTokenData] = useState<any>(null)
  const [comparison, setComparison] = useState<any>(null)
  const [error, setError] = useState('')
  const { registerRefresh } = useRefresh()

  async function loadData() {
    try {
      const hours = timeRange === '1d' ? 24 : timeRange === '3d' ? 72 : timeRange === '7d' ? 168 : 720
      const query = `?hours=${hours}`

      const [cost, tokens, comp] = await Promise.all([
        api.analyticsCost(query),
        api.analyticsTokens(query),
        api.analyticsComparison(query)
      ])

      setCostData(cost)
      setTokenData(tokens)
      setComparison(comp)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    }
  }

  useEffect(() => {
    registerRefresh(loadData)
    loadData()
  }, [timeRange])

  useEffect(() => {
    const timer = setInterval(() => {
      loadData()
    }, 60000)
    return () => clearInterval(timer)
  }, [timeRange])

  const costTrendData = costData?.trend?.map((item: any) => ({
    date: item.period?.slice(0, 10) || '',
    cost: parseFloat(item.cost?.toFixed(2)) || 0,
    tokens: item.tokens || 0
  })) || []

  const tokenTrendData = tokenData?.trend?.map((item: any) => ({
    date: item.period?.slice(0, 10) || '',
    prompt: item.prompt_tokens || 0,
    completion: item.completion_tokens || 0,
    total: item.total_tokens || 0
  })) || []

  const providerCostData = costData?.by_provider?.map((item: any) => ({
    name: item.provider_name || 'Unknown',
    value: parseFloat(item.cost?.toFixed(2)) || 0
  })) || []

  const modelCostData = costData?.by_model?.slice(0, 10).map((item: any) => ({
    name: item.model_code || 'Unknown',
    value: parseFloat(item.cost?.toFixed(2)) || 0
  })) || []

  const timeRangeLabel = timeRange === '1d' ? '24h' : timeRange === '3d' ? '3天' : timeRange === '7d' ? '7天' : '30天'

  const formatChange = (comp: ComparisonData) => {
    if (!comp) return { value: '-', className: '' }
    const rate = comp.change_rate?.toFixed(1) || '0'
    const isUp = comp.direction === 'up'
    const isDown = comp.direction === 'down'
    const className = isUp ? 'text-emerald-400' : isDown ? 'text-red-400' : 'text-slate-400'
    const arrow = isUp ? '↑' : isDown ? '↓' : ''
    return { value: `${arrow}${rate}%`, className }
  }

  const requestChange = formatChange(comparison?.request_count)
  const tokenChange = formatChange(comparison?.total_tokens)
  const costChange = formatChange(comparison?.total_cost)

  const COLORS = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899', '#84cc16']

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
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

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title={`请求数 (${timeRangeLabel})`}
          value={comparison?.request_count?.current?.toLocaleString() || '-'}
          subValue={requestChange.value}
          icon={<LuActivity className="h-4 w-4 text-slate-400" />}
          href="/dashboard/logs"
        />
        <StatCard
          title={`Token (${timeRangeLabel})`}
          value={tokenData?.total_tokens ? `${(tokenData.total_tokens / 1000).toFixed(1)}k` : '-'}
          subValue={tokenChange.value}
          icon={<LuActivity className="h-4 w-4 text-slate-400" />}
          href="/dashboard/logs"
        />
        <StatCard
          title={`预估费用 (${timeRangeLabel})`}
          value={costData?.total_cost ? `$${costData.total_cost.toFixed(2)}` : '-'}
          subValue={costChange.value}
          icon={<HiOutlineCurrencyDollar className="h-4 w-4 text-slate-400" />}
        />
        <StatCard
          title="平均延迟"
          value="-"
          icon={<TbClockHour4 className="h-4 w-4 text-slate-400" />}
          href="/dashboard/logs"
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>成本趋势</CardTitle>
          </CardHeader>
          <CardContent>
            {costTrendData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <AreaChart data={costTrendData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorCost" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="date" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ background: 'var(--card)', border: '1px solid var(--border)', borderRadius: '8px' }}
                    labelStyle={{ color: 'var(--foreground)' }}
                    formatter={(value: number) => [`$${value.toFixed(2)}`, '费用']}
                  />
                  <Area type="monotone" dataKey="cost" stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorCost)" name="费用" />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                暂无成本数据
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Token 消耗趋势</CardTitle>
          </CardHeader>
          <CardContent>
            {tokenTrendData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={tokenTrendData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="date" stroke="var(--muted-foreground)" fontSize={12} tickLine={false} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{ background: 'var(--card)', border: '1px solid var(--border)', borderRadius: '8px' }}
                    labelStyle={{ color: 'var(--foreground)' }}
                    formatter={(value: number) => [value.toLocaleString(), 'Token']}
                  />
                  <Bar dataKey="prompt" stackId="a" fill="#3b82f6" name="Prompt" />
                  <Bar dataKey="completion" stackId="a" fill="#8b5cf6" name="Completion" />
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
            <CardTitle>供应商费用占比</CardTitle>
          </CardHeader>
          <CardContent>
            {providerCostData.length > 0 ? (
              <ResponsiveContainer width="100%" height={280}>
                <PieChart>
                  <Pie
                    data={providerCostData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={2}
                    dataKey="value"
                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  >
                    {providerCostData.map((_: any, index: number) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip formatter={(value: number) => `$${value.toFixed(2)}`} />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                暂无数据
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>模型费用排行</CardTitle>
          </CardHeader>
          <CardContent>
            {modelCostData.length > 0 ? (
              <div className="space-y-3">
                {modelCostData.map((item: any, index: number) => (
                  <div key={item.name} className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="flex h-6 w-6 items-center justify-center rounded-full text-xs font-medium" style={{ background: COLORS[index % COLORS.length], color: '#fff' }}>
                        {index + 1}
                      </div>
                      <span className="text-sm font-mono">{item.name}</span>
                    </div>
                    <span className="text-sm font-medium">${item.value.toFixed(2)}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex h-[280px] items-center justify-center rounded-[10px] border border-dashed text-sm" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
                暂无数据
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {error ? <div className="rounded-[10px] border px-4 py-3 text-sm" style={{ borderColor: 'rgba(239,95,114,0.25)', background: 'rgba(239,95,114,0.08)', color: '#ffb4bd' }}>{error}</div> : null}
    </div>
  )
}
