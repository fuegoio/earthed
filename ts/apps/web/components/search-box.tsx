"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Search } from "lucide-react";
import { cn } from "@workspace/ui/lib/utils";

/** Topbar search box. Submits to /search?q=… using the current query as default. */
export function SearchBox({ className }: { className?: string }) {
  const router = useRouter();
  const params = useSearchParams();
  const [q, setQ] = useState(params.get("q") ?? "");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = q.trim();
    if (!trimmed) return;
    router.push(`/search?q=${encodeURIComponent(trimmed)}`);
  }

  return (
    <form onSubmit={handleSubmit} className={cn("relative w-full max-w-xs", className)}>
      <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <input
        type="search"
        name="q"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder="Search entries…"
        aria-label="Search entries"
        className={cn(
          "h-9 w-full rounded-md border border-input bg-transparent pl-9 pr-3 text-sm",
          "placeholder:text-muted-foreground",
          "focus-visible:border-ring focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30",
          "transition-[color,box-shadow]",
        )}
      />
    </form>
  );
}
