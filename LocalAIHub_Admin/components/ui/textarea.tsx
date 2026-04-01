import * as React from 'react'

import { cn } from '@/lib/utils'

const Textarea = React.forwardRef<HTMLTextAreaElement, React.ComponentProps<'textarea'>>(({ className, style, ...props }, ref) => {
  return (
    <textarea
      ref={ref}
      className={cn('flex min-h-[96px] w-full rounded-[10px] border px-4 py-3 text-sm outline-none transition placeholder:text-slate-400 focus-visible:ring-2', className)}
      style={{ borderColor: 'var(--border)', background: 'var(--input)', color: 'var(--foreground)', ...style }}
      {...props}
    />
  )
})
Textarea.displayName = 'Textarea'

export { Textarea }
