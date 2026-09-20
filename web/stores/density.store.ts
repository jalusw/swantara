"use client";

import cookie from "cookiejs";
import { create } from "zustand";

export const COMPACT = "compact";
export const NORMAL = "normal";
export const COMFORTABLE = "comfortable";
const DENSITIES = [COMPACT, NORMAL, COMFORTABLE] as const;

export type DataDensity = (typeof DENSITIES)[number];

type DensityState = {
  density: DataDensity;
};

type DensityActions = {
  setDensity: (density: DataDensity) => void;
  hydrateDensity: () => void;
};

export type DensityStore = DensityState & DensityActions;

const STORAGE_KEY = "data-density";
const COOKIE_KEY = "data-density";

function getStoredDensity(): DataDensity {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === COMPACT || stored === NORMAL || stored === COMFORTABLE) {
      return stored;
    }
  } catch {}
  return NORMAL;
}

function setCookie(value: string) {
  cookie.set(COOKIE_KEY, value, {
    path: "/",
    expires: 365,
    sameSite: "Lax",
    secure: process.env.NODE_ENV === "production",
  });
}

export const useDensityStore = create<DensityStore>((set) => ({
  density: NORMAL,
  setDensity: (density: DataDensity) => {
    set({ density });
    try {
      localStorage.setItem(STORAGE_KEY, density);
    } catch {}
    setCookie(density);
  },
  hydrateDensity: () => {
    set({ density: getStoredDensity() });
  },
}));
