"use client"

import { useState } from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import {
  LayoutList,
  Circle,
  Star,
  Folder,
  Loader2,
  Plus,
  Trash2,
  ListChecks,
  Key,
} from "lucide-react"
import {
  getClient,
  listFeeds,
  listCategories,
  deleteCategory,
  unwrap,
} from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import { Logo } from "@/components/logo"
import { FeedIcon } from "@/components/feed-icon"
import { buttonVariants } from "@workspace/ui/components/button"
import { ConfirmDialog } from "@/components/confirm-dialog"
import { CategoryCreateDialog } from "@/components/category-create-dialog"
import { cn } from "@workspace/ui/lib/utils"
import { SidebarSeparator } from "@workspace/ui/components/separator"
import type { Feed, Category } from "@/lib/types"

/** Shared link class for sidebar nav rows. */
const navLinkClass = cn(
  "flex items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium",
  "text-sidebar-foreground/70 transition-colors",
  "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
)

/** Active variant — applied when the link matches the current route. */
const navLinkActiveClass = cn(
  "bg-sidebar-accent text-sidebar-accent-foreground"
)

function isActive(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/"
  return pathname === href || pathname.startsWith(href + "/")
}

function SidebarNav() {
  const pathname = usePathname()
  const navItems = [
    { href: "/", label: "All", icon: LayoutList },
    { href: "/unread", label: "Unread", icon: Circle },
    { href: "/starred", label: "Starred", icon: Star },
    { href: "/lists", label: "Feed lists", icon: ListChecks },
  ]

  return (
    <nav className="flex flex-col gap-0.5">
      {navItems.map((item) => {
        const active = isActive(pathname, item.href)
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
        )
      })}
    </nav>
  )
}

function FeedList({ feeds }: { feeds: Feed[] }) {
  const pathname = usePathname()
  if (feeds.length === 0) {
    return (
      <p className="px-3 py-2 text-xs text-muted-foreground">
        No feeds yet. Add one to get started.
      </p>
    )
  }

  return (
    <ul className="flex flex-col gap-0.5">
      {feeds.map((feed) => {
        const href = `/feeds/${feed.id}`
        const active = isActive(pathname, href)
        return (
          <li key={feed.id}>
            <Link
              href={href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "flex items-center gap-2.5 rounded-md px-3 py-1.5 text-sm",
                "text-sidebar-foreground/70 transition-colors truncate",
                active
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
              )}
            >
              <FeedIcon
                siteUrl={feed.site_url}
                className="size-3.5 shrink-0 rounded-sm"
              />
              <span className="truncate">{feed.title}</span>
            </Link>
          </li>
        )
      })}
    </ul>
  )
}

function CategoryRow({ category }: { category: Category }) {
  const pathname = usePathname()
  const queryClient = useQueryClient()
  const [pending, setPending] = useState(false)

  async function handleDelete() {
    setPending(true)
    try {
      const { error } = await deleteCategory({
        client: await getClient(),
        path: { categoryId: category.id },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["categories"] })
      await queryClient.invalidateQueries({ queryKey: ["feeds"] })
      toast.success(`Deleted category "${category.title}"`)
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not delete category"))
    } finally {
      setPending(false)
    }
  }

  const href = `/categories/${category.id}`
  const active = isActive(pathname, href)

  return (
    <li className="group flex items-center">
      <Link
        href={href}
        aria-current={active ? "page" : undefined}
        className={cn(
          "flex min-w-0 flex-1 items-center gap-2.5 rounded-md px-3 py-1.5 text-sm",
          "text-sidebar-foreground/70 transition-colors",
          active
            ? "bg-sidebar-accent text-sidebar-accent-foreground"
            : "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
        )}
      >
        <Folder
          className={cn(
            "size-3.5 shrink-0",
            active ? "text-primary" : "text-muted-foreground"
          )}
        />
        <span className="truncate">{category.title}</span>
      </Link>
      <ConfirmDialog
        trigger={
          <button
            type="button"
            aria-label={`Delete category ${category.title}`}
            disabled={pending}
            className={cn(
              "mr-1 hidden size-6 shrink-0 items-center justify-center rounded-sm text-muted-foreground",
              "hover:bg-sidebar-accent hover:text-destructive",
              "group-hover:flex"
            )}
          >
            <Trash2 className="size-3.5" />
          </button>
        }
        title="Delete category?"
        description={`"${category.title}" will be removed. Feeds in it are kept but unassigned.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
      />
    </li>
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
        <CategoryRow key={cat.id} category={cat} />
      ))}
    </ul>
  )
}

function SidebarContent() {
  const pathname = usePathname()
  const { data: feeds, isLoading: feedsLoading } = useQuery<Feed[]>({
    queryKey: ["feeds"],
    queryFn: async () => unwrap(listFeeds({ client: await getClient() })),
  })

  const { data: categories, isLoading: categoriesLoading } = useQuery<
    Category[]
  >({
    queryKey: ["categories"],
    queryFn: async () =>
      unwrap(listCategories({ client: await getClient() })),
  })

  const tokensActive = isActive(pathname, "/settings/tokens")

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
              Loading feeds…
            </div>
          ) : (
            <FeedList feeds={feeds ?? []} />
          )}
        </div>

        <SidebarSeparator />

        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between px-3 pb-1">
            <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Categories
            </h3>
            <CategoryCreateDialog />
          </div>
          {categoriesLoading ? (
            <div className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground">
              <Loader2 className="size-3.5 animate-spin" />
              Loading categories…
            </div>
          ) : (
            <CategoryList categories={categories ?? []} />
          )}
        </div>

        <SidebarSeparator />

        <Link
          href="/settings/tokens"
          aria-current={tokensActive ? "page" : undefined}
          className={cn(navLinkClass, tokensActive && navLinkActiveClass)}
        >
          <Key className={cn("size-4", tokensActive && "text-primary")} />
          API tokens
        </Link>
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
            className="absolute inset-0 bg-black/50 animate-in fade-in duration-200"
            onClick={onClose}
            aria-hidden="true"
          />
          <aside className="absolute left-0 top-0 flex h-full w-64 flex-col border-r border-sidebar-border bg-sidebar animate-in slide-in-from-left duration-200">
            <SidebarContent />
          </aside>
        </div>
      )}
    </>
  )
}
