import { cn } from "@workspace/ui/lib/utils"

interface ApiErrorProps {
  message: string
  /** HTTP status from the failed response, if known. */
  status?: number
  className?: string
}

/**
 * Server component rendered in place of API-backed content when the request
 * failed. Server-side fetch errors surface here (not as a toast — server
 * components can't call `toast()`), per the conventions in `apps/web/AGENTS.md`.
 */
export function ApiError({ message, status, className }: ApiErrorProps) {
  return (
    <div
      role="alert"
      className={cn(
        "rounded-md border border-destructive/50 bg-destructive/10 p-4 text-foreground",
        className
      )}
    >
      <div className="flex items-center gap-2 font-medium">
        <span className="text-destructive">Something went wrong</span>
        {status && status > 0 ? (
          <span className="rounded-full border border-border px-1.5 py-0.5 text-xs text-muted-foreground">
            {status}
          </span>
        ) : null}
      </div>
      <p className="mt-1 text-sm text-muted-foreground">{message}</p>
    </div>
  )
}
