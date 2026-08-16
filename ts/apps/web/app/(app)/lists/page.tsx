"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { ListChecks, Plus, Globe, Users, Compass } from "lucide-react";
import { Skeleton } from "@workspace/ui/components/skeleton";
import { buttonVariants } from "@workspace/ui/components/button";
import { PageHeader } from "@/components/page-header";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty";
import { getClient, listMyFeedLists, listFollowedFeedLists, unwrap } from "@/lib/planetary";
import { cn } from "@workspace/ui/lib/utils";
import type { FeedList } from "@/lib/types";

export default function FeedListsPage() {
  const { data: mine, isLoading: mineLoading } = useQuery<FeedList[]>({
    queryKey: ["feed-lists", "mine"],
    queryFn: async () => unwrap(listMyFeedLists({ client: await getClient() })),
  });
  const { data: followed, isLoading: followedLoading } = useQuery<FeedList[]>({
    queryKey: ["feed-lists", "followed"],
    queryFn: async () => unwrap(listFollowedFeedLists({ client: await getClient() })),
  });

  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader
        title="Feed lists"
        icon={<ListChecks className="size-4 text-muted-foreground" />}
        actions={
          <>
            <Link
              href="/lists/discover"
              className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
            >
              <Compass className="size-4" />
              Discover
            </Link>
            <Link href="/lists/new" className={cn(buttonVariants({ size: "sm" }))}>
              <Plus className="size-4" />
              New list
            </Link>
          </>
        }
        metadata="Curated collections of feeds you can share and follow."
      />

      {/* My lists */}
      <section className="p-4">
        <h2 className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          My lists
        </h2>
        {mineLoading ? (
          <LoadingRow />
        ) : (mine ?? []).length === 0 ? (
          <Empty className="border py-10">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ListChecks className="size-6 text-primary" />
              </EmptyMedia>
              <EmptyTitle>No lists yet</EmptyTitle>
              <EmptyDescription>
                Create a list, add feeds to it, and share it publicly so others can follow.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Link href="/lists/new" className={cn(buttonVariants({ size: "sm" }))}>
                <Plus className="size-4" />
                Create a list
              </Link>
            </EmptyContent>
          </Empty>
        ) : (
          <div className="flex flex-col gap-2">
            {(mine ?? []).map((list) => (
              <FeedListRow key={list.id} list={list} mine />
            ))}
          </div>
        )}
      </section>

      {/* Followed lists */}
      <section className="p-4 pt-0">
        <h2 className="mb-2 flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <Users className="size-3.5" />
          Lists you follow
        </h2>
        {followedLoading ? (
          <LoadingRow />
        ) : (followed ?? []).length === 0 ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Users className="size-6 text-primary" />
              </EmptyMedia>
              <EmptyTitle>Not following any lists</EmptyTitle>
              <EmptyDescription>
                Discover public lists curated by others and follow them to keep them bookmarked.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Link
                href="/lists/discover"
                className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
              >
                <Compass className="size-4" />
                Discover lists
              </Link>
            </EmptyContent>
          </Empty>
        ) : (
          <div className="flex flex-col gap-2">
            {(followed ?? []).map((list) => (
              <FeedListRow key={list.id} list={list} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function LoadingRow() {
  return (
    <div className="flex flex-col gap-2">
      {Array.from({ length: 2 }).map((_, i) => (
        <div key={i} className="flex items-center gap-3 rounded-lg border border-border p-4">
          <Skeleton className="size-10 shrink-0 rounded-lg" />
          <div className="min-w-0 flex-1 space-y-2">
            <Skeleton className="h-4 w-40" />
            <Skeleton className="h-3 w-24" />
          </div>
        </div>
      ))}
    </div>
  );
}

function FeedListRow({ list, mine = false }: { list: FeedList; mine?: boolean }) {
  return (
    <Link
      href={`/lists/${list.id}`}
      className="flex items-center gap-3 rounded-lg border border-border p-4 transition-colors hover:bg-muted/50"
    >
      <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
        <ListChecks className="size-5 text-primary" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p className="truncate font-medium">{list.title}</p>
          {list.is_public ? (
            <span className="flex shrink-0 items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
              <Globe className="size-3" />
              Public
            </span>
          ) : (
            <span className="shrink-0 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
              Private
            </span>
          )}
        </div>
        {list.description && (
          <p className="line-clamp-1 text-sm text-muted-foreground">{list.description}</p>
        )}
        <p className="mt-0.5 text-xs text-muted-foreground">
          {list.feed_count} {list.feed_count === 1 ? "feed" : "feeds"}
          {!mine && list.owner_email ? ` · by ${list.owner_email}` : ""}
        </p>
      </div>
    </Link>
  );
}
