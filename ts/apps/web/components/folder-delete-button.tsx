"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Trash2, Loader2 } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { getClient, deleteFolder } from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";
import type { Folder } from "@/lib/types";

/**
 * Delete button for the folder page header. Renders a destructive button
 * that opens a confirmation dialog before deleting the folder. After
 * deletion it invalidates the feeds/folders queries and redirects home.
 */
export function FolderDeleteButton({ folder }: { folder: Folder }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [pending, setPending] = useState(false);

  async function handleDelete() {
    setPending(true);
    try {
      const { error } = await deleteFolder({
        client: await getClient(),
        path: { folderId: folder.id },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["folders"] });
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      await queryClient.invalidateQueries({ queryKey: ["entries"] });
      toast.success(`Deleted folder "${folder.title}"`);
      router.push("/");
      router.refresh();
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not delete folder"));
    } finally {
      setPending(false);
    }
  }

  return (
    <ConfirmDialog
      trigger={
        <Button variant="ghost" size="sm" disabled={pending} className="text-muted-foreground hover:text-destructive hover:bg-destructive/10">
          {pending ? (
            <Loader2 className="size-3.5 animate-spin" />
          ) : (
            <Trash2 className="size-3.5" />
          )}
          Delete
        </Button>
      }
      title="Delete folder?"
      description={`"${folder.title}" will be removed. Feeds in it are kept but unassigned.`}
      confirmLabel="Delete"
      onConfirm={handleDelete}
    />
  );
}
