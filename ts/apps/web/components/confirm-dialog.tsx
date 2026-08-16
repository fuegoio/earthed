"use client";

import { useState } from "react";
import { AlertDialog } from "@base-ui/react/alert-dialog";
import { Loader2 } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { cn } from "@workspace/ui/lib/utils";

/**
 * Reusable confirmation dialog built on base-ui's AlertDialog. The resource-
 * specific mutation lives in the parent's `onConfirm` handler — this is a UI
 * primitive only, not a resource-aware wrapper.
 */
export function ConfirmDialog({
  trigger,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  destructive = true,
  onConfirm,
}: {
  trigger: React.ReactElement;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
  onConfirm: () => Promise<void> | void;
}) {
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);

  async function handleConfirm() {
    setPending(true);
    try {
      await onConfirm();
      setOpen(false);
    } finally {
      setPending(false);
    }
  }

  return (
    <AlertDialog.Root open={open} onOpenChange={setOpen}>
      <AlertDialog.Trigger render={trigger} />
      <AlertDialog.Portal>
        <AlertDialog.Backdrop
          className="fixed inset-0 z-50 bg-black/50"
          data-testid="confirm-backdrop"
        />
        <AlertDialog.Popup
          className={cn(
            "fixed left-1/2 top-1/2 z-50 w-full max-w-sm -translate-x-1/2 -translate-y-1/2",
            "rounded-lg border border-border bg-popover p-6 shadow-lg",
          )}
        >
          <AlertDialog.Title className="font-serif text-lg font-bold tracking-tight">
            {title}
          </AlertDialog.Title>
          {description && (
            <AlertDialog.Description className="mt-2 text-sm text-muted-foreground">
              {description}
            </AlertDialog.Description>
          )}
          <div className="mt-5 flex justify-end gap-2">
            <AlertDialog.Close
              render={
                <Button variant="ghost" disabled={pending}>
                  {cancelLabel}
                </Button>
              }
            />
            <Button
              variant={destructive ? "destructive" : "default"}
              onClick={handleConfirm}
              disabled={pending}
            >
              {pending && <Loader2 className="size-4 animate-spin" />}
              {confirmLabel}
            </Button>
          </div>
        </AlertDialog.Popup>
      </AlertDialog.Portal>
    </AlertDialog.Root>
  );
}
