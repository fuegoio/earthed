"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu as MenuIcon } from "lucide-react";
import { AppSidebar } from "@/components/app-sidebar";
import { SettingsSidebar } from "@/components/settings-sidebar";
import { Logo } from "@/components/logo";
import { Button } from "@workspace/ui/components/button";

export function AppShell({
  children,
  userEmail,
}: {
  children: React.ReactNode;
  userEmail: string;
}) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const pathname = usePathname();
  const isSettings = pathname.startsWith("/settings");

  const Sidebar = isSettings ? SettingsSidebar : AppSidebar;

  return (
    <div className="flex h-svh overflow-hidden">
      {/* Left filler — extends sidebar bg to the viewport edge on wide screens */}
      <div className="hidden flex-1 bg-sidebar lg:block" />

      <div className="flex w-full min-w-0 max-w-5xl shrink-0 overflow-hidden">
        <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} userEmail={userEmail} />
        <div className="flex min-w-0 flex-1 flex-col">
          <header className="flex shrink-0 items-center gap-2 border-b border-border bg-background py-3 pl-1 pr-4 lg:hidden">
            <Button
              variant="ghost"
              size="icon"
              aria-label="Toggle sidebar"
              onClick={() => setSidebarOpen(true)}
            >
              <MenuIcon className="size-4" />
            </Button>
            <Link href="/" className="flex items-center gap-2 font-serif text-lg font-bold">
              <Logo className="size-5" />
              Planetary
            </Link>
          </header>
          <main className="flex-1 overflow-y-auto bg-background">{children}</main>
        </div>
      </div>

      {/* Right filler — matches main content bg on wide screens */}
      <div className="hidden flex-1 bg-background lg:block" />
    </div>
  );
}
