import { Skeleton } from "@workspace/ui/components/skeleton";
import { EntryCardSkeleton } from "@/components/entry-card-skeleton";

export default function FeedLoading() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      {/* Header matching FeedDetail layout */}
      <div className="flex flex-col gap-3 border-b border-border px-4 py-4">
        <div className="flex items-start gap-3">
          <Skeleton className="size-8 shrink-0 rounded-lg" />
          <div className="min-w-0 flex-1">
            <Skeleton className="h-5 w-48" />
            <Skeleton className="mt-1.5 h-3 w-32" />
          </div>
        </div>
        {/* Button row matching the actual actions */}
        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-24 rounded-md" />
          <Skeleton className="h-8 w-32 rounded-md" />
          <Skeleton className="h-8 w-20 rounded-md" />
        </div>
      </div>
      {/* Entry skeletons */}
      <div className="flex flex-col divide-y divide-border">
        {Array.from({ length: 6 }).map((_, i) => (
          <EntryCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}
