"use client";

import { create } from "zustand";

import { APP_THEMES, type AppTheme } from "@/components/theme";

export const [LIGHT, DARK] = APP_THEMES;

const THEMES = [...APP_THEMES, "system"] as const;

export type Theme = (typeof THEMES)[number];

export type ResolvedTheme = AppTheme;

type ThemeState = {
  theme: Theme;
  resolvedTheme: ResolvedTheme;
};

type ThemeActions = {
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
  hydrateTheme: () => void;
};

export type ThemeStore = ThemeState & ThemeActions;

const STORAGE_KEY = "theme";

function getSystemTheme(): ResolvedTheme {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? DARK : LIGHT;
}

function resolveTheme(theme: Theme): ResolvedTheme {
  return theme === "system" ? getSystemTheme() : theme;
}

const COLOR_SCHEMES: Record<ResolvedTheme, string> = {
  [LIGHT]: "light",
  [DARK]: "dark",
};

function applyTheme(theme: Theme) {
  const resolved = resolveTheme(theme);
  const root = document.documentElement;
  root.classList.remove(LIGHT, DARK);
  root.classList.add(resolved);
  root.style.colorScheme = COLOR_SCHEMES[resolved];
}

function getStoredTheme(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === LIGHT || stored === DARK || stored === "system") {
      return stored;
    }
  } catch {}
  return DARK;
}

export const useThemeStore = create<ThemeStore>((set) => ({
  theme: DARK,
  resolvedTheme: DARK,
  setTheme: (theme: Theme) => {
    set({ theme, resolvedTheme: resolveTheme(theme) });
    applyTheme(theme);
    try {
      localStorage.setItem(STORAGE_KEY, theme);
    } catch {}
  },
  toggleTheme: () => {
    const { resolvedTheme } = useThemeStore.getState();
    const next: ResolvedTheme = resolvedTheme === DARK ? LIGHT : DARK;
    set({ theme: next, resolvedTheme: next });
    applyTheme(next);
    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch {}
  },
  hydrateTheme: () => {
    const theme = getStoredTheme();
    set({ theme, resolvedTheme: resolveTheme(theme) });
    applyTheme(theme);
  },
}));

export function listenToSystemTheme(): () => void {
  const mq = window.matchMedia("(prefers-color-scheme: dark)");
  const handler = () => {
    const { theme } = useThemeStore.getState();
    if (theme === "system") {
      useThemeStore.setState({ resolvedTheme: getSystemTheme() });
    }
  };
  mq.addEventListener("change", handler);
  return () => mq.removeEventListener("change", handler);
}
