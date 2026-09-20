"use client";

import {
  HotkeysProvider as TanstackHotkeysProvider,
  type HotkeysProviderProps as TanstackHotkeysProviderProps,
} from "@tanstack/react-hotkeys";

export type HotkeysProviderProps = {
  children: React.ReactNode;
  defaultOptions?: TanstackHotkeysProviderProps["defaultOptions"];
};

export function HotkeysProvider({ children, defaultOptions }: HotkeysProviderProps) {
  return (
    <TanstackHotkeysProvider defaultOptions={defaultOptions}>{children}</TanstackHotkeysProvider>
  );
}
