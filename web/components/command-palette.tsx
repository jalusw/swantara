"use client";

import { usePathname } from "next/navigation";
import { createContext, type ReactNode, useContext, useEffect, useMemo, useState } from "react";

type CommandPaletteContextValue = {
  open: boolean;
  setOpen: (open: boolean) => void;
};

const CommandPaletteContext = createContext<CommandPaletteContextValue | null>(null);

export function CommandPaletteProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();
  const isOrgPage = pathname.includes("/org/");

  useEffect(() => {
    if (!isOrgPage) return;

    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setOpen((current) => !current);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [isOrgPage]);

  useEffect(() => {
    if (!isOrgPage && open) setOpen(false);
  }, [isOrgPage, open]);

  const value = useMemo<CommandPaletteContextValue>(() => ({ open, setOpen }), [open]);

  return (
    <div data-slot="command-palette">
      <CommandPaletteContext.Provider value={value}>{children}</CommandPaletteContext.Provider>
    </div>
  );
}

export function useCommandPalette() {
  const context = useContext(CommandPaletteContext);
  if (!context) {
    throw new Error("useCommandPalette must be used within a CommandPaletteProvider");
  }
  return context;
}
