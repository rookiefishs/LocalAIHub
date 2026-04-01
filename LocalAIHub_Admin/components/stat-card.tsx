import type { ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import Link from 'next/link'

export function StatCard({ title, value, subValue, hint, icon, href }: { title: string; value: string | number; subValue?: string; hint?: string; icon?: ReactNode; href?: string }) {
  const content = (
    <>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 p-5 pb-3">
        <CardTitle className="text-sm font-medium" style={{ color: 'var(--muted-foreground)' }}>{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent className="p-5 pt-0">
        <div className="text-3xl font-semibold">
          {value}
          {subValue && <span className="ml-1 text-lg font-normal" style={{ color: 'var(--muted-foreground)' }}>{subValue}</span>}
        </div>
        {hint ? <div className="mt-2 text-xs" style={{ color: 'var(--muted-foreground)' }}>{hint}</div> : null}
      </CardContent>
    </>
  )

  if (href) {
    return (
      <Link href={href} className="block transition-all duration-200 hover:-translate-y-1 cursor-pointer">
        <Card className="cursor-pointer transition-all duration-200 hover:border-[var(--primary)]/50">
          {content}
        </Card>
      </Link>
    )
  }

  return (
    <Card className="transition-all duration-200 hover:border-[var(--primary)]/30 hover:-translate-y-1">
      {content}
    </Card>
  )
}
