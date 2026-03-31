import type { ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function StatCard({ title, value, subValue, hint, icon }: { title: string; value: string | number; subValue?: string; hint?: string; icon?: ReactNode }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 p-5 pb-3">
        <CardTitle className="text-sm font-medium" style={{ color: 'var(--muted-foreground)' }}>{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent className="p-5 pt-0">
        <div className="text-3xl font-semibold">{value}</div>
        {subValue ? <div className="mt-1 text-xs" style={{ color: 'var(--muted-foreground)' }}>{subValue}</div> : null}
        {hint ? <div className="mt-2 text-xs" style={{ color: 'var(--muted-foreground)' }}>{hint}</div> : null}
      </CardContent>
    </Card>
  )
}
