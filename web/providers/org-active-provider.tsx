"use client";

import { createContext } from "react";

export const OrgActiveContext = createContext<number | null>(null);

export function OrgActiveProvider({
  orgId,
  children,
}: {
  orgId: number;
  children: React.ReactNode;
}) {
  return <OrgActiveContext.Provider value={orgId}>{children}</OrgActiveContext.Provider>;
}
