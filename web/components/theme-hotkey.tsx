"use client";

import { useEffect } from "react";
import { useThemeStore } from "@/stores/theme.store";

export function ThemeHotkey() {
  const toggleTheme = useThemeStore((s) => s.toggleTheme);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.shiftKey && event.key.toLowerCase() === "t") {
        event.preventDefault();
        toggleTheme();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [toggleTheme]);

  return <div data-slot="theme-hotkey" />;
}
