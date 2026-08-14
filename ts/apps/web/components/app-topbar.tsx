"use client"

import Link from "next/link"
import { useRouter } from "next/navigation"
import { Menu as MenuIcon, LogOut, Key, FileDown, Upload } from "lucide-react"
import { Menu } from "@base-ui/react/menu"
import { Avatar } from "@base-ui/react/avatar"
import { Button } from "@workspace/ui/components/button"
import { ThemeToggle } from "@/components/theme-toggle"
import { SearchBox } from "@/components/search-box"
import { signout } from "@/lib/auth"
import { cn } from "@workspace/ui/lib/utils"

export function AppTopbar({
  onMenuClick,
  userEmail,
}: {
  onMenuClick: () => void
  userEmail: string
}) {
  const router = useRouter()

  async function handleSignout() {
    await signout()
    router.push("/login")
    router.refresh()
  }

  return (
    <header className="flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background px-3">
      <Button
        variant="ghost"
        size="icon"
        className="lg:hidden"
        aria-label="Toggle sidebar"
        onClick={onMenuClick}
      >
        <MenuIcon className="size-4" />
      </Button>

      <div className="flex flex-1 items-center gap-2">
        <div className="hidden max-w-xs flex-1 md:block">
          <SearchBox />
        </div>
      </div>

      <ThemeToggle />

      <Menu.Root>
        <Menu.Trigger
          className={cn(
            "flex items-center gap-2 rounded-md py-1 pl-1 pr-2",
            "hover:bg-muted transition-colors",
            "focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30"
          )}
          aria-label="User menu"
        >
          <Avatar.Root className="flex size-7 items-center justify-center rounded-full bg-primary text-xs font-medium text-primary-foreground">
            <Avatar.Fallback>
              {userEmail.charAt(0).toUpperCase()}
            </Avatar.Fallback>
          </Avatar.Root>
          <span className="hidden text-sm font-medium md:inline">
            {userEmail}
          </span>
        </Menu.Trigger>
        <Menu.Portal>
          <Menu.Positioner
            className={cn(
              "z-50 min-w-48 overflow-hidden rounded-md border border-border bg-popover p-1",
              "shadow-md"
            )}
            align="end"
          >
            <Menu.Popup>
              <div className="px-2 py-1.5">
                <p className="text-sm font-medium truncate">{userEmail}</p>
              </div>
              <div className="my-1 h-px bg-border" />
              <Menu.Item
                className={cn(
                  "flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm",
                  "hover:bg-accent hover:text-accent-foreground transition-colors",
                  "focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30"
                )}
                render={
                  <Link href="/settings/tokens" />
                }
              >
                <Key className="size-4" />
                API tokens
              </Menu.Item>
              <div className="my-1 h-px bg-border" />
              <Menu.Item
                className={cn(
                  "flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm",
                  "hover:bg-accent hover:text-accent-foreground transition-colors",
                  "focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30"
                )}
                render={
                  <Link href="/settings/opml" />
                }
              >
                <Upload className="size-4" />
                Import OPML
              </Menu.Item>
              <Menu.Item
                className={cn(
                  "flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm",
                  "hover:bg-accent hover:text-accent-foreground transition-colors",
                  "focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30"
                )}
                render={
                  <Link href="/settings/opml" />
                }
              >
                <FileDown className="size-4" />
                Export OPML
              </Menu.Item>
              <div className="my-1 h-px bg-border" />
              <Menu.Item
                className={cn(
                  "flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm",
                  "hover:bg-accent hover:text-accent-foreground transition-colors",
                  "focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30"
                )}
                onClick={handleSignout}
              >
                <LogOut className="size-4" />
                Sign out
              </Menu.Item>
            </Menu.Popup>
          </Menu.Positioner>
        </Menu.Portal>
      </Menu.Root>
    </header>
  )
}
