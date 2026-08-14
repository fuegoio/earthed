import { cn } from "@workspace/ui/lib/utils"

/**
 * Empty-state surface. Composed inline per the conventions in AGENTS.md:
 * `Empty > EmptyHeader > { EmptyMedia, EmptyTitle, EmptyDescription }` then
 * `EmptyContent` for the primary action.
 *
 * Use the `border` utility on `Empty` (outlined style) to match surrounding
 * surface density. Each empty state owns the primary CTA for its view.
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
        "flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed border-border px-6 py-16 text-center",
        className
      )}
    >
      {children}
    </div>
  )
}

export function EmptyHeader({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("flex flex-col items-center gap-3", className)}>
      {children}
    </div>
  )
}

export function EmptyMedia({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        "flex size-12 items-center justify-center rounded-xl bg-primary/10",
        className
      )}
    >
      {children}
    </div>
  )
}

export function EmptyTitle({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <p
      className={cn("font-serif text-lg font-bold tracking-tight", className)}
    >
      {children}
    </p>
  )
}

export function EmptyDescription({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <p className={cn("max-w-sm text-sm text-muted-foreground", className)}>
      {children}
    </p>
  )
}

export function EmptyContent({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return <div className={cn("mt-1", className)}>{children}</div>
}
