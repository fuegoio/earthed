"use client";

import { useState, useCallback } from "react";
import { usePathname } from "next/navigation";
import { AppSidebar } from "@/components/app-sidebar";
import { SettingsSidebar } from "@/components/settings-sidebar";
import { ShellContext } from "@/components/shell-context";

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
  const openSidebar = useCallback(() => setSidebarOpen(true), []);

  return (
    <ShellContext value={{ openSidebar }}>
      <div className="flex h-svh overflow-hidden">
        {/* Left filler — extends sidebar bg to the viewport edge on wide screens */}
        <div className="hidden flex-1 bg-sidebar lg:block" />

        <div className="flex w-full min-w-0 max-w-5xl shrink-0 overflow-hidden">
          <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} userEmail={userEmail} />
          <div className="flex min-w-0 flex-1 flex-col">
            <main className="flex-1 overflow-y-auto bg-background">{children}</main>
          </div>
        </div>

        {/* Right filler — matches main content bg on wide screens */}
        <div className="hidden flex-1 bg-background lg:block" />
      </div>
    </ShellContext>
  );
}
