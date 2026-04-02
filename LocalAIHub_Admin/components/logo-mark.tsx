'use client'

import { useMemo } from 'react'

interface LogoMarkProps {
  className?: string
}

export function LogoMark({ className }: LogoMarkProps) {
  const src = useMemo(() => {
    const basePath = process.env.NODE_ENV === 'development' ? '' : (process.env.NEXT_PUBLIC_BASE_PATH || '/localaihub-admin')
    return `${basePath}/logo.png`
  }, [])

  return (
    <img
      src={src}
      alt="LocalAIHub"
      width={64}
      height={64}
      className={className}
    />
  )
}
