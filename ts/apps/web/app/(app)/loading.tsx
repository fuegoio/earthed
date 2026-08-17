import { Skeleton } from "@workspace/ui/components/skeleton";
import { EntryCardSkeleton } from "@/components/entry-card-skeleton";

export default function Loading() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      {/* Matches PageHeader: sticky top-0 border-b px-4 py-3 */}
      <div className="sticky top-0 z-10 border-b border-border bg-background px-4 py-3">
        <div className="flex items-center gap-2">
          <Skeleton className="-ml-2 size-9 shrink-0 rounded-md lg:hidden" />
          <Skeleton className="h-6 w-32" />
        </div>
      </div>
      <div className="flex flex-col divide-y divide-border">
        {Array.from({ length: 6 }).map((_, i) => (
          <EntryCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}
