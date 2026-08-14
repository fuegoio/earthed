"use client"

import { useEffect, useRef, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { ExternalLink, ArrowLeft, Circle } from "lucide-react"
import Link from "next/link"
import { Button, buttonVariants } from "@workspace/ui/components/button"
import { StarToggle } from "@/components/star-toggle"
import { FeedIcon } from "@/components/feed-icon"
import { getClient, updateEntries } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { formatDateTime } from "@/lib/format"
import { cn } from "@workspace/ui/lib/utils"
import type { Entry, Feed } from "@/lib/types"

/**
 * Reading view for a single entry. Renders sanitized HTML content (the API
 * already sanitizes via bluemonday). Marks the entry read on open if it was
 * unread, and exposes mark-unread + star actions.
 */
export function EntryReader({
  entry: initialEntry,
  feed,
}: {
  entry: Entry
  feed?: Feed
}) {
  const queryClient = useQueryClient()
  const [entry, setEntry] = useState(initialEntry)
  const markedRef = useRef(false)

  // Mark as read on open (one-time side effect). Idempotent, so StrictMode
  // double-invoke in dev is harmless.
  useEffect(() => {
    if (markedRef.current) return
    markedRef.current = true
    if (entry.status !== "unread") return
    void (async () => {
      const { error } = await updateEntries({
        client: await getClient(),
        body: { entry_ids: [entry.id], status: "read" },
      })
      if (error) {
        toast.error(getApiErrorMessage(error, "Could not mark as read"))
        return
      }
      setEntry((e) => ({ ...e, status: "read" }))
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
    })()
  }, [entry.id, entry.status, queryClient])

  async function handleMarkUnread() {
    setEntry((e) => ({ ...e, status: "unread" }))
    const { error } = await updateEntries({
      client: await getClient(),
      body: { entry_ids: [entry.id], status: "unread" },
    })
    if (error) {
      toast.error(getApiErrorMessage(error, "Could not mark as unread"))
      setEntry((e) => ({ ...e, status: "read" }))
      return
    }
    await queryClient.invalidateQueries({ queryKey: ["entries"] })
  }

  return (
    <article className="mx-auto w-full max-w-2xl px-4 py-6 sm:px-6 sm:py-10">
      <Link
        href="/"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="size-3.5" />
        Back
      </Link>

      <header className="flex flex-col gap-3">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {feed && (
            <>
              <FeedIcon siteUrl={feed.site_url} className="size-4 rounded-sm" />
              <span className="truncate font-medium text-foreground">
                {feed.title}
              </span>
            </>
          )}
        </div>
        <h1 className="font-serif text-2xl font-bold leading-tight tracking-tight sm:text-3xl">
          {entry.title || "Untitled"}
        </h1>
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground">
          {entry.author && <span>{entry.author}</span>}
          {entry.author && <span aria-hidden>·</span>}
          <time>{formatDateTime(entry.published_at)}</time>
        </div>
        <div className="flex items-center gap-2">
          {entry.url && (
            <a
              href={entry.url}
              target="_blank"
              rel="noopener noreferrer"
              className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
            >
              <ExternalLink className="size-3.5" />
              Open original
            </a>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={handleMarkUnread}
            disabled={entry.status === "unread"}
          >
            <Circle className="size-3.5" />
            Mark unread
          </Button>
          <div className="ml-auto">
            <StarToggle
              entryId={entry.id}
              starred={entry.starred}
              size="icon-sm"
            />
          </div>
        </div>
      </header>

      {entry.tags && entry.tags.length > 0 && (
        <div className="mt-4 flex flex-wrap gap-1.5">
          {entry.tags.map((tag) => (
            <span
              key={tag}
              className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground"
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      <div
        className="typeset mt-6 max-w-none"
        // Content is sanitized server-side by the API (bluemonday).
        dangerouslySetInnerHTML={{ __html: (entry.content || entry.description) ?? "" }}
      />
    </article>
  )
}
