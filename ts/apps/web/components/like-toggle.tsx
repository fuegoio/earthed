"use client"

import { useState, useOptimistic, startTransition } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Heart } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { getClient, toggleEntryLiked } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { cn } from "@workspace/ui/lib/utils"

/**
 * Toggles the liked flag on an entry. Optimistically flips the heart
 * immediately and rolls back on error. Invalidates entry queries so
 * counts stay fresh after the mutation settles.
 */
export function LikeToggle({
  entryId,
  liked: likedProp,
  size = "icon-sm",
  className,
}: {
  entryId: number
  liked: boolean
  size?: "icon-xs" | "icon-sm" | "icon"
  className?: string
}) {
  const queryClient = useQueryClient()
  const [liked, setOptimistic] = useOptimistic(likedProp)
  const [pending, setPending] = useState(false)

  async function handleToggle() {
    const next = !liked
    startTransition(() => {
      setOptimistic(next)
      setPending(true)
    })
    try {
      const { error } = await toggleEntryLiked({
        client: await getClient(),
        path: { entryId },
        body: { liked: next },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["entries"] })
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not update entry"))
    } finally {
      setPending(false)
    }
  }

  return (
    <Button
      variant="ghost"
      size={size}
      aria-label={liked ? "Remove like" : "Like entry"}
      aria-pressed={liked}
      disabled={pending}
      onClick={(e) => {
        e.preventDefault()
        e.stopPropagation()
        handleToggle()
      }}
      className={cn(className)}
    >
      <Heart
        className={cn(
          "transition-colors",
          liked
            ? "fill-red-500 text-red-500"
            : "text-muted-foreground hover:text-foreground"
        )}
      />
    </Button>
  )
}
