import { cn } from "@workspace/ui/lib/utils"

/**
 * Simple empty-state surface. Composed inline (per the conventions there is no
 * shared shadcn Empty installed yet) — kept minimal and composable.
 */
export function Empty({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border px-6 py-16 text-center",
        className
      )}
    >
      {children}
    </div>
  )
}

export function EmptyTitle({ children }: { children: React.ReactNode }) {
  return <p className="font-serif text-lg font-bold tracking-tight">{children}</p>
}

export function EmptyDescription({ children }: { children: React.ReactNode }) {
  return <p className="max-w-sm text-sm text-muted-foreground">{children}</p>
}
