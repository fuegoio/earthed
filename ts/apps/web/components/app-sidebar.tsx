"use client"

import Link from "next/link"
import { useQuery } from "@tanstack/react-query"
import {
  Rss,
  LayoutList,
  Circle,
  Star,
  Folder,
  Loader2,
  Plus,
} from "lucide-react"
import { getClient, listFeeds, listCategories, unwrap } from "@/lib/planetary"
import { Logo } from "@/components/logo"
import { buttonVariants } from "@workspace/ui/components/button"
import { cn } from "@workspace/ui/lib/utils"
import { SidebarSeparator } from "@workspace/ui/components/separator"
import type { Feed, Category } from "@/lib/types"

function SidebarNav() {
  const navItems = [
    { href: "/", label: "All", icon: LayoutList },
    { href: "/unread", label: "Unread", icon: Circle },
    { href: "/starred", label: "Starred", icon: Star },
  ]

  return (
    <nav className="flex flex-col gap-0.5">
      {navItems.map((item) => (
        <Link
          key={item.href}
          href={item.href}
          className={cn(
            "flex items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium",
            "text-sidebar-foreground/80 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
            "transition-colors"
          )}
        >
          <item.icon className="size-4" />
          {item.label}
        </Link>
      ))}
    </nav>
  )
}

function faviconUrl(siteUrl: string): string {
  return `https://www.google.com/s2/favicons?domain=${siteUrl}&sz=64`
}

function FeedIcon({ siteUrl }: { siteUrl: string }) {
  if (siteUrl) {
    return (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={faviconUrl(siteUrl)}
        alt=""
        className="size-3.5 shrink-0 rounded-sm"
        width={14}
        height={14}
      />
    )
  }
  return <Rss className="size-3.5 shrink-0 text-muted-foreground" />
}

function FeedList({ feeds }: { feeds: Feed[] }) {
  if (feeds.length === 0) {
    return (
      <p className="px-3 py-2 text-xs text-muted-foreground">
        No feeds yet. Add one to get started.
      </p>
    )
  }

  return (
    <ul className="flex flex-col gap-0.5">
      {feeds.map((feed) => (
        <li key={feed.id}>
          <Link
            href={`/feeds/${feed.id}`}
            className={cn(
              "flex items-center gap-2.5 rounded-md px-3 py-1.5 text-sm",
              "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
              "transition-colors truncate"
            )}
          >
            <FeedIcon siteUrl={feed.site_url ?? ""} />
            <span className="truncate">{feed.title}</span>
          </Link>
        </li>
      ))}
    </ul>
  )
}

function CategoryList({ categories }: { categories: Category[] }) {
  if (categories.length === 0) {
    return (
      <p className="px-3 py-2 text-xs text-muted-foreground">
        No categories yet.
      </p>
    )
  }

  return (
    <ul className="flex flex-col gap-0.5">
      {categories.map((cat) => (
        <li key={cat.id}>
          <Link
            href={`/categories/${cat.id}`}
            className={cn(
              "flex items-center gap-2.5 rounded-md px-3 py-1.5 text-sm",
              "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
              "transition-colors truncate"
            )}
          >
            <Folder className="size-3.5 shrink-0 text-muted-foreground" />
            <span className="truncate">{cat.title}</span>
          </Link>
        </li>
      ))}
    </ul>
  )
}

function SidebarContent() {
  const { data: feeds, isLoading: feedsLoading } = useQuery<Feed[]>({
    queryKey: ["feeds"],
    queryFn: async () => unwrap(listFeeds({ client: await getClient() })),
  })

  const { data: categories, isLoading: categoriesLoading } = useQuery<
    Category[]
  >({
    queryKey: ["categories"],
    queryFn: async () => unwrap(listCategories({ client: await getClient() })),
  })

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-14 shrink-0 items-center gap-2 border-b border-sidebar-border px-4">
        <Link
          href="/"
          className="flex items-center gap-2 font-serif text-lg font-bold"
        >
          <Logo className="size-5" />
          Planetary
        </Link>
      </div>

      <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-3">
        <SidebarNav />

        <SidebarSeparator />

        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between px-3 pb-1">
            <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Feeds
            </h3>
            <Link
              href="/feeds/new"
              aria-label="Subscribe to a feed"
              className={cn(
                buttonVariants({ variant: "ghost", size: "icon-xs" }),
                "text-muted-foreground hover:text-foreground"
              )}
            >
              <Plus className="size-3.5" />
            </Link>
          </div>
          {feedsLoading ? (
            <div className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground">
              <Loader2 className="size-3.5 animate-spin" />
              Loading feeds...
            </div>
          ) : (
            <FeedList feeds={feeds ?? []} />
          )}
        </div>

        <SidebarSeparator />

        <div className="flex flex-col gap-1">
          <h3 className="px-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Categories
          </h3>
          {categoriesLoading ? (
            <div className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground">
              <Loader2 className="size-3.5 animate-spin" />
              Loading categories...
            </div>
          ) : (
            <CategoryList categories={categories ?? []} />
          )}
        </div>
      </div>
    </div>
  )
}

export function AppSidebar({
  open,
  onClose,
}: {
  open: boolean
  onClose: () => void
}) {
  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden w-64 shrink-0 border-r border-sidebar-border bg-sidebar lg:flex lg:flex-col">
        <SidebarContent />
      </aside>

      {/* Mobile sidebar drawer */}
      {open && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={onClose}
            aria-hidden="true"
          />
          <aside className="absolute left-0 top-0 flex h-full w-64 flex-col border-r border-sidebar-border bg-sidebar">
            <SidebarContent />
          </aside>
        </div>
      )}
    </>
  )
}
