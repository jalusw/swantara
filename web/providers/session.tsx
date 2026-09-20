"use client";

import { createContext, type ReactNode, useContext } from "react";

import { hasSessionHint } from "@/lib/constants/cookies";
import { useMeOrganizationsQuery, useMeQuery } from "@/lib/hooks/use-me-query";

type SessionValue = {
  user: { id: number; email: string } | undefined;
  organizations: { id: number; name?: string }[] | undefined;
  isAuthenticated: boolean;
  isLoading: boolean;
};

export const SessionContext = createContext<SessionValue | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  const hasToken = hasSessionHint();
  const me = useMeQuery({ enabled: hasToken });
  const organizations = useMeOrganizationsQuery({ enabled: hasToken });

  const value: SessionValue = {
    user: me.data?.user,
    organizations: organizations.data?.organizations,
    isAuthenticated: !!me.data?.user,
    isLoading: me.isLoading || organizations.isLoading,
  };

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionValue {
  const ctx = useContext(SessionContext);
  if (!ctx) {
    throw new Error("useSession must be used within a SessionProvider");
  }
  return ctx;
}
