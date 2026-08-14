"use client"

import { useState, useOptimistic, startTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { ExternalLink, ArrowLeft, Circle, CheckCircle } from "lucide-react"
import Link from "next/link"
import { Button, buttonVariants } from "@workspace/ui/components/button"
import { StarToggle } from "@/components/star-toggle"
import { LikeToggle } from "@/components/like-toggle"
import { FeedIcon } from "@/components/feed-icon"
import { getClient, updateEntries } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { formatDateTime } from "@/lib/format"
import { cn } from "@workspace/ui/lib/utils"
import type { Entry, Feed } from "@/lib/types"

/**
 * Reading view for a single entry. Hacker News-style: just the title,
 * metadata, and a short description. No full article body is rendered.
 * Does NOT auto-mark as read on open; the user must explicitly mark or
 * open the link.
 */
export function EntryReader({
  entry: initialEntry,
  feed,
}: {
  entry: Entry
  feed?: Feed
}) {
  const queryClient = useQueryClient()
  const [status, setStatus] = useState(initialEntry.status)
  const [optimisticStatus, setOptimisticStatus] = useOptimistic(status)
  const [pending, setPending] = useState(false)

  function handleToggleRead() {
    const next = optimisticStatus === "unread" ? "read" : "unread"
    startTransition(() => {
      setOptimisticStatus(next)
      setPending(true)
    })
    void (async () => {
      const { error } = await updateEntries({
        client: await getClient(),
        body: { entry_ids: [initialEntry.id], status: next },
      })
      if (error) {
        toast.error(getApiErrorMessage(error, "Could not update entry"))
        return
      }
      setStatus(next)
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
    })().finally(() => setPending(false))
  }

  const isUnread = optimisticStatus === "unread"

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
          {initialEntry.title || "Untitled"}
        </h1>
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground">
          {initialEntry.author && <span>{initialEntry.author}</span>}
          {initialEntry.author && <span aria-hidden>·</span>}
          <time>{formatDateTime(initialEntry.published_at)}</time>
        </div>
        <div className="flex items-center gap-2">
          {initialEntry.url && (
            <a
              href={initialEntry.url}
              target="_blank"
              rel="noopener noreferrer"
              className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
            >
              <ExternalLink className="size-3.5" />
              Open
            </a>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={handleToggleRead}
            disabled={pending}
          >
            {isUnread ? (
              <>
                <Circle className="size-3.5" />
                Mark as read
              </>
            ) : (
              <>
                <CheckCircle className="size-3.5" />
                Mark unread
              </>
            )}
          </Button>
          <div className="ml-auto flex items-center gap-0.5">
            <LikeToggle
              entryId={initialEntry.id}
              liked={initialEntry.liked}
              size="icon-sm"
            />
            <StarToggle
              entryId={initialEntry.id}
              starred={initialEntry.starred}
              size="icon-sm"
            />
          </div>
        </div>
      </header>

      {initialEntry.tags && initialEntry.tags.length > 0 && (
        <div className="mt-4 flex flex-wrap gap-1.5">
          {initialEntry.tags.map((tag) => (
            <span
              key={tag}
              className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground"
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      {initialEntry.description && (
        <p className="mt-6 text-base leading-relaxed text-muted-foreground">
          {initialEntry.description}
        </p>
      )}
    </article>
  )
}
