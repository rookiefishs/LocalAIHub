'use client'

import * as React from 'react'
import * as SelectPrimitive from '@radix-ui/react-select'
import { Check, ChevronDown } from 'lucide-react'
import { motion } from 'framer-motion'

import { cn } from '@/lib/utils'

const Select = SelectPrimitive.Root
const SelectValue = SelectPrimitive.Value

const SelectTrigger = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Trigger
    ref={ref}
    className={cn('flex h-11 w-full items-center justify-between rounded-[10px] border px-4 py-3 text-sm transition-all duration-150 hover:border-[var(--primary)] hover:shadow-[0_8px_20px_rgba(15,23,42,0.10)] dark:hover:shadow-[0_10px_24px_rgba(0,0,0,0.45)] dark:hover:bg-white/5', className)}
    style={{ borderColor: 'var(--border)', background: 'var(--input)', color: 'var(--foreground)' }}
    {...props}
  >
    {children}
    <SelectPrimitive.Icon asChild>
      <ChevronDown className="h-4 w-4 text-slate-400 transition-transform duration-200" />
    </SelectPrimitive.Icon>
  </SelectPrimitive.Trigger>
))
SelectTrigger.displayName = SelectPrimitive.Trigger.displayName

const SelectContent = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, position = 'popper', ...props }, ref) => (
  <SelectPrimitive.Portal>
    <SelectPrimitive.Content
      ref={ref}
      className={cn(
        'relative z-50 overflow-hidden rounded-[10px] border shadow-md',
        'data-[state=open]:animate-in data-[state=closed]:animate-out',
        'data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0',
        'data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95',
        'data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2',
        className
      )}
      style={{ borderColor: 'var(--border)', background: 'var(--popover)', color: 'var(--popover-foreground)' }}
      position={position}
      {...props}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: -4 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        transition={{ duration: 0.15, ease: 'easeOut' }}
      >
        <SelectPrimitive.Viewport
          className="p-1"
          style={{
            width: 'var(--radix-select-trigger-width)',
            minWidth: 'var(--radix-select-trigger-width)',
          }}
        >
          {children}
        </SelectPrimitive.Viewport>
      </motion.div>
    </SelectPrimitive.Content>
  </SelectPrimitive.Portal>
))
SelectContent.displayName = SelectPrimitive.Content.displayName

const SelectItem = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Item 
    ref={ref} 
    className={cn(
      'relative flex w-full cursor-default select-none items-center rounded-[8px] py-2.5 pl-9 pr-3 text-sm outline-none transition-all mx-1',
      'data-[highlighted]:bg-[var(--foreground)] data-[highlighted]:text-[var(--background)]',
      'data-[state=checked]:bg-[var(--foreground)] data-[state=checked]:text-[var(--background)] data-[state=checked]:font-medium',
      className
    )} 
    {...props}
  >
    <span className="absolute left-3 flex h-3.5 w-3.5 items-center justify-center">
      <SelectPrimitive.ItemIndicator>
        <Check className="h-4 w-4" style={{ color: 'var(--background)' }} />
      </SelectPrimitive.ItemIndicator>
    </span>
    <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
  </SelectPrimitive.Item>
))
SelectItem.displayName = SelectPrimitive.Item.displayName

export { Select, SelectContent, SelectItem, SelectTrigger, SelectValue }
