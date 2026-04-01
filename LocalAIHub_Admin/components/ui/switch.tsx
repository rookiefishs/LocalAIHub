import * as React from 'react'

import { cn } from '@/lib/utils'

const Switch = React.forwardRef<HTMLInputElement, React.ComponentProps<'input'> & { label?: string }>(
  ({ className, label, checked, onChange, ...props }, ref) => {
    return (
      <label className="flex items-center gap-3 cursor-pointer">
        <div className="relative">
          <input
            ref={ref}
            type="checkbox"
            className="sr-only"
            checked={checked}
            onChange={onChange}
            {...props}
          />
          <div 
            className={cn(
              'w-11 h-6 rounded-full transition-colors',
              checked ? 'bg-[var(--foreground)]' : 'bg-[var(--muted)]'
            )}
          />
          <div 
            className={cn(
              'absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform',
              checked && 'translate-x-5'
            )}
          />
        </div>
        {label && (
          <span className="text-sm" style={{ color: 'var(--foreground)' }}>{label}</span>
        )}
      </label>
    )
  }
)
Switch.displayName = 'Switch'

export { Switch }
