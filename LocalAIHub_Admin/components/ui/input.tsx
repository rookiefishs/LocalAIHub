import * as React from 'react'

import { cn } from '@/lib/utils'

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<'input'>>(({ className, style, ...props }, ref) => {
  return (
    <input
      ref={ref}
      className={cn('flex h-11 w-full rounded-xl border px-4 py-3 text-sm outline-none transition placeholder:text-slate-400 focus-visible:ring-2', className)}
      style={{ borderColor: 'var(--border)', background: 'var(--input)', color: 'var(--foreground)', boxShadow: 'none', ...style }}
      {...props}
    />
  )
})
Input.displayName = 'Input'

export { Input }
