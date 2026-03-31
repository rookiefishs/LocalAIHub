'use client'

import { createContext, ReactNode, useContext, useMemo, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { FiCheckCircle, FiXCircle } from 'react-icons/fi'

type ToastState = { message: string; type: 'success' | 'error' } | null

const ToastContext = createContext<{ showSuccess: (msg: string) => void; showError: (msg: string) => void }>({
  showSuccess: () => undefined,
  showError: () => undefined,
})

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toast, setToast] = useState<ToastState>(null)

  const value = useMemo(
    () => ({
      showSuccess: (message: string) => {
        setToast({ message, type: 'success' })
        setTimeout(() => setToast(null), 2200)
      },
      showError: (message: string) => {
        setToast({ message, type: 'error' })
        setTimeout(() => setToast(null), 2600)
      },
    }),
    []
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      <AnimatePresence>
        {toast ? (
          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 20, scale: 0.95 }}
            transition={{ duration: 0.2 }}
            className="fixed right-4 top-20 z-[60] min-w-[280px]"
          >
            <div className="flex items-center gap-3 rounded-xl border px-4 py-4 shadow-xl" style={{ borderColor: toast.type === 'success' ? 'rgba(52,211,153,0.5)' : 'rgba(239,95,114,0.5)', background: toast.type === 'success' ? 'rgba(16,185,129,0.95)' : 'rgba(239,68,68,0.95)', color: '#fff' }}>
              {toast.type === 'success' ? (
                <FiCheckCircle className="h-5 w-5 flex-shrink-0" />
              ) : (
                <FiXCircle className="h-5 w-5 flex-shrink-0" />
              )}
              <span className="text-sm font-medium">{toast.message}</span>
            </div>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </ToastContext.Provider>
  )
}

export function useToast() {
  return useContext(ToastContext)
}
