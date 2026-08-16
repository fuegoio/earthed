"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Dialog } from "@base-ui/react/dialog";
import { Loader2, FolderPlus } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";
import { getClient, createFolder } from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";
import { cn } from "@workspace/ui/lib/utils";

/** Dialog for creating a new folder. Invalidates ["folders"] on success. */
export function FolderCreateDialog() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [pending, setPending] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    setPending(true);
    try {
      const { error } = await createFolder({
        client: await getClient(),
        body: { title: title.trim() },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["folders"] });
      setTitle("");
      setOpen(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not create folder"));
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger
        render={
          <Button
            variant="ghost"
            size="icon-xs"
            aria-label="New folder"
            className="text-muted-foreground hover:text-foreground"
          >
            <FolderPlus className="size-3.5" />
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
          <Dialog.Title className="font-serif text-lg font-bold tracking-tight">
            New folder
          </Dialog.Title>
          <Dialog.Description className="mt-1 text-sm text-muted-foreground">
            Group related feeds together.
          </Dialog.Description>
          <form onSubmit={handleSubmit} className="mt-4 flex flex-col gap-3">
            <div className="flex flex-col gap-2">
              <Label htmlFor="folder-title">Title</Label>
              <Input
                id="folder-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g. Tech, News, Design"
                autoFocus
              />
            </div>
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
                Create
              </Button>
            </div>
          </form>
        </Dialog.Popup>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
