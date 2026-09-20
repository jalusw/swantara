"use client";

import { useEffect } from "react";
import { listenToSystemTheme, useThemeStore } from "@/stores/theme.store";

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const hydrateTheme = useThemeStore((s) => s.hydrateTheme);

  useEffect(() => {
    hydrateTheme();
  }, [hydrateTheme]);

  useEffect(() => listenToSystemTheme(), []);

  return <>{children}</>;
}
