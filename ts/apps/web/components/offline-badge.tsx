"use client";

import { useSyncExternalStore } from "react";
import { WifiOff } from "lucide-react";
import { cn } from "@workspace/ui/lib/utils";

function subscribe(callback: () => void) {
  window.addEventListener("online", callback);
  window.addEventListener("offline", callback);
  return () => {
    window.removeEventListener("online", callback);
    window.removeEventListener("offline", callback);
  };
}

function getSnapshot() {
  return navigator.onLine;
}

function getServerSnapshot() {
  return true;
}

/**
 * Shows a wifi-off icon when the browser has no network connection.
 */
export function OfflineBadge({ className }: { className?: string }) {
  const isOnline = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);

  if (isOnline) return null;

  return (
    <div className={cn("bg-muted rounded-full p-1", className)}>
      <WifiOff className="size-3 text-muted-foreground" />
    </div>
  );
}
