"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import {
  LayoutList,
  Circle,
  Star,
  Plus,
  ListChecks,
  LogOut,
  Key,
  FileDown,
  Upload,
  Sun,
  Moon,
} from "lucide-react";
import { useTheme } from "next-themes";
import { Menu } from "@base-ui/react/menu";
import { Avatar } from "@base-ui/react/avatar";
import { Skeleton } from "@workspace/ui/components/skeleton";
import { getClient, listFeeds, listFolders, unwrap } from "@/lib/planetary";
import { signout } from "@/lib/auth";
import { Logo } from "@/components/logo";
import { OfflineBadge } from "@/components/offline-badge";
import { FeedTree } from "@/components/feed-tree";
import { SearchBox } from "@/components/search-box";
import { buttonVariants } from "@workspace/ui/components/button";
import { FolderCreateDialog } from "@/components/folder-create-dialog";
import { cn } from "@workspace/ui/lib/utils";
import { SidebarSeparator } from "@workspace/ui/components/separator";
import type { Feed, Folder } from "@/lib/types";

const navLinkClass = cn(
  "flex items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium",
  "text-sidebar-foreground/70 transition-colors",
  "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
);

const navLinkActiveClass = cn("bg-sidebar-accent text-sidebar-accent-foreground");

function isActive(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname === href || pathname.startsWith(href + "/");
}

function SidebarNav() {
  const pathname = usePathname();
  const navItems = [
    { href: "/", label: "All", icon: LayoutList },
    { href: "/unread", label: "Unread", icon: Circle },
    { href: "/starred", label: "Starred", icon: Star },
    { href: "/lists", label: "Feed lists", icon: ListChecks },
  ];

  return (
    <nav className="flex flex-col gap-0.5">
      {navItems.map((item) => {
        const active = isActive(pathname, item.href);
        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={active ? "page" : undefined}
            className={cn(navLinkClass, active && navLinkActiveClass)}
          >
            <item.icon className={cn("size-4", active && "text-primary")} />
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}

function AccountButton({ userEmail }: { userEmail: string }) {
  const router = useRouter();
  const { resolvedTheme, setTheme } = useTheme();

  async function handleSignout() {
    await signout();
    router.push("/login");
    router.refresh();
  }

  const isDark = resolvedTheme === "dark";

  const menuItemClass = cn(
    "flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm",
    "hover:bg-accent hover:text-accent-foreground transition-colors",
    "focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30",
  );

  return (
    <Menu.Root>
      <Menu.Trigger
        className={cn(
          "flex min-w-0 flex-1 items-center gap-2 rounded-md py-1.5 pl-1.5 pr-2",
          "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors",
          "focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30",
        )}
        aria-label="Account menu"
      >
        <Avatar.Root className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-medium text-primary-foreground">
          <Avatar.Fallback>{userEmail.charAt(0).toUpperCase()}</Avatar.Fallback>
        </Avatar.Root>
        <span className="hidden truncate text-sm font-medium sm:inline">{userEmail}</span>
      </Menu.Trigger>
      <Menu.Portal>
        <Menu.Positioner
          className={cn(
            "z-50 min-w-48 overflow-hidden rounded-md border border-border bg-popover p-1",
            "shadow-md",
          )}
          align="start"
          side="top"
        >
          <Menu.Popup>
            <div className="px-2 py-1.5">
              <p className="text-sm font-medium truncate">{userEmail}</p>
            </div>
            <div className="my-1 h-px bg-border" />
            <Menu.Item className={menuItemClass} render={<Link href="/settings/tokens" />}>
              <Key className="size-4" />
              API tokens
            </Menu.Item>
            <div className="my-1 h-px bg-border" />
            <Menu.Item className={menuItemClass} render={<Link href="/settings/opml" />}>
              <Upload className="size-4" />
              Import OPML
            </Menu.Item>
            <Menu.Item className={menuItemClass} render={<Link href="/settings/opml" />}>
              <FileDown className="size-4" />
              Export OPML
            </Menu.Item>
            <div className="my-1 h-px bg-border" />
            <Menu.Item
              className={menuItemClass}
              onClick={() => setTheme(isDark ? "light" : "dark")}
            >
              {isDark ? (
                <>
                  <Sun className="size-4" />
                  Light mode
                </>
              ) : (
                <>
                  <Moon className="size-4" />
                  Dark mode
                </>
              )}
            </Menu.Item>
            <div className="my-1 h-px bg-border" />
            <Menu.Item className={menuItemClass} onClick={handleSignout}>
              <LogOut className="size-4" />
              Sign out
            </Menu.Item>
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  );
}

function SidebarContent({ userEmail }: { userEmail: string }) {
  const { data: feeds, isLoading: feedsLoading } = useQuery<Feed[]>({
    queryKey: ["feeds"],
    queryFn: async () => unwrap(listFeeds({ client: await getClient() })),
  });

  const { data: folders, isLoading: foldersLoading } = useQuery<Folder[]>({
    queryKey: ["folders"],
    queryFn: async () => unwrap(listFolders({ client: await getClient() })),
  });

  const isLoading = feedsLoading || foldersLoading;

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-14 shrink-0 items-center gap-2 px-4 w-full">
        <Link href="/" className="flex items-center gap-2 font-serif text-lg font-bold px-1">
          <Logo className="size-5" />
          Planetary
        </Link>
        <div className="flex-1" />
        <OfflineBadge />
      </div>

      <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-3">
        <SearchBox className="max-w-none" />

        <SidebarNav />

        <SidebarSeparator />

        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between px-3 pb-1">
            <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Feeds
            </h3>
            <div className="flex items-center gap-0.5">
              <FolderCreateDialog />
              <Link
                href="/feeds/new"
                aria-label="Subscribe to a feed"
                className={cn(
                  buttonVariants({ variant: "ghost", size: "icon-xs" }),
                  "text-muted-foreground hover:text-foreground",
                )}
              >
                <Plus className="size-3.5" />
              </Link>
            </div>
          </div>
          {isLoading ? (
            <div className="flex flex-col gap-0.5">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-center gap-2.5 px-3 py-2">
                  <Skeleton className="size-3.5 shrink-0 rounded-sm" />
                  <Skeleton className="h-3 flex-1" />
                </div>
              ))}
            </div>
          ) : (feeds ?? []).length === 0 && (folders ?? []).length === 0 ? (
            <p className="px-3 py-2 text-xs text-muted-foreground">
              No feeds yet. Add one to get started.
            </p>
          ) : (
            <FeedTree feeds={feeds ?? []} folders={folders ?? []} />
          )}
        </div>
      </div>

      <div className="shrink-0 p-3">
        <AccountButton userEmail={userEmail} />
      </div>
    </div>
  );
}

export function AppSidebar({
  open,
  onClose,
  userEmail,
}: {
  open: boolean;
  onClose: () => void;
  userEmail: string;
}) {
  return (
    <>
      <aside className="hidden w-64 shrink-0 border-r border-sidebar-border bg-sidebar lg:flex lg:flex-col">
        <SidebarContent userEmail={userEmail} />
      </aside>

      {open && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div
            className="absolute inset-0 bg-black/50 animate-in fade-in duration-200"
            onClick={onClose}
            aria-hidden="true"
          />
          <aside
            className="absolute left-0 top-0 flex h-full w-64 flex-col border-r border-sidebar-border bg-sidebar animate-in slide-in-from-left duration-200"
            onClick={(e) => {
              if ((e.target as HTMLElement).closest("a")) onClose();
            }}
          >
            <SidebarContent userEmail={userEmail} />
          </aside>
        </div>
      )}
    </>
  );
}
