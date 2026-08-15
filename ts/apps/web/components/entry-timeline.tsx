"use client"

import { useEffect, useRef } from "react"
import Link from "next/link"
import {
  useInfiniteQuery,
  useQuery,
  type InfiniteData,
} from "@tanstack/react-query"
import { Rss } from "lucide-react"
import { Skeleton } from "@workspace/ui/components/skeleton"
import { EntryCard } from "@/components/entry-card"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { getClient, listEntries, listFeeds, unwrap } from "@/lib/planetary"
import type { Entry, Feed } from "@/lib/types"
import { buttonVariants } from "@workspace/ui/components/button"
import { cn } from "@workspace/ui/lib/utils"

export type EntryFilter = {
  feed_id?: number
  folder_id?: number
  status?: "unread" | "read" | "removed"
  starred?: boolean
  search?: string
}

const PAGE_SIZE = 50

/**
 * Filtered, paginated entry list with infinite scroll. Fetches the user's feeds
 * in parallel so each card can show the owning feed's favicon + title.
 */
export function EntryTimeline({
  filter,
  emptyTitle = "Nothing here yet",
  emptyDescription = "Subscribe to feeds and your latest articles will appear here.",
}: {
  filter: EntryFilter
  emptyTitle?: string
  emptyDescription?: string
}) {
  const { data: feeds } = useQuery<Feed[]>({
    queryKey: ["feeds"],
    queryFn: async () => unwrap(listFeeds({ client: await getClient() })),
  })
  const feedMap = new Map<number, Feed>()
  for (const f of feeds ?? []) feedMap.set(f.id, f)

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    error,
    refetch,
  } = useInfiniteQuery<
    Entry[],
    Error,
    InfiniteData<Entry[]>,
    ["entries", EntryFilter],
    number
  >({
    queryKey: ["entries", filter],
    queryFn: async ({ pageParam }) => {
      const result = await listEntries({
        client: await getClient(),
        query: { ...filter, limit: PAGE_SIZE, offset: pageParam },
      })
      if (result.error) throw result.error
      return (result.data ?? []) as Entry[]
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage, _all, lastParam) =>
      lastPage.length < PAGE_SIZE ? undefined : lastParam + PAGE_SIZE,
  })

  const sentinelRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const el = sentinelRef.current
    if (!el || !hasNextPage || isFetchingNextPage) return
    const obs = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) fetchNextPage()
      },
      { rootMargin: "600px" }
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  const entries = data?.pages.flat() ?? []

  if (isLoading) {
    return (
      <div className="divide-y divide-border">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="flex gap-3 px-4 py-3">
            <Skeleton className="size-10 shrink-0 rounded-lg" />
            <div className="min-w-0 flex-1 space-y-2">
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-3 w-1/2" />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4">
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Rss className="size-6 text-primary" />
            </EmptyMedia>
            <EmptyTitle>Couldn&apos;t load entries</EmptyTitle>
            <EmptyDescription>
              Something went wrong fetching your timeline. Try again.
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <button
              onClick={() => refetch()}
              className={cn(buttonVariants({ size: "sm" }))}
            >
              Retry
            </button>
          </EmptyContent>
        </Empty>
      </div>
    )
  }

  if (entries.length === 0) {
    return (
      <div className="p-4">
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Rss className="size-6 text-primary" />
            </EmptyMedia>
            <EmptyTitle>{emptyTitle}</EmptyTitle>
            <EmptyDescription>{emptyDescription}</EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <Link
              href="/feeds/new"
              className={cn(buttonVariants({ size: "sm" }))}
            >
              Subscribe to a feed
            </Link>
          </EmptyContent>
        </Empty>
      </div>
    )
  }

  return (
    <div className="divide-y divide-border">
      {entries.map((entry) => (
        <EntryCard
          key={entry.id}
          entry={entry}
          feed={feedMap.get(entry.feed_id)}
        />
      ))}
      <div ref={sentinelRef} className="h-px" />
      {isFetchingNextPage && (
        <div className="divide-y divide-border">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="flex gap-3 px-4 py-3">
              <Skeleton className="size-10 shrink-0 rounded-lg" />
              <div className="min-w-0 flex-1 space-y-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
