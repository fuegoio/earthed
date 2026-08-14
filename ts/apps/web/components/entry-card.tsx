"use client"

import Link from "next/link"
import { StarToggle } from "@/components/star-toggle"
import { LikeToggle } from "@/components/like-toggle"
import { FeedIcon } from "@/components/feed-icon"
import { formatRelative, htmlSnippet } from "@/lib/format"
import { cn } from "@workspace/ui/lib/utils"
import type { Entry, Feed } from "@/lib/types"

/**
 * A single entry row in a timeline. Links to the reading view. Shows the feed
 * favicon + title, relative publish time, title, a content snippet, and a star
 * toggle. Unread entries get a brighter title + a leading dot.
 */
export function EntryCard({ entry, feed }: { entry: Entry; feed?: Feed }) {
  const unread = entry.status === "unread"
  const snippet =
    htmlSnippet(entry.description, 200) || htmlSnippet(entry.content, 200)

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
        <LikeToggle entryId={entry.id} liked={entry.liked} size="icon-sm" />
        <StarToggle entryId={entry.id} starred={entry.starred} size="icon-sm" />
      </div>
    </Link>
  )
}
