"use client";

import { createContext, useContext } from "react";

type ShellContextValue = {
  openSidebar: () => void;
  closeSidebar: () => void;
};

export const ShellContext = createContext<ShellContextValue | null>(null);

export function useShell(): ShellContextValue | null {
  return useContext(ShellContext);
}
