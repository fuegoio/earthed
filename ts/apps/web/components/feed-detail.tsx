"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  ExternalLink,
  Trash2,
  CheckCheck,
  Loader2,
  RefreshCw,
  Rss,
  FolderOpen,
} from "lucide-react";
import { Menu } from "@base-ui/react/menu";
import { Button, buttonVariants } from "@workspace/ui/components/button";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { FeedIcon } from "@/components/feed-icon";
import { PageHeader } from "@/components/page-header";
import { EntryTimeline } from "@/components/entry-timeline";
import {
  getClient,
  markFeedRead,
  deleteFeed,
  refreshFeed,
  updateFeed,
  listFolders,
  unwrap,
} from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";
import { cn } from "@workspace/ui/lib/utils";
import type { Feed, Folder } from "@/lib/types";

/**
 * Feed detail view: header with site link, refresh, mark-all-read, and delete
 * actions, plus the feed's entry timeline. The feed is refreshed server-side
 * before this component renders; the refresh button here is for on-demand use.
 */
export function FeedDetail({ feed }: { feed: Feed }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [marking, setMarking] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [movingFolder, setMovingFolder] = useState(false);

  const { data: folders } = useQuery<Folder[]>({
    queryKey: ["folders"],
    queryFn: async () => unwrap(listFolders({ client: await getClient() })),
  });

  async function handleRefresh() {
    setRefreshing(true);
    try {
      const { error } = await refreshFeed({
        client: await getClient(),
        path: { feedId: feed.id },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["entries"] });
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      toast.success("Feed refreshed");
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not refresh feed"));
    } finally {
      setRefreshing(false);
    }
  }

  async function handleMarkAllRead() {
    setMarking(true);
    try {
      const { error } = await markFeedRead({
        client: await getClient(),
        path: { feedId: feed.id },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["entries"] });
      toast.success(`Marked all entries in "${feed.title}" as read`);
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not mark entries as read"));
    } finally {
      setMarking(false);
    }
  }

  async function handleDelete() {
    setDeleting(true);
    try {
      const { error } = await deleteFeed({
        client: await getClient(),
        path: { feedId: feed.id },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      await queryClient.invalidateQueries({ queryKey: ["entries"] });
      toast.success(`Unsubscribed from "${feed.title}"`);
      router.push("/");
      router.refresh();
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not delete feed"));
    } finally {
      setDeleting(false);
    }
  }

  async function handleMoveFolder(folderId: number | undefined) {
    setMovingFolder(true);
    try {
      const { error } = await updateFeed({
        client: await getClient(),
        path: { feedId: feed.id },
        body: { folder_id: folderId },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      await queryClient.invalidateQueries({ queryKey: ["entries"] });
      toast.success("Feed moved");
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not move feed"));
    } finally {
      setMovingFolder(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="sticky top-0 z-10 bg-background">
        <PageHeader
          title={feed.title || "Untitled feed"}
          icon={<FeedIcon siteUrl={feed.site_url} className="size-5 shrink-0 rounded-md" />}
          actions={
            <div className="flex items-center gap-1">
              <Menu.Root>
                <Menu.Trigger
                  disabled={movingFolder}
                  aria-label="Move to folder"
                  className={cn(buttonVariants({ variant: "ghost", size: "icon-sm" }))}
                >
                  {movingFolder ? (
                    <Loader2 className="size-3.5 animate-spin" />
                  ) : (
                    <FolderOpen className="size-3.5" />
                  )}
                </Menu.Trigger>
                <Menu.Portal>
                  <Menu.Positioner
                    className={cn(
                      "z-50 min-w-48 overflow-hidden rounded-md border border-border bg-popover p-1",
                      "shadow-md",
                    )}
                    align="end"
                  >
                    <Menu.Popup>
                      <Menu.Item
                        className="flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
                        onClick={() => handleMoveFolder(undefined)}
                      >
                        No folder
                      </Menu.Item>
                      {folders?.map((f) => (
                        <Menu.Item
                          key={f.id}
                          className="flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
                          onClick={() => handleMoveFolder(f.id)}
                        >
                          {f.title}
                        </Menu.Item>
                      ))}
                    </Menu.Popup>
                  </Menu.Positioner>
                </Menu.Portal>
              </Menu.Root>
              {feed.site_url && (
                <a
                  href={feed.site_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label="Open website"
                  className={cn(buttonVariants({ variant: "ghost", size: "icon-sm" }))}
                >
                  <ExternalLink className="size-3.5" />
                </a>
              )}
              {feed.feed_url && (
                <a
                  href={feed.feed_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label="Open feed XML"
                  className={cn(buttonVariants({ variant: "ghost", size: "icon-sm" }))}
                >
                  <Rss className="size-3.5" />
                </a>
              )}
            </div>
          }
          metadata={
            <>
              {feed.description && <p>{feed.description}</p>}
              {feed.parsing_error && (
                <p className="mt-1 text-destructive">Last parse error: {feed.parsing_error}</p>
              )}
            </>
          }
        />
        <div className="flex items-center gap-2 border-b border-border px-4 py-2">
          <Button variant="outline" size="xs" onClick={handleMarkAllRead} disabled={marking}>
            {marking ? <Loader2 className="animate-spin" /> : <CheckCheck />}
            Mark all as read
          </Button>
          <Button variant="outline" size="xs" onClick={handleRefresh} disabled={refreshing}>
            {refreshing ? <Loader2 className="animate-spin" /> : <RefreshCw />}
            Refresh
          </Button>
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon-xs" disabled={deleting} className="ml-auto text-muted-foreground" aria-label="Unsubscribe from feed">
                {deleting ? <Loader2 className="animate-spin" /> : <Trash2 />}
              </Button>
            }
            title="Unsubscribe from feed?"
            description={`This removes "${feed.title}" and all its entries. This cannot be undone.`}
            confirmLabel="Unsubscribe"
            onConfirm={handleDelete}
          />
        </div>
      </div>
      <EntryTimeline
        filter={{ feed_id: feed.id }}
        emptyTitle="No articles yet"
        emptyDescription="This feed hasn't produced any entries. It may not have been refreshed yet."
      />
    </div>
  );
}
