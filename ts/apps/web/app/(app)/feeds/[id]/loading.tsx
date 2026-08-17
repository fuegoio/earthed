import { Skeleton } from "@workspace/ui/components/skeleton";
import { EntryCardSkeleton } from "@/components/entry-card-skeleton";

export default function FeedLoading() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      {/* Header matching FeedDetail's sticky PageHeader */}
      <div className="sticky top-0 z-10 bg-background">
        <div className="border-b border-border px-4 py-3">
          <div className="flex items-center gap-2">
            <Skeleton className="-ml-2 size-9 shrink-0 rounded-md lg:hidden" />
            <Skeleton className="size-5 shrink-0 rounded-md" />
            <Skeleton className="h-5 w-40 flex-1" />
            {/* Actions: rename, trash, external link, rss — 4 × icon-sm (size-8) */}
            <div className="flex shrink-0 items-center gap-1">
              <Skeleton className="size-8 rounded-md" />
              <Skeleton className="size-8 rounded-md" />
              <Skeleton className="size-8 rounded-md" />
              <Skeleton className="size-8 rounded-md" />
            </div>
          </div>
          <Skeleton className="mt-1.5 h-4 w-56 ml-[52px] lg:ml-0" />
        </div>
        {/* Button row matching feed-detail's actions bar */}
        <div className="flex items-center gap-2 border-b border-border px-4 py-2 pl-[52px] lg:pl-4">
          <Skeleton className="h-9 w-32 rounded-md" />
          <Skeleton className="h-9 w-20 rounded-md" />
          <Skeleton className="h-9 w-28 rounded-md" />
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
