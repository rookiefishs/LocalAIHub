'use client'

import Image from 'next/image'

interface LogoMarkProps {
  className?: string
}

export function LogoMark({ className }: LogoMarkProps) {
  return (
    <Image
      src="/logo.png"
      alt="LocalAIHub"
      width={64}
      height={64}
      className={className}
    />
  )
}
