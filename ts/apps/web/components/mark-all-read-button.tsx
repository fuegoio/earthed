"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { CheckCheck } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { getClient, updateEntries } from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";

export function MarkAllReadButton() {
  const queryClient = useQueryClient();
  const [pending, setPending] = useState(false);

  async function handleClick() {
    setPending(true);
    const { error } = await updateEntries({
      client: await getClient(),
      body: { entry_ids: null, status: "read" },
    });
    setPending(false);
    if (error) {
      toast.error(getApiErrorMessage(error, "Could not mark all as read"));
      return;
    }
    await queryClient.invalidateQueries({ queryKey: ["entries"] });
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      aria-label="Mark all as read"
      disabled={pending}
      onClick={handleClick}
    >
      <CheckCheck className="size-4" />
    </Button>
  );
}
