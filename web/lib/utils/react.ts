"use client";

import { useMediaQuery } from "usehooks-ts";

export function useReducedMotion(): boolean {
  return useMediaQuery("(prefers-reduced-motion: reduce)", {
    defaultValue: false,
    initializeWithValue: false,
  });
}
