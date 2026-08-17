import { Menu } from "lucide-react";
import { Skeleton } from "@workspace/ui/components/skeleton";
import { Button } from "@workspace/ui/components/button";
import { EntryCardSkeleton } from "@/components/entry-card-skeleton";

export default function Loading() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      {/* Matches PageHeader: sticky top-0 border-b px-4 py-3 */}
      <div className="sticky top-0 z-10 border-b border-border bg-background px-4 py-3">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" disabled aria-hidden="true" className="-ml-2 shrink-0 lg:hidden">
            <Menu className="size-4" />
          </Button>
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
