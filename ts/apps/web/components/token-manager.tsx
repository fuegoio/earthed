"use client";

import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Dialog } from "@base-ui/react/dialog";
import { Key, Plus, Trash2, Loader2, Copy, Check } from "lucide-react";
import { Skeleton } from "@workspace/ui/components/skeleton";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";
import { ConfirmDialog } from "@/components/confirm-dialog";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty";
import { getClient, listTokens, createToken, deleteToken, unwrap } from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";
import { formatDateTime } from "@/lib/format";
import { cn } from "@workspace/ui/lib/utils";
import type { APIToken, CreatedToken } from "@/lib/types";

/** API token management: list, create (revealing the secret once), delete. */
export function TokenManager() {
  const queryClient = useQueryClient();
  const { data: tokens, isLoading } = useQuery<APIToken[]>({
    queryKey: ["tokens"],
    queryFn: async () => unwrap(listTokens({ client: await getClient() })),
  });

  const [createOpen, setCreateOpen] = useState(false);
  const [label, setLabel] = useState("");
  const [pending, setPending] = useState(false);
  const [created, setCreated] = useState<CreatedToken | null>(null);
  const [copied, setCopied] = useState(false);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!label.trim()) return;
    setPending(true);
    try {
      const { data, error } = await createToken({
        client: await getClient(),
        body: { label: label.trim() },
      });
      if (error) throw error;
      setCreated(data as CreatedToken);
      setLabel("");
      await queryClient.invalidateQueries({ queryKey: ["tokens"] });
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not create token"));
    } finally {
      setPending(false);
    }
  }

  function closeCreate() {
    setCreateOpen(false);
    setCreated(null);
    setCopied(false);
  }

  async function copyToken(token: string) {
    try {
      await navigator.clipboard.writeText(token);
      setCopied(true);
      toast.success("Token copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Could not copy token");
    }
  }

  async function handleDelete(id: number, label: string) {
    const { error } = await deleteToken({
      client: await getClient(),
      path: { tokenId: id },
    });
    if (error) throw error;
    await queryClient.invalidateQueries({ queryKey: ["tokens"] });
    toast.success(`Deleted token "${label}"`);
  }

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-6 sm:px-6">
      <div className="flex items-center justify-between">
        <h1 className="flex items-center gap-2 font-serif text-2xl font-bold tracking-normal">
          <Key className="size-5" />
          API tokens
        </h1>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="size-4" />
          New token
        </Button>
      </div>
      <p className="mt-1 text-sm text-muted-foreground">
        Tokens let the CLI and other clients authenticate to your account. Treat them like passwords
        — they&apos;re only shown once.
      </p>

      <div className="mt-6 flex flex-col gap-3">
        {isLoading ? (
          <div className="flex flex-col gap-3">
            {Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 rounded-lg border border-border p-4">
                <div className="min-w-0 flex-1 space-y-2">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-3 w-48" />
                </div>
                <Skeleton className="size-8 rounded-md" />
              </div>
            ))}
          </div>
        ) : (tokens ?? []).length === 0 ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Key className="size-6 text-primary" />
              </EmptyMedia>
              <EmptyTitle>No API tokens</EmptyTitle>
              <EmptyDescription>
                Create a token to use the Planetary CLI or integrate other clients.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          (tokens ?? []).map((token) => (
            <div
              key={token.id}
              className="flex items-center gap-3 rounded-lg border border-border p-4"
            >
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{token.label}</p>
                <p className="text-xs text-muted-foreground">
                  Created {formatDateTime(token.created_at)}
                  {token.last_used_at
                    ? ` · Last used ${formatDateTime(token.last_used_at)}`
                    : " · Never used"}
                </p>
              </div>
              <ConfirmDialog
                trigger={
                  <Button variant="ghost" size="icon-sm" className="text-muted-foreground" aria-label="Delete token">
                    <Trash2 className="size-3.5" />
                  </Button>
                }
                title="Delete token?"
                description={`"${token.label}" will stop working immediately. This cannot be undone.`}
                confirmLabel="Delete"
                onConfirm={() => handleDelete(token.id, token.label)}
              />
            </div>
          ))
        )}
      </div>

      {/* Create / reveal dialog */}
      <Dialog.Root
        open={createOpen}
        onOpenChange={(o) => {
          if (!o) closeCreate();
          setCreateOpen(o);
        }}
      >
        <Dialog.Portal>
          <Dialog.Backdrop className="fixed inset-0 z-50 bg-black/50" />
          <Dialog.Popup
            className={cn(
              "fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2",
              "rounded-lg border border-border bg-popover p-6 shadow-lg",
            )}
          >
            {!created ? (
              <>
                <Dialog.Title className="font-serif text-lg font-bold tracking-normal">
                  New API token
                </Dialog.Title>
                <Dialog.Description className="mt-1 text-sm text-muted-foreground">
                  Give it a label so you remember where it&apos;s used.
                </Dialog.Description>
                <form onSubmit={handleCreate} className="mt-4 flex flex-col gap-3">
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="token-label">Label</Label>
                    <Input
                      id="token-label"
                      value={label}
                      onChange={(e) => setLabel(e.target.value)}
                      placeholder="e.g. planetary-cli"
                      autoFocus
                    />
                  </div>
                  <div className="flex justify-end gap-2 pt-2">
                    <Button variant="ghost" type="button" onClick={closeCreate}>
                      Cancel
                    </Button>
                    <Button type="submit" disabled={pending || !label.trim()}>
                      {pending && <Loader2 className="size-4 animate-spin" />}
                      Create token
                    </Button>
                  </div>
                </form>
              </>
            ) : (
              <>
                <Dialog.Title className="font-serif text-lg font-bold tracking-normal">
                  Token created
                </Dialog.Title>
                <Dialog.Description className="mt-1 text-sm text-muted-foreground">
                  Copy this token now — you won&apos;t be able to see it again.
                </Dialog.Description>
                <div className="mt-4 flex items-center gap-2">
                  <code className="min-w-0 flex-1 truncate rounded-md border border-border bg-muted px-3 py-2 text-sm">
                    {created.token}
                  </code>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => copyToken(created.token)}
                    aria-label="Copy token"
                  >
                    {copied ? (
                      <Check className="size-4 text-green-600" />
                    ) : (
                      <Copy className="size-4" />
                    )}
                  </Button>
                </div>
                <div className="mt-5 flex justify-end">
                  <Button onClick={closeCreate}>Done</Button>
                </div>
              </>
            )}
          </Dialog.Popup>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  );
}
