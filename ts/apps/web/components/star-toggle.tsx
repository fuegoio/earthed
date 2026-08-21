"use client";

import { useState, useOptimistic, startTransition } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Star } from "lucide-react";
import { motion, useAnimationControls } from "motion/react";
import { Button } from "@workspace/ui/components/button";
import { getClient, toggleEntryStarred } from "@/lib/sunred";
import { getApiErrorMessage } from "@/lib/errors";
import { cn } from "@workspace/ui/lib/utils";

/**
 * Toggles the starred flag on an entry. Optimistically flips the star
 * immediately and rolls back on error. Invalidates entry queries so
 * starred/unread counts stay fresh after the mutation settles.
 *
 * Plays a scale-pop via framer-motion on every toggle for tactile feedback.
 */
export function StarToggle({
  entryId,
  starred: starredProp,
  size = "icon-sm",
  className,
}: {
  entryId: number;
  starred: boolean;
  size?: "icon-xs" | "icon-sm" | "icon";
  className?: string;
}) {
  const queryClient = useQueryClient();
  const [starred, setOptimistic] = useOptimistic(starredProp);
  const [pending, setPending] = useState(false);
  const controls = useAnimationControls();

  async function handleToggle() {
    const next = !starred;
    // Fire the pop sequence immediately — no state needed
    void controls.start({ scale: [1, 0.72, 1.22, 1], transition: { duration: 0.2, ease: "easeOut" } });
    startTransition(() => {
      setOptimistic(next);
      setPending(true);
    });
    try {
      const { error } = await toggleEntryStarred({
        client: await getClient(),
        path: { entryId },
        body: { starred: next },
      });
      if (error) throw error;
      await queryClient.invalidateQueries({ queryKey: ["entries"] });
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not update entry"));
    } finally {
      setPending(false);
    }
  }

  return (
    <Button
      variant="ghost"
      size={size}
      aria-label={starred ? "Remove from starred" : "Add to starred"}
      aria-pressed={starred}
      disabled={pending}
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        handleToggle();
      }}
      className={cn(className)}
    >
      <motion.span animate={controls} className="flex">
        <Star
          className={cn(
            "transition-[color,fill] duration-200",
            starred ? "fill-primary text-primary" : "text-muted-foreground",
          )}
        />
      </motion.span>
    </Button>
  );
}
