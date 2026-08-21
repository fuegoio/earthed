"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Loader2, Pencil } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@workspace/ui/components/dialog";
import { getClient, updateFeed } from "@/lib/sunred";
import { getApiErrorMessage } from "@/lib/errors";
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
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger
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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Rename feed</DialogTitle>
          <DialogDescription>Give this feed a custom display name.</DialogDescription>
        </DialogHeader>
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
          <DialogFooter>
            <DialogClose
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
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
