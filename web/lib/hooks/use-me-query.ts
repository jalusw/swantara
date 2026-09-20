"use client";

import { useQuery } from "@tanstack/react-query";

import { meOrganizationsQueryOptions, meQueryOptions } from "@/lib/queries/me";
import { getSwantaraService } from "@/lib/services/swantara";

export function useMeQuery({ enabled = true } = {}) {
  return useQuery({ ...meQueryOptions(getSwantaraService()), enabled });
}

export function useMeOrganizationsQuery({ enabled = true } = {}) {
  return useQuery({
    ...meOrganizationsQueryOptions(getSwantaraService()),
    enabled,
  });
}
