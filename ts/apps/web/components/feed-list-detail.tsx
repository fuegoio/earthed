"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Dialog } from "@base-ui/react/dialog"
import {
  ArrowLeft,
  Globe,
  Lock,
  Plus,
  Trash2,
  Loader2,
  Pencil,
  Share2,
  UserPlus,
  UserCheck,
  Download,
  Check,
  Rss,
} from "lucide-react"
import Link from "next/link"
import { Button } from "@workspace/ui/components/button"
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"
import { ConfirmDialog } from "@/components/confirm-dialog"
import { Empty, EmptyDescription, EmptyTitle } from "@/components/empty"
import { FeedIcon } from "@/components/feed-icon"
import {
  getClient,
  getFeedList,
  addFeedListFeed,
  removeFeedListFeed,
  deleteFeedList,
  updateFeedList,
  followFeedList,
  unfollowFeedList,
  importFeedList,
  unwrap,
} from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { cn } from "@workspace/ui/lib/utils"
import type { FeedList, FeedListFeed } from "@/lib/types"

/**
 * Feed list detail: shows the list metadata and its feeds. Owners can edit,
 * delete, and add/remove feeds. Anyone who can see the list can follow it and
 * import all its feeds into their subscriptions.
 */
export function FeedListDetail({
  list: initial,
  currentUserId,
}: {
  list: FeedList
  currentUserId: number
}) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const isOwner = initial.user_id === currentUserId

  const { data: list } = useQuery<FeedList>({
    queryKey: ["feed-list", initial.id],
    queryFn: async () => unwrap(getFeedList({ client: await getClient(), path: { listId: initial.id } })),
    initialData: initial,
  })

  const feeds = list?.feeds ?? []
  const isFollowing = !!list?.is_following

  const [addUrl, setAddUrl] = useState("")
  const [addTitle, setAddTitle] = useState("")
  const [adding, setAdding] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [shareCopied, setShareCopied] = useState(false)
  const [followPending, setFollowPending] = useState(false)
  const [importing, setImporting] = useState(false)

  async function handleAddFeed(e: React.FormEvent) {
    e.preventDefault()
    if (!addUrl.trim()) return
    setAdding(true)
    try {
      const { error } = await addFeedListFeed({
        client: await getClient(),
        path: { listId: initial.id },
        body: { feed_url: addUrl.trim(), title: addTitle.trim() || undefined },
      })
      if (error) throw error
      setAddUrl("")
      setAddTitle("")
      await queryClient.invalidateQueries({ queryKey: ["feed-list", initial.id] })
      toast.success("Added feed to list")
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not add feed"))
    } finally {
      setAdding(false)
    }
  }

  async function handleRemoveFeed(item: FeedListFeed) {
    const { error } = await removeFeedListFeed({
      client: await getClient(),
      path: { listId: initial.id, itemId: item.id },
    })
    if (error) throw error
    await queryClient.invalidateQueries({ queryKey: ["feed-list", initial.id] })
    toast.success("Removed feed from list")
  }

  async function handleDelete() {
    const { error } = await deleteFeedList({
      client: await getClient(),
      path: { listId: initial.id },
    })
    if (error) throw error
    await queryClient.invalidateQueries({ queryKey: ["feed-lists"] })
    toast.success("Feed list deleted")
    router.push("/lists")
    router.refresh()
  }

  async function handleFollow() {
    setFollowPending(true)
    try {
      const { error } = await followFeedList({
        client: await getClient(),
        path: { listId: initial.id },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["feed-list", initial.id] })
      await queryClient.invalidateQueries({ queryKey: ["feed-lists"] })
      toast.success(`Following "${list?.title}"`)
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not follow list"))
    } finally {
      setFollowPending(false)
    }
  }

  async function handleUnfollow() {
    setFollowPending(true)
    try {
      const { error } = await unfollowFeedList({
        client: await getClient(),
        path: { listId: initial.id },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["feed-list", initial.id] })
      await queryClient.invalidateQueries({ queryKey: ["feed-lists"] })
      toast.success(`Unfollowed "${list?.title}"`)
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not unfollow list"))
    } finally {
      setFollowPending(false)
    }
  }

  async function handleImport() {
    setImporting(true)
    try {
      const { data, error } = await importFeedList({
        client: await getClient(),
        path: { listId: initial.id },
      })
      if (error) throw error
      const r = data
      await queryClient.invalidateQueries({ queryKey: ["feeds"] })
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
      toast.success(
        `Imported ${r?.imported ?? 0} new feed${(r?.imported ?? 0) === 1 ? "" : "s"} (${r?.skipped ?? 0} already subscribed)`
      )
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not import feeds"))
    } finally {
      setImporting(false)
    }
  }

  async function handleShare() {
    const url = `${window.location.origin}/lists/${initial.id}`
    try {
      await navigator.clipboard.writeText(url)
      setShareCopied(true)
      toast.success("List link copied to clipboard")
      setTimeout(() => setShareCopied(false), 2000)
    } catch {
      toast.error("Could not copy link")
    }
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-6 sm:px-6">
      <Link
        href="/lists"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="size-3.5" />
        Back to lists
      </Link>

      {/* Header */}
      <div className="flex items-start gap-4">
        <div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-primary/10">
          <Rss className="size-6 text-primary" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="font-serif text-2xl font-bold tracking-tight">
              {list?.title}
            </h1>
            <span className="flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
              {list?.is_public ? (
                <>
                  <Globe className="size-3" />
                  Public
                </>
              ) : (
                <>
                  <Lock className="size-3" />
                  Private
                </>
              )}
            </span>
          </div>
          {list?.description && (
            <p className="mt-1 text-sm text-muted-foreground">
              {list.description}
            </p>
          )}
          <p className="mt-1 text-xs text-muted-foreground">
            {feeds.length} {feeds.length === 1 ? "feed" : "feeds"}
            {!isOwner && list?.owner_email ? ` · by ${list.owner_email}` : ""}
          </p>
        </div>
      </div>

      {/* Actions */}
      <div className="mt-5 flex flex-wrap items-center gap-2">
        {isOwner ? (
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setEditOpen(true)}
            >
              <Pencil className="size-3.5" />
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleShare}
              disabled={!list?.is_public}
            >
              {shareCopied ? (
                <Check className="size-3.5" />
              ) : (
                <Share2 className="size-3.5" />
              )}
              Share
            </Button>
            <ConfirmDialog
              trigger={
                <Button variant="destructive" size="sm">
                  <Trash2 className="size-3.5" />
                  Delete
                </Button>
              }
              title="Delete feed list?"
              description={`"${list?.title}" and its ${feeds.length} feed${feeds.length === 1 ? "" : "s"} will be removed. Subscriptions are not affected.`}
              confirmLabel="Delete"
              onConfirm={handleDelete}
            />
          </>
        ) : (
          <>
            {isFollowing ? (
              <Button
                variant="outline"
                size="sm"
                onClick={handleUnfollow}
                disabled={followPending}
              >
                <UserCheck className="size-3.5" />
                Unfollow
              </Button>
            ) : (
              <Button
                size="sm"
                onClick={handleFollow}
                disabled={followPending}
              >
                <UserPlus className="size-3.5" />
                Follow
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={handleImport}
              disabled={importing}
            >
              {importing ? (
                <Loader2 className="size-3.5 animate-spin" />
              ) : (
                <Download className="size-3.5" />
              )}
              Import all feeds
            </Button>
          </>
        )}
      </div>

      {/* Feeds */}
      <div className="mt-8">
        <h2 className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Feeds in this list
        </h2>
        {feeds.length === 0 ? (
          <Empty>
            <div className="flex size-12 items-center justify-center rounded-xl bg-primary/10">
              <Rss className="size-6 text-primary" />
            </div>
            <EmptyTitle>No feeds yet</EmptyTitle>
            <EmptyDescription>
              {isOwner
                ? "Add feed URLs below to build your list."
                : "This list doesn't have any feeds yet."}
            </EmptyDescription>
          </Empty>
        ) : (
          <ul className="flex flex-col gap-1 rounded-lg border border-border p-2">
            {feeds.map((item) => (
              <li
                key={item.id}
                className="group flex items-center gap-3 rounded-md px-2 py-2 hover:bg-muted/50"
              >
                <FeedIcon
                  siteUrl={item.site_url}
                  className="size-4 shrink-0 rounded-sm"
                />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">
                    {item.title || item.feed_url}
                  </p>
                  <p className="truncate text-xs text-muted-foreground">
                    {item.feed_url}
                  </p>
                </div>
                {isOwner && (
                  <ConfirmDialog
                    trigger={
                      <button
                        type="button"
                        aria-label="Remove feed"
                        className="hidden size-7 items-center justify-center rounded-sm text-muted-foreground hover:bg-muted hover:text-destructive group-hover:flex"
                      >
                        <Trash2 className="size-3.5" />
                      </button>
                    }
                    title="Remove feed?"
                    description={`"${item.title || item.feed_url}" will be removed from this list.`}
                    confirmLabel="Remove"
                    onConfirm={() => handleRemoveFeed(item)}
                  />
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Add feed (owner only) */}
      {isOwner && (
        <form
          onSubmit={handleAddFeed}
          className="mt-6 flex flex-col gap-3 rounded-lg border border-border p-4"
        >
          <h3 className="text-sm font-medium">Add a feed</h3>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Input
              value={addUrl}
              onChange={(e) => setAddUrl(e.target.value)}
              placeholder="https://example.com/feed.xml"
              type="url"
              className="flex-1"
            />
            <Input
              value={addTitle}
              onChange={(e) => setAddTitle(e.target.value)}
              placeholder="Title (optional)"
              className="flex-1"
            />
            <Button type="submit" disabled={adding || !addUrl.trim()}>
              {adding ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Plus className="size-4" />
              )}
              Add
            </Button>
          </div>
        </form>
      )}

      <EditListDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        list={list ?? initial}
      />
    </div>
  )
}

function EditListDialog({
  open,
  onOpenChange,
  list,
}: {
  open: boolean
  onOpenChange: (o: boolean) => void
  list: FeedList
}) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState(list.title)
  const [description, setDescription] = useState(list.description ?? "")
  const [isPublic, setIsPublic] = useState(list.is_public)
  const [pending, setPending] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!title.trim()) return
    setPending(true)
    try {
      const { error } = await updateFeedList({
        client: await getClient(),
        path: { listId: list.id },
        body: {
          title: title.trim(),
          description: description.trim(),
          is_public: isPublic,
        },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["feed-list", list.id] })
      await queryClient.invalidateQueries({ queryKey: ["feed-lists"] })
      toast.success("List updated")
      onOpenChange(false)
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not update list"))
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Backdrop className="fixed inset-0 z-50 bg-black/50" />
        <Dialog.Popup
          className={cn(
            "fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2",
            "rounded-lg border border-border bg-popover p-6 shadow-lg"
          )}
        >
          <Dialog.Title className="font-serif text-lg font-bold tracking-tight">
            Edit list
          </Dialog.Title>
          <form onSubmit={handleSubmit} className="mt-4 flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <Label htmlFor="edit-title">Title</Label>
              <Input
                id="edit-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                maxLength={255}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="edit-desc">Description</Label>
              <textarea
                id="edit-desc"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                maxLength={2000}
                rows={3}
                className={cn(
                  "w-full resize-y rounded-md border border-input bg-transparent px-3 py-2 text-sm",
                  "placeholder:text-muted-foreground",
                  "focus-visible:border-ring focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30"
                )}
              />
            </div>
            <label className="flex cursor-pointer items-start gap-3 rounded-md border border-border p-3">
              <input
                type="checkbox"
                checked={isPublic}
                onChange={(e) => setIsPublic(e.target.checked)}
                className="mt-0.5 size-4 accent-primary"
              />
              <span className="flex flex-col">
                <span className="text-sm font-medium">Public</span>
                <span className="text-sm text-muted-foreground">
                  Public lists can be discovered and followed by others.
                </span>
              </span>
            </label>
            <div className="flex justify-end gap-2 pt-2">
              <Dialog.Close
                render={
                  <Button variant="ghost" type="button">
                    Cancel
                  </Button>
                }
              />
              <Button type="submit" disabled={pending || !title.trim()}>
                {pending && <Loader2 className="size-4 animate-spin" />}
                Save
              </Button>
            </div>
          </form>
        </Dialog.Popup>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
