"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Share2 } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { getClient, shareArticle, unshareArticle } from "@/lib/sunred";
import { getApiErrorMessage } from "@/lib/errors";
import { cn } from "@workspace/ui/lib/utils";
import type { Entry, Feed } from "@/lib/types";

/**
 * Toggles sharing of an entry to the user's social timeline.
 * On first share, sends metadata to /api/v1/social/shares.
 * On unshare, calls DELETE /api/v1/social/shares/{shareId}.
 *
 * `shareId` is null when the article is not yet shared; after a successful
 * share it becomes the server-assigned id so subsequent clicks unshare.
 */
export function ShareToggle({
  entry,
  feed,
  shareId: initialShareId,
  size = "icon-sm",
  className,
}: {
  entry: Entry;
  feed?: Feed;
  shareId: number | null;
  size?: "icon-xs" | "icon-sm" | "icon";
  className?: string;
}) {
  const [shareId, setShareId] = useState<number | null>(initialShareId);
  const [pending, setPending] = useState(false);

  const shared = shareId !== null;

  async function handleToggle() {
    setPending(true);
    try {
      if (shared && shareId !== null) {
        const { error } = await unshareArticle({
          client: await getClient(),
          path: { shareId },
        });
        if (error) throw error;
        setShareId(null);
        toast.success("Removed from your shares");
      } else {
        const { data, error } = await shareArticle({
          client: await getClient(),
          body: {
            article_url: entry.url,
            title: entry.title,
            description: entry.description ?? "",
            feed_url: feed?.feed_url ?? entry.feed_url ?? "",
            feed_title: feed?.title ?? entry.feed_title ?? "",
            feed_site_url: feed?.site_url ?? entry.feed_site_url ?? "",
            author: entry.author ?? "",
            published_at: entry.published_at ?? undefined,
          },
        });
        if (error) throw error;
        setShareId(data?.id ?? null);
        toast.success("Shared to your timeline");
      }
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not update share"));
    } finally {
      setPending(false);
    }
  }

  return (
    <Button
      variant="ghost"
      size={size}
      aria-label={shared ? "Remove from your shares" : "Share this article"}
      aria-pressed={shared}
      disabled={pending}
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        handleToggle();
      }}
      className={cn(className)}
    >
      <Share2
        className={cn(
          "transition-[color,fill] duration-200",
          shared ? "text-primary" : "text-muted-foreground",
        )}
      />
    </Button>
  );
}
