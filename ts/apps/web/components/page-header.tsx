import type { ReactNode } from "react"
import { cn } from "@workspace/ui/lib/utils"

/**
 * Shared page header used across the app for a consistent title bar.
 *
 * Renders a single border-bottom row with an optional leading icon, a
 * truncated serif title, and optional trailing actions. An optional
 * metadata line sits beneath the title row for secondary context (site
 * URL, feed count, author/date, etc.). Pages wrap this in their
 * `mx-auto w-full max-w-3xl` container.
 */
export function PageHeader({
  title,
  icon,
  actions,
  metadata,
  className,
}: {
  title: ReactNode
  icon?: ReactNode
  actions?: ReactNode
  metadata?: ReactNode
  className?: string
}) {
  return (
    <div className={cn("sticky top-0 z-10 border-b border-border bg-background px-4 py-3", className)}>
      <div className="flex items-center justify-between gap-3">
        <h1 className={cn(
          "flex min-w-0 items-center gap-2 font-serif text-lg font-bold tracking-tight",
          !icon && "pl-8"
        )}>
          {icon}
          <span className="truncate">{title}</span>
        </h1>
        {actions ? (
          <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
            {actions}
          </div>
        ) : null}
      </div>
      {metadata ? (
        <div className="mt-1.5 text-sm text-muted-foreground">{metadata}</div>
      ) : null}
    </div>
  )
}
