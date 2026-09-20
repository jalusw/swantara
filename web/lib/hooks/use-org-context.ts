"use client";

import { useContext } from "react";

import { OrgActiveContext } from "@/providers/org-active-provider";

export type ActiveOrg = {
  id: number;
  name?: string;
};

export function useActiveOrg(): ActiveOrg | null {
  const orgId = useContext(OrgActiveContext);
  return orgId == null ? null : { id: orgId };
}

export function useOrganizationId(): number | null {
  return useContext(OrgActiveContext);
}
