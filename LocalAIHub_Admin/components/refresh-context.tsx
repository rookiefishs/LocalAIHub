'use client'

import { createContext, useContext, useState, ReactNode } from 'react'

interface RefreshContextType {
  registerRefresh: (fn: () => void | Promise<void>) => void
  triggerRefresh: () => void
}

const RefreshContext = createContext<RefreshContextType | null>(null)

export function RefreshProvider({ children }: { children: ReactNode }) {
  const [refreshFn, setRefreshFn] = useState<() => void | Promise<void>>(() => {})

  function registerRefresh(fn: () => void | Promise<void>) {
    setRefreshFn(() => fn)
  }

  function triggerRefresh() {
    refreshFn()
  }

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