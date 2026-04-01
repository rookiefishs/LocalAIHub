import * as React from 'react'

import { cn } from '@/lib/utils'

const Checkbox = React.forwardRef<HTMLInputElement, React.ComponentProps<'input'> & { label?: string }>(
  ({ className, label, ...props }, ref) => {
    return (
      <label className="flex items-center gap-2 cursor-pointer">
        <input
          ref={ref}
          type="checkbox"
          className={cn('h-4 w-4 rounded border-0 cursor-pointer accent-[var(--foreground)]', className)}
          style={{ background: props.checked ? 'var(--foreground)' : 'var(--input)', borderColor: 'var(--border)' }}
          {...props}
        />
        {label && (
          <span className="text-sm" style={{ color: 'var(--foreground)' }}>{label}</span>
        )}
      </label>
    )
  }
)
Checkbox.displayName = 'Input'

export { Checkbox }
