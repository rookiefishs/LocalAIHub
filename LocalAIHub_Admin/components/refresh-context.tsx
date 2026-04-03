'use client'

import { createContext, useContext, useState, useCallback, ReactNode } from 'react'

interface RefreshContextType {
  registerRefresh: (fn: () => void | Promise<void>) => void
  triggerRefresh: () => void
}

const RefreshContext = createContext<RefreshContextType | null>(null)

export function RefreshProvider({ children }: { children: ReactNode }) {
  const [refreshFn, setRefreshFn] = useState<() => void | Promise<void>>(() => {})

  const registerRefresh = useCallback((fn: () => void | Promise<void>) => {
    setRefreshFn(() => fn)
  }, [])

  const triggerRefresh = useCallback(() => {
    refreshFn()
  }, [refreshFn])

  return (
    <RefreshContext.Provider value={{ registerRefresh, triggerRefresh }}>
      {children}
    </RefreshContext.Provider>
  )
}

export function useRefresh() {
  const context = useContext(RefreshContext)
  if (!context) {
    throw new Error('useRefresh must be used within RefreshProvider')
  }
  return context
}