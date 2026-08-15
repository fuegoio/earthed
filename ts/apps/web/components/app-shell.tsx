"use client"

import { useState } from "react"
import Link from "next/link"
import { Menu as MenuIcon } from "lucide-react"
import { AppSidebar } from "@/components/app-sidebar"
import { ThemeToggle } from "@/components/theme-toggle"
import { Logo } from "@/components/logo"
import { Button } from "@workspace/ui/components/button"

export function AppShell({
  children,
  userEmail,
}: {
  children: React.ReactNode
  userEmail: string
}) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="mx-auto flex h-svh w-full max-w-5xl overflow-hidden">
      <AppSidebar
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        userEmail={userEmail}
      />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background px-3 lg:hidden">
          <Button
            variant="ghost"
            size="icon"
            aria-label="Toggle sidebar"
            onClick={() => setSidebarOpen(true)}
          >
            <MenuIcon className="size-4" />
          </Button>
          <Link
            href="/"
            className="flex items-center gap-2 font-serif text-lg font-bold"
          >
            <Logo className="size-5" />
            Planetary
          </Link>
          <div className="ml-auto">
            <ThemeToggle />
          </div>
        </header>
        <main className="flex-1 overflow-y-auto">{children}</main>
      </div>
    </div>
  )
}
