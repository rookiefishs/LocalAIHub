"use client"

import { cn } from "@/lib/utils"
import { CircleIcon } from "lucide-react"

interface UpstreamHealthCardProps {
  name: string
  status: "healthy" | "degraded" | "down"
  latency: number
  uptime: number
}

const statusConfig = {
  healthy: {
    label: "正常",
    color: "text-green-500",
    bgColor: "bg-green-500",
  },
  degraded: {
    label: "降级",
    color: "text-yellow-500",
    bgColor: "bg-yellow-500",
  },
  down: {
    label: "离线",
    color: "text-red-500",
    bgColor: "bg-red-500",
  },
}

export function UpstreamHealthCard({
  name,
  status,
  latency,
  uptime,
}: UpstreamHealthCardProps) {
  const config = statusConfig[status]

  return (
    <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/30 border border-border/50">
      <div className="flex items-center gap-3">
        <CircleIcon
          className={cn("h-2.5 w-2.5", config.color)}
          fill="currentColor"
        />
        <div>
          <p className="text-sm font-medium">{name}</p>
          <p className="text-xs text-muted-foreground">{config.label}</p>
        </div>
      </div>
      <div className="flex items-center gap-4 text-right">
        <div>
          <p className="text-sm font-mono">
            {status === "down" ? "-" : `${latency}ms`}
          </p>
          <p className="text-xs text-muted-foreground">延迟</p>
        </div>
        <div>
          <p className="text-sm font-mono">
            {status === "down" ? "-" : `${uptime}%`}
          </p>
          <p className="text-xs text-muted-foreground">可用性</p>
        </div>
      </div>
    </div>
  )
}
