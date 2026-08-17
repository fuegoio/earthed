"use client";

import type { ReactNode } from "react";
import { Menu as MenuIcon } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { useShell } from "@/components/shell-context";
import { cn } from "@workspace/ui/lib/utils";

/**
 * Shared page header used across the app for a consistent title bar.
 *
 * On mobile the sidebar menu button is inlined as the leading element so
 * there is no separate top bar. On desktop the title has a small left
 * indent that aligns with the sidebar-less layout.
 *
 * Renders a single border-bottom row with an optional leading icon, a
 * truncated serif title, and optional trailing actions. An optional
 * metadata line sits beneath the title row for secondary context (site
 * URL, feed count, author/date, etc.). Pages wrap this in their
 * `mx-auto w-full max-w-3xl` container.
 */
export function PageHeader({
  title,
  icon,
  actions,
  metadata,
  className,
}: {
  title: ReactNode;
  icon?: ReactNode;
  actions?: ReactNode;
  metadata?: ReactNode;
  className?: string;
}) {
  const shell = useShell();

  return (
    <div
      className={cn("sticky top-0 z-10 border-b border-border bg-background px-4 py-3", className)}
    >
      <div className="flex items-center justify-between gap-3">
        <h1
          className={cn(
            "flex min-w-0 items-center gap-2 font-serif text-lg font-bold tracking-normal",
          )}
        >
          {/* Mobile menu button — replaces the separate top bar */}
          {shell && (
            <Button
              variant="ghost"
              size="icon"
              aria-label="Toggle sidebar"
              onClick={shell.openSidebar}
              className="-ml-2 shrink-0 lg:hidden"
            >
              <MenuIcon className="size-4" />
            </Button>
          )}
          {icon && <span className="shrink-0">{icon}</span>}
          <span className="truncate">{title}</span>
        </h1>
        {actions ? (
          <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">{actions}</div>
        ) : null}
      </div>
      {metadata ? <div className="mt-1.5 pl-[52px] text-sm text-muted-foreground lg:pl-0">{metadata}</div> : null}
    </div>
  );
}
