'use client'

interface LogoMarkProps {
  className?: string
}

export function LogoMark({ className }: LogoMarkProps) {
  return (
    <img
      src="/localaihub-admin/logo.png"
      alt="LocalAIHub"
      width={64}
      height={64}
      className={className}
    />
  )
}
