"use client"

import { useState, useOptimistic, startTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { ExternalLink } from "lucide-react"
import Link from "next/link"
import { Button } from "@workspace/ui/components/button"
import { StarToggle } from "@/components/star-toggle"
import { LikeToggle } from "@/components/like-toggle"
import { FeedIcon } from "@/components/feed-icon"
import { getClient, updateEntries } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { formatRelative, htmlSnippet } from "@/lib/format"
import { cn } from "@workspace/ui/lib/utils"
import type { Entry, Feed } from "@/lib/types"

/**
 * A single entry row in a timeline. Links to the reading view. Shows the feed
 * favicon + title, relative publish time, title, a description snippet, and
 * star/like toggles. An "open in new tab" button opens the original URL and
 * marks the entry as read. Unread entries get a brighter title + a leading dot.
 */
export function EntryCard({ entry, feed }: { entry: Entry; feed?: Feed }) {
  const queryClient = useQueryClient()
  const [readOptimistic, setReadOptimistic] = useOptimistic(entry.status === "read")
  const [pending, setPending] = useState(false)

  const unread = !readOptimistic
  const snippet = htmlSnippet(entry.description, 200)

  function handleOpenExternal(e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    if (!entry.url) return
    startTransition(() => {
      setReadOptimistic(true)
      setPending(true)
    })
    void (async () => {
      const { error } = await updateEntries({
        client: await getClient(),
        body: { entry_ids: [entry.id], status: "read" },
      })
      if (error) {
        toast.error(getApiErrorMessage(error, "Could not mark as read"))
        return
      }
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
    })().finally(() => setPending(false))
    window.open(entry.url, "_blank", "noopener,noreferrer")
  }

  return (
    <Link
      href={`/entries/${entry.id}`}
      className={cn(
        "group flex gap-3 px-4 py-3 transition-colors hover:bg-muted/50",
        unread ? "" : "opacity-60"
      )}
    >
      <div className="flex w-5 shrink-0 justify-center pt-1">
        {unread && (
          <span
            className="size-2 rounded-full bg-primary"
            aria-label="Unread"
          />
        )}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <FeedIcon siteUrl={feed?.site_url} className="size-3.5 rounded-sm" />
          <span className="truncate">{feed?.title ?? "Unknown feed"}</span>
          <span aria-hidden>·</span>
          <time className="shrink-0">{formatRelative(entry.published_at)}</time>
        </div>
        <h3
          className={cn(
            "mt-1 line-clamp-2 text-sm",
            unread ? "font-semibold text-foreground" : "font-medium"
          )}
        >
          {entry.title || "Untitled"}
        </h3>
        {snippet && (
          <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
            {snippet}
          </p>
        )}
      </div>
      <div className="flex shrink-0 items-start gap-0.5">
        {entry.url && (
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Open in new tab"
            onClick={handleOpenExternal}
            disabled={pending}
          >
            <ExternalLink className="size-3.5" />
          </Button>
        )}
        <LikeToggle entryId={entry.id} liked={entry.liked} size="icon-sm" />
        <StarToggle entryId={entry.id} starred={entry.starred} size="icon-sm" />
      </div>
    </Link>
  )
}
