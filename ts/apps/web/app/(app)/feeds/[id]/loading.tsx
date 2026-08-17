import {
  Menu,
  Pencil,
  Trash2,
  ExternalLink,
  Rss,
  CheckCheck,
  RefreshCw,
  FolderOpen,
  ChevronDown,
} from "lucide-react";
import { Skeleton } from "@workspace/ui/components/skeleton";
import { Button, buttonVariants } from "@workspace/ui/components/button";
import { EntryCardSkeleton } from "@/components/entry-card-skeleton";
import { cn } from "@workspace/ui/lib/utils";

export default function FeedLoading() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      {/* Header matching FeedDetail's sticky PageHeader */}
      <div className="sticky top-0 z-10 bg-background">
        <div className="border-b border-border px-4 py-3">
          <div className="flex items-center justify-between gap-3">
            <div className="flex min-w-0 flex-1 items-center gap-2">
              {/* Menu button — mobile only, matches PageHeader */}
              <Button
                variant="ghost"
                size="icon"
                disabled
                aria-hidden="true"
                className="-ml-2 shrink-0 lg:hidden"
              >
                <Menu className="size-4" />
              </Button>
              {/* Feed favicon placeholder */}
              <Skeleton className="size-5 shrink-0 rounded-md" />
              {/* Feed title */}
              <Skeleton className="h-5 w-40" />
            </div>
            {/* Actions: rename, delete, external link, rss */}
            <div className="flex shrink-0 items-center gap-1">
              <Button variant="ghost" size="icon-sm" disabled aria-hidden="true" className="text-muted-foreground">
                <Pencil className="size-3.5" />
              </Button>
              <Button variant="ghost" size="icon-sm" disabled aria-hidden="true" className="text-muted-foreground">
                <Trash2 className="size-3.5" />
              </Button>
              <Button variant="ghost" size="icon-sm" disabled aria-hidden="true" className="text-muted-foreground">
                <ExternalLink className="size-3.5" />
              </Button>
              <Button variant="ghost" size="icon-sm" disabled aria-hidden="true" className="text-muted-foreground">
                <Rss className="size-3.5" />
              </Button>
            </div>
          </div>
          {/* Description line */}
          <Skeleton className="mt-1.5 h-4 w-56 ml-[52px] lg:ml-0" />
        </div>
        {/* Action bar: mark all read, refresh, folder picker */}
        <div className="flex items-center gap-2 border-b border-border px-4 py-2 pl-[52px] lg:pl-4">
          <Button variant="outline" size="sm" disabled aria-hidden="true">
            <CheckCheck className="size-3.5" />
            Mark all as read
          </Button>
          <Button variant="outline" size="sm" disabled aria-hidden="true">
            <RefreshCw className="size-3.5" />
            Refresh
          </Button>
          <button
            disabled
            aria-hidden="true"
            className={cn(buttonVariants({ variant: "outline", size: "sm" }), "cursor-default")}
          >
            <FolderOpen className="size-3.5" />
            <Skeleton className="h-3.5 w-16" />
            <ChevronDown className="size-3 text-muted-foreground" />
          </button>
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
