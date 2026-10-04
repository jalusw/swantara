"use server";

import { cookies } from "next/headers";
import { defaultLocale, isValidLocale, type Locale, localeCookieName } from "./config";

export async function setAppLocale(locale: Locale) {
  const value = isValidLocale(locale) ? locale : defaultLocale;
  const store = await cookies();
  store.set(localeCookieName, value, {
    path: "/",
    maxAge: 60 * 60 * 24 * 365,
    sameSite: "lax",
  });
}

export async function getAppLocale(): Promise<Locale> {
  const store = await cookies();
  const raw = store.get(localeCookieName)?.value;
  return isValidLocale(raw) ? raw : defaultLocale;
}
