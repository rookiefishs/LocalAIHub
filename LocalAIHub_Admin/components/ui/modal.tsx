'use client'

import { ReactNode, useEffect, useState, useRef } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'

export function Modal({ open, title, children, onClose, footer }: { open: boolean; title: string; children: ReactNode; onClose: () => void; footer?: ReactNode }) {
  const [mounted, setMounted] = useState(false)
  const [show, setShow] = useState(false)
  const [opening, setOpening] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => { setMounted(true) }, [])

  useEffect(() => {
    if (open) {
      setShow(true)
      setOpening(true)
      if (timerRef.current) clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => setOpening(false), 15)
    } else {
      setOpening(true)
      if (timerRef.current) clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => { setShow(false); setOpening(false) }, 150)
    }
    return () => { if (timerRef.current) clearTimeout(timerRef.current) }
  }, [open])

  if (!mounted) return null

  return createPortal(
    <div
      className={`fixed inset-0 z-50 flex items-center justify-center transition-opacity duration-150 ${show ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
      style={{ background: 'rgba(0,0,0,0.7)' }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className={`relative w-full max-w-2xl rounded-[10px] border p-0 transition-all duration-150 ${show && !opening ? 'opacity-100 scale-100 translate-y-0' : 'opacity-0 scale-95 translate-y-4'}`}
        style={{ background: 'var(--card)', borderColor: 'var(--border)' }}
      >
        <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
          <div className="text-base font-medium" style={{ color: 'var(--foreground)' }}>{title}</div>
          <button
            onClick={onClose}
            className="rounded-sm p-1 opacity-70 transition-opacity hover:opacity-100"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="px-6 pt-6 pb-6">{children}</div>
        {footer && (
          <div className="flex justify-end gap-2 border-t px-6 py-4" style={{ borderColor: 'var(--border)' }}>
            {footer}
          </div>
        )}
      </div>
    </div>,
    document.body
  )
}
