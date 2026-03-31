"use client"

import {
  Area,
  AreaChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from "recharts"

// Generate mock data for the last 24 hours
const generateChartData = () => {
  const data = []
  const now = new Date()
  for (let i = 23; i >= 0; i--) {
    const hour = new Date(now.getTime() - i * 60 * 60 * 1000)
    data.push({
      time: `${hour.getHours().toString().padStart(2, '0')}:00`,
      requests: Math.floor(Math.random() * 3000) + 2000,
      errors: Math.floor(Math.random() * 30) + 5,
    })
  }
  return data
}

const data = generateChartData()

export function OverviewChart() {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <AreaChart data={data}>
        <defs>
          <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="hsl(180 60% 50%)" stopOpacity={0.3} />
            <stop offset="95%" stopColor="hsl(180 60% 50%)" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid
          strokeDasharray="3 3"
          stroke="hsl(var(--border))"
          vertical={false}
        />
        <XAxis
          dataKey="time"
          stroke="hsl(var(--muted-foreground))"
          fontSize={12}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          stroke="hsl(var(--muted-foreground))"
          fontSize={12}
          tickLine={false}
          axisLine={false}
          tickFormatter={(value) => `${(value / 1000).toFixed(1)}K`}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: "hsl(var(--card))",
            border: "1px solid hsl(var(--border))",
            borderRadius: "8px",
            color: "hsl(var(--foreground))",
          }}
          labelStyle={{ color: "hsl(var(--muted-foreground))" }}
          formatter={(value: number) => [value.toLocaleString(), "请求数"]}
        />
        <Area
          type="monotone"
          dataKey="requests"
          stroke="hsl(180 60% 50%)"
          strokeWidth={2}
          fillOpacity={1}
          fill="url(#colorRequests)"
        />
      </AreaChart>
    </ResponsiveContainer>
  )
}
