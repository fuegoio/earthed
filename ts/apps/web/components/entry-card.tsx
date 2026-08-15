"use client"

import { useState, useOptimistic, startTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { StarToggle } from "@/components/star-toggle"
import { LikeToggle } from "@/components/like-toggle"
import { FeedIcon } from "@/components/feed-icon"
import { getClient, updateEntries } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { formatRelative, htmlSnippet } from "@/lib/format"
import { cn } from "@workspace/ui/lib/utils"
import type { Entry, Feed } from "@/lib/types"

/**
 * A single entry row in a timeline. The entire row is a link that opens
 * the original article in a new tab (and marks the entry as read). The
 * leading dot toggles read/unread on click. Unread entries get a solid
 * primary dot; read entries show a faded dot on hover so they can be
 * marked unread again.
 */
export function EntryCard({ entry, feed }: { entry: Entry; feed?: Feed }) {
  const queryClient = useQueryClient()
  const [readOptimistic, setReadOptimistic] = useOptimistic(entry.status === "read")
  const [pending, setPending] = useState(false)

  const unread = !readOptimistic
  const snippet = htmlSnippet(entry.description, 200)

  function toggleRead(e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    const next = readOptimistic ? "unread" : "read"
    startTransition(() => {
      setReadOptimistic(next === "read")
      setPending(true)
    })
    void (async () => {
      const { error } = await updateEntries({
        client: await getClient(),
        body: { entry_ids: [entry.id], status: next },
      })
      if (error) {
        toast.error(getApiErrorMessage(error, "Could not update entry"))
        return
      }
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
    })().finally(() => setPending(false))
  }

  function handleClick(e: React.MouseEvent<HTMLAnchorElement>) {
    if (!entry.url) {
      e.preventDefault()
      return
    }
    // Mark as read when opening the article.
    if (unread) {
      startTransition(() => setReadOptimistic(true))
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
      })()
    }
  }

  return (
    <a
      href={entry.url ?? "#"}
      target={entry.url ? "_blank" : undefined}
      rel={entry.url ? "noopener noreferrer" : undefined}
      onClick={handleClick}
      className={cn(
        "group flex gap-3 px-4 py-3 transition-colors hover:bg-muted/50",
        unread ? "" : "opacity-60"
      )}
    >
      <button
        type="button"
        onClick={toggleRead}
        disabled={pending}
        aria-label={unread ? "Mark as read" : "Mark as unread"}
        aria-pressed={unread}
        className="flex size-5 shrink-0 items-start justify-center pt-1"
      >
        {unread ? (
          <span className="size-2 rounded-full bg-primary transition-opacity" />
        ) : (
          <span className="size-2 rounded-full bg-muted-foreground opacity-0 transition-opacity group-hover:opacity-50" />
        )}
      </button>
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
      <div
        className="flex shrink-0 items-start gap-0.5"
        onClick={(e) => e.stopPropagation()}
      >
        <LikeToggle entryId={entry.id} liked={entry.liked} size="icon-sm" />
        <StarToggle entryId={entry.id} starred={entry.starred} size="icon-sm" />
      </div>
    </a>
  )
}
