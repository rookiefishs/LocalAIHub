import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import { Loader2 } from 'lucide-react'

import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 rounded-[10px] text-sm font-normal transition-all duration-150 ease-out disabled:pointer-events-none disabled:opacity-50 active:scale-[0.97]',
  {
    variants: {
      variant: {
        default: 'bg-[var(--foreground)] text-[var(--background)] font-medium hover:opacity-90 hover:shadow-[0_8px_20px_rgba(15,23,42,0.12)] dark:hover:shadow-[0_10px_24px_rgba(0,0,0,0.45)] hover:translate-y-[-0.5px]',
        secondary: 'border border-[var(--border)] bg-[var(--secondary)] text-[var(--foreground)] hover:bg-[var(--muted)] font-normal hover:shadow-[0_8px_20px_rgba(15,23,42,0.12)] dark:hover:shadow-[0_10px_24px_rgba(0,0,0,0.45)] hover:translate-y-[-0.5px] dark:hover:border-slate-400',
        ghost: 'hover:bg-[var(--muted)] text-[var(--foreground)] font-normal hover:translate-y-[-0.5px] dark:hover:bg-white/10',
        destructive: 'border border-[var(--danger)] bg-transparent text-[var(--danger)] hover:bg-[var(--danger)] hover:text-white font-normal hover:shadow-[0_8px_20px_rgba(15,23,42,0.12)] dark:hover:shadow-[0_10px_24px_rgba(0,0,0,0.45)] hover:translate-y-[-0.5px] dark:hover:border-rose-300',
      },
      size: {
        default: 'h-9 px-3.5 py-2',
        sm: 'h-8 px-2.5',
        lg: 'h-11 px-6 text-base',
        icon: 'h-9 w-9',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
  loading?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, loading, disabled, style, children, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button'
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || loading}
        style={variant === 'secondary' ? { borderColor: 'var(--border)', background: 'rgba(255,255,255,0.03)', ...style } : style}
        {...props}
      >
        {loading && <Loader2 className="h-4 w-4 animate-spin" />}
        {children}
      </Comp>
    )
  }
)
Button.displayName = 'Button'

export { Button, buttonVariants }
