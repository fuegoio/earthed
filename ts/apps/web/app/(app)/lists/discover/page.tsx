"use client";

import Link from "next/link";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Compass, Globe, UserPlus, Check } from "lucide-react";
import { Skeleton } from "@workspace/ui/components/skeleton";
import { Button, buttonVariants } from "@workspace/ui/components/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty";
import { getClient, discoverFeedLists, followFeedList, unwrap } from "@/lib/earthed";
import { getApiErrorMessage } from "@/lib/errors";
import { cn } from "@workspace/ui/lib/utils";
import type { FeedList } from "@/lib/types";

export default function DiscoverPage() {
  const queryClient = useQueryClient();
  const { data: lists, isLoading } = useQuery<FeedList[]>({
    queryKey: ["feed-lists", "discover"],
    queryFn: async () => unwrap(discoverFeedLists({ client: await getClient() })),
  });

  async function handleFollow(list: FeedList) {
    const { error } = await followFeedList({
      client: await getClient(),
      path: { listId: list.id },
    });
    if (error) {
      toast.error(getApiErrorMessage(error, "Could not follow list"));
      return;
    }
    toast.success(`Following "${list.title}"`);
    await queryClient.invalidateQueries({ queryKey: ["feed-lists"] });
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-6">
      <div className="flex items-center justify-between">
        <h1 className="flex items-center gap-2 font-serif text-2xl font-bold tracking-normal">
          <Compass className="size-5" />
          Discover
        </h1>
        <Link href="/lists" className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}>
          My lists
        </Link>
      </div>
      <p className="mt-1 text-sm text-muted-foreground">
        Public feed lists shared by the community. Follow one to keep it in your sidebar, then
        import its feeds in a click.
      </p>

      <div className="mt-6 flex flex-col gap-3">
        {isLoading ? (
          <div className="flex flex-col gap-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-start gap-3 rounded-lg border border-border p-4">
                <div className="min-w-0 flex-1 space-y-2">
                  <Skeleton className="h-4 w-40" />
                  <Skeleton className="h-3 w-24" />
                </div>
                <Skeleton className="h-8 w-20 rounded-md" />
              </div>
            ))}
          </div>
        ) : (lists ?? []).length === 0 ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Globe className="size-6 text-primary" />
              </EmptyMedia>
              <EmptyTitle>No public lists yet</EmptyTitle>
              <EmptyDescription>
                Be the first — create a list and make it public so others can follow.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Link href="/lists/new" className={cn(buttonVariants({ size: "sm" }))}>
                Create a list
              </Link>
            </EmptyContent>
          </Empty>
        ) : (
          (lists ?? []).map((list) => (
            <div
              key={list.id}
              className="flex items-start gap-3 rounded-lg border border-border p-4"
            >
              <div className="min-w-0 flex-1">
                <Link
                  href={`/lists/${list.id}`}
                  className="font-medium hover:text-primary transition-colors"
                >
                  {list.title}
                </Link>
                {list.description && (
                  <p className="mt-0.5 line-clamp-2 text-sm text-muted-foreground">
                    {list.description}
                  </p>
                )}
                <p className="mt-1 text-xs text-muted-foreground">
                  {list.feed_count} {list.feed_count === 1 ? "feed" : "feeds"}
                  {list.owner_email ? ` · by ${list.owner_email}` : ""}
                </p>
              </div>
              {list.is_following ? (
                <Button variant="outline" size="sm" disabled>
                  <Check className="size-3.5" />
                  Following
                </Button>
              ) : (
                <Button variant="outline" size="sm" onClick={() => handleFollow(list)}>
                  <UserPlus className="size-3.5" />
                  Follow
                </Button>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
