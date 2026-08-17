import { Skeleton } from "@workspace/ui/components/skeleton";
import { EntryCardSkeleton } from "@/components/entry-card-skeleton";

export default function FeedLoading() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      {/* Header matching FeedDetail's sticky PageHeader */}
      <div className="sticky top-0 z-10 bg-background">
        <div className="flex flex-col gap-3 border-b border-border px-4 py-3">
          <div className="flex items-center gap-2">
            <Skeleton className="size-9 shrink-0 rounded-md lg:hidden" />
            <Skeleton className="size-5 shrink-0 rounded-md" />
            <Skeleton className="h-5 w-48" />
          </div>
          {/* Button row matching the actual actions */}
          <div className="flex items-center gap-2">
            <Skeleton className="h-7 w-20 rounded-md" />
            <Skeleton className="h-7 w-24 rounded-md" />
            <Skeleton className="h-7 w-7 rounded-md" />
          </div>
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
