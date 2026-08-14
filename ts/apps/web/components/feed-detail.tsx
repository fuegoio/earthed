"use client"

import { useState, useEffect, useRef } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { ExternalLink, Trash2, CheckCheck, Loader2, RefreshCw } from "lucide-react"
import { Button, buttonVariants } from "@workspace/ui/components/button"
import { ConfirmDialog } from "@/components/confirm-dialog"
import { FeedIcon } from "@/components/feed-icon"
import { EntryTimeline } from "@/components/entry-timeline"
import { getClient, markFeedRead, deleteFeed, refreshFeed } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { cn } from "@workspace/ui/lib/utils"
import type { Feed } from "@/lib/types"

/**
 * Feed detail view: header with site link, mark-all-read, refresh, and delete
 * actions, plus the feed's entry timeline. The feed is refreshed on mount so
 * the latest articles are fetched without waiting for the scheduler.
 */
export function FeedDetail({ feed }: { feed: Feed }) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const [marking, setMarking] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const refreshedRef = useRef(false)

  async function handleRefresh() {
    setRefreshing(true)
    try {
      const { error } = await refreshFeed({
        client: await getClient(),
        path: { feedId: feed.id },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
      await queryClient.invalidateQueries({ queryKey: ["feeds"] })
      toast.success("Feed refreshed")
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not refresh feed"))
    } finally {
      setRefreshing(false)
    }
  }

  // Auto-refresh on mount (once per component instance)
  useEffect(() => {
    if (refreshedRef.current) return
    refreshedRef.current = true
    handleRefresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function handleMarkAllRead() {
    setMarking(true)
    try {
      const { error } = await markFeedRead({
        client: await getClient(),
        path: { feedId: feed.id },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
      toast.success(`Marked all entries in "${feed.title}" as read`)
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not mark entries as read"))
    } finally {
      setMarking(false)
    }
  }

  async function handleDelete() {
    setDeleting(true)
    try {
      const { error } = await deleteFeed({
        client: await getClient(),
        path: { feedId: feed.id },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["feeds"] })
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
      toast.success(`Unsubscribed from "${feed.title}"`)
      router.push("/")
      router.refresh()
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not delete feed"))
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="flex flex-col gap-3 border-b border-border px-4 py-4">
        <div className="flex items-start gap-3">
          <FeedIcon
            siteUrl={feed.site_url}
            className="size-8 shrink-0 rounded-lg"
          />
          <div className="min-w-0 flex-1">
            <h1 className="truncate font-serif text-lg font-bold tracking-tight">
              {feed.title || "Untitled feed"}
            </h1>
            {feed.site_url && (
              <a
                href={feed.site_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
              >
                <span className="truncate">{feed.site_url}</span>
                <ExternalLink className="size-3 shrink-0" />
              </a>
            )}
            {feed.parsing_error && (
              <p className="mt-1 text-sm text-destructive">
                Last parse error: {feed.parsing_error}
              </p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={refreshing}
          >
            {refreshing ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <RefreshCw className="size-3.5" />
            )}
            Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleMarkAllRead}
            disabled={marking}
          >
            {marking ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <CheckCheck className="size-3.5" />
            )}
            Mark all as read
          </Button>
          <ConfirmDialog
            trigger={
              <Button variant="destructive" size="sm" disabled={deleting}>
                {deleting ? (
                  <Loader2 className="size-3.5 animate-spin" />
                ) : (
                  <Trash2 className="size-3.5" />
                )}
                Unsubscribe
              </Button>
            }
            title="Unsubscribe from feed?"
            description={`This removes "${feed.title}" and all its entries. This cannot be undone.`}
            confirmLabel="Unsubscribe"
            onConfirm={handleDelete}
          />
          <Link
            href="/"
            className={cn(buttonVariants({ variant: "ghost", size: "sm" }), "ml-auto")}
          >
            Back
          </Link>
        </div>
      </div>
      <EntryTimeline
        filter={{ feed_id: feed.id }}
        emptyTitle="No articles yet"
        emptyDescription="This feed hasn't produced any entries. It may not have been refreshed yet."
      />
    </div>
  )
}
