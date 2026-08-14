"use client"

import Link from "next/link"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Compass, Globe, Loader2, UserPlus, Check } from "lucide-react"
import { Button, buttonVariants } from "@workspace/ui/components/button"
import { Empty, EmptyDescription, EmptyTitle } from "@/components/empty"
import {
  getClient,
  discoverFeedLists,
  followFeedList,
  unwrap,
} from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { cn } from "@workspace/ui/lib/utils"
import type { FeedList } from "@/lib/types"

export default function DiscoverPage() {
  const queryClient = useQueryClient()
  const { data: lists, isLoading } = useQuery<FeedList[]>({
    queryKey: ["feed-lists", "discover"],
    queryFn: async () =>
      unwrap(discoverFeedLists({ client: await getClient() })),
  })

  async function handleFollow(list: FeedList) {
    const { error } = await followFeedList({
      client: await getClient(),
      path: { listId: list.id },
    })
    if (error) {
      toast.error(getApiErrorMessage(error, "Could not follow list"))
      return
    }
    toast.success(`Following "${list.title}"`)
    await queryClient.invalidateQueries({ queryKey: ["feed-lists"] })
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-6">
      <div className="flex items-center justify-between">
        <h1 className="flex items-center gap-2 font-serif text-2xl font-bold tracking-tight">
          <Compass className="size-5" />
          Discover
        </h1>
        <Link href="/lists" className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}>
          My lists
        </Link>
      </div>
      <p className="mt-1 text-sm text-muted-foreground">
        Public feed lists shared by the community. Follow one to keep it in your
        sidebar, then import its feeds in a click.
      </p>

      <div className="mt-6 flex flex-col gap-3">
        {isLoading ? (
          <div className="flex items-center gap-2 py-10 text-sm text-muted-foreground">
            <Loader2 className="size-4 animate-spin" />
            Loading public lists…
          </div>
        ) : (lists ?? []).length === 0 ? (
          <Empty>
            <div className="flex size-12 items-center justify-center rounded-xl bg-primary/10">
              <Globe className="size-6 text-primary" />
            </div>
            <EmptyTitle>No public lists yet</EmptyTitle>
            <EmptyDescription>
              Be the first — create a list and make it public so others can
              follow.
            </EmptyDescription>
            <Link
              href="/lists/new"
              className={cn(buttonVariants({ size: "sm" }), "mt-2")}
            >
              Create a list
            </Link>
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
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleFollow(list)}
                >
                  <UserPlus className="size-3.5" />
                  Follow
                </Button>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  )
}
