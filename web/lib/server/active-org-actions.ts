"use server";

import { cookies } from "next/headers";
import { activeOrgCookieKey, activeOrgCookieOptions } from "./active-org-cookie";

export async function setActiveOrg(id: number): Promise<void> {
  if (!Number.isInteger(id) || id <= 0) return;
  const store = await cookies();
  store.set(activeOrgCookieKey, String(id), activeOrgCookieOptions);
}

export async function clearActiveOrg(): Promise<void> {
  const store = await cookies();
  store.delete(activeOrgCookieKey);
}
