"use client"

import { useState } from "react"
import { AppSidebar } from "@/components/app-sidebar"
import { AppTopbar } from "@/components/app-topbar"

export function AppShell({
  children,
  userEmail,
}: {
  children: React.ReactNode
  userEmail: string
}) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="flex min-h-svh">
      <AppSidebar
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
      />
      <div className="flex min-w-0 flex-1 flex-col">
        <AppTopbar
          onMenuClick={() => setSidebarOpen(true)}
          userEmail={userEmail}
        />
        <main className="flex-1">{children}</main>
      </div>
    </div>
  )
}
