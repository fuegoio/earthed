"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Dialog } from "@base-ui/react/dialog";
import { Loader2, Pencil } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";
import { getClient, updateFeed } from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";
import { cn } from "@workspace/ui/lib/utils";
import type { Feed } from "@/lib/types";

/**
 * Edit button + dialog for renaming a feed. Renders an icon trigger that
 * opens a small dialog pre-filled with the current title. On save, patches
 * the feed name via the API and invalidates the feeds query.
 */
export function FeedRenameDialog({ feed }: { feed: Feed }) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState(feed.title);
  const [pending, setPending] = useState(false);

  function handleOpenChange(next: boolean) {
    if (next) setTitle(feed.title);
    setOpen(next);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = title.trim();
    if (!trimmed || trimmed === feed.title) {
      setOpen(false);
      return;
    }
    setPending(true);
    try {
      const { error } = await updateFeed({
        client: await getClient(),
        path: { feedId: feed.id },
        body: { title: trimmed },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      toast.success("Feed renamed");
      setOpen(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not rename feed"));
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <Dialog.Trigger
        render={
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Rename feed"
            className="text-muted-foreground hover:text-foreground"
          >
            <Pencil className="size-3.5" />
          </Button>
        }
      />
      <Dialog.Portal>
        <Dialog.Backdrop className="fixed inset-0 z-50 bg-black/50" />
        <Dialog.Popup
          className={cn(
            "fixed left-1/2 top-1/2 z-50 w-full max-w-sm -translate-x-1/2 -translate-y-1/2",
            "rounded-lg border border-border bg-popover p-6 shadow-lg",
          )}
        >
          <Dialog.Title className="font-serif text-lg font-bold tracking-normal">
            Rename feed
          </Dialog.Title>
          <Dialog.Description className="mt-1 text-sm text-muted-foreground">
            Give this feed a custom display name.
          </Dialog.Description>
          <form onSubmit={handleSubmit} className="mt-4 flex flex-col gap-3">
            <div className="flex flex-col gap-2">
              <Label htmlFor="feed-title">Name</Label>
              <Input
                id="feed-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Feed name"
                autoFocus
                autoComplete="off"
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Dialog.Close
                render={
                  <Button variant="ghost" type="button" disabled={pending}>
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
  );
}
