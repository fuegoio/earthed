"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { ExternalLink, Trash2, CheckCheck, Loader2, RefreshCw } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { ConfirmDialog } from "@/components/confirm-dialog"
import { FeedIcon } from "@/components/feed-icon"
import { PageHeader } from "@/components/page-header"
import { EntryTimeline } from "@/components/entry-timeline"
import { getClient, markFeedRead, deleteFeed, refreshFeed } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import type { Feed } from "@/lib/types"

/**
 * Feed detail view: header with site link, refresh, mark-all-read, and delete
 * actions, plus the feed's entry timeline. The feed is refreshed server-side
 * before this component renders; the refresh button here is for on-demand use.
 */
export function FeedDetail({ feed }: { feed: Feed }) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const [marking, setMarking] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [refreshing, setRefreshing] = useState(false)

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
      <PageHeader
        title={feed.title || "Untitled feed"}
        icon={
          <FeedIcon
            siteUrl={feed.site_url}
            className="size-5 shrink-0 rounded-md"
          />
        }
        actions={
          <>
            <Button
              variant="outline"
              size="xs"
              onClick={handleRefresh}
              disabled={refreshing}
            >
              {refreshing ? (
                <Loader2 className="animate-spin" />
              ) : (
                <RefreshCw />
              )}
              Refresh
            </Button>
            <Button
              variant="outline"
              size="xs"
              onClick={handleMarkAllRead}
              disabled={marking}
            >
              {marking ? (
                <Loader2 className="animate-spin" />
              ) : (
                <CheckCheck />
              )}
              Mark all as read
            </Button>
            <ConfirmDialog
              trigger={
                <Button variant="destructive" size="xs" disabled={deleting}>
                  {deleting ? (
                    <Loader2 className="animate-spin" />
                  ) : (
                    <Trash2 />
                  )}
                  Unsubscribe
                </Button>
              }
              title="Unsubscribe from feed?"
              description={`This removes "${feed.title}" and all its entries. This cannot be undone.`}
              confirmLabel="Unsubscribe"
              onConfirm={handleDelete}
            />
          </>
        }
        metadata={
          <>
            {feed.site_url && (
              <a
                href={feed.site_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 hover:text-foreground transition-colors"
              >
                <span className="truncate">{feed.site_url}</span>
                <ExternalLink className="size-3 shrink-0" />
              </a>
            )}
            {feed.parsing_error && (
              <p className="mt-1 text-destructive">
                Last parse error: {feed.parsing_error}
              </p>
            )}
          </>
        }
      />
      <EntryTimeline
        filter={{ feed_id: feed.id }}
        emptyTitle="No articles yet"
        emptyDescription="This feed hasn't produced any entries. It may not have been refreshed yet."
      />
    </div>
  )
}
