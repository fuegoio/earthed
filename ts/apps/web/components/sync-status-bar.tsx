"use client";

import { useQuery } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { getClient, getMe, unwrap } from "@/lib/sunred";
import type { User } from "@/lib/types";

/**
 * Shows a waiting banner while the post-login PDS backfill runs.
 *
 * The app layout fetches the user server-side and passes the initial
 * pds_sync_status here. When it is "syncing" the component polls /v1/me until
 * the status settles to "idle" or "failed", then hides itself. For users who
 * are not syncing, the query stays disabled so no extra request is made.
 */
export function SyncStatusBar({ initialStatus }: { initialStatus: string }) {
  const { data } = useQuery<User>({
    queryKey: ["me"],
    queryFn: async () => unwrap(getMe({ client: await getClient() })),
    // Only poll while the server-rendered status indicated a sync in flight.
    enabled: initialStatus === "syncing",
    refetchInterval: (query) =>
      query.state.data?.pds_sync_status === "syncing" ? 2000 : false,
  });

  const status = data?.pds_sync_status ?? initialStatus;
  if (status !== "syncing") return null;

  return (
    <div className="flex items-center gap-2 border-b bg-muted/40 px-4 py-2 text-sm text-muted-foreground">
      <Loader2 className="size-4 shrink-0 animate-spin" />
      <span>Importing your feeds and follows from your PDS&hellip;</span>
    </div>
  );
}
