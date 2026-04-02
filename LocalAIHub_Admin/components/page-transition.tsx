'use client'

import { ReactNode } from 'react'
import { usePathname } from 'next/navigation'
import { motion } from 'framer-motion'

export function PageTransition({ children }: { children: ReactNode }) {
  const pathname = usePathname()

  return (
    <motion.div
      key={pathname}
      initial={{ opacity: 0, y: 20, scale: 0.98 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ duration: 0.3, ease: [0.25, 0.46, 0.45, 0.94] }}
      className="min-h-full"
    >
      {children}
    </motion.div>
  )
}
