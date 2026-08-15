import { Skeleton } from "@workspace/ui/components/skeleton"

/**
 * Loading placeholder that mirrors the EntryCard layout: read/unread dot,
 * metadata row (favicon + feed title + time), two-line title, two-line
 * snippet, and a right-side action button.
 */
export function EntryCardSkeleton() {
  return (
    <div className="flex gap-3 px-4 py-3">
      {/* read/unread dot */}
      <div className="flex size-5 shrink-0 items-start justify-center pt-1">
        <Skeleton className="size-2 rounded-full" />
      </div>
      {/* content */}
      <div className="min-w-0 flex-1">
        {/* metadata row */}
        <div className="flex items-center gap-2">
          <Skeleton className="size-3.5 shrink-0 rounded-sm" />
          <Skeleton className="h-3 w-24" />
          <span aria-hidden className="text-xs text-muted-foreground">
            ·
          </span>
          <Skeleton className="h-3 w-12" />
        </div>
        {/* title */}
        <Skeleton className="mt-1 h-4 w-3/4" />
        {/* snippet (two lines) */}
        <Skeleton className="mt-1 h-3 w-full" />
        <Skeleton className="mt-1 h-3 w-1/2" />
      </div>
      {/* right action */}
      <div className="flex shrink-0 items-start gap-0.5">
        <Skeleton className="size-8 rounded-md" />
      </div>
    </div>
  )
}
