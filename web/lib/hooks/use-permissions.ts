"use client";

import { useQuery } from "@tanstack/react-query";
import { hasPermission } from "@/components/permission-gate";
import { permissionsQueryOptions } from "@/lib/queries/me";
import { getSwantaraService } from "@/lib/services/swantara";
import { useOrganizationId } from "./use-org-context";

export function usePermissions() {
  const organizationId = useOrganizationId();

  const query = useQuery({
    ...permissionsQueryOptions(getSwantaraService(), organizationId ?? 0),
    enabled: organizationId != null,
  });

  const codes = query.data?.permissions?.map((permission) => permission.code) ?? [];

  return {
    permissions: codes,
    has: (required: string | string[], requireAll = true) =>
      (organizationId != null && query.isPending) || hasPermission(required, codes, requireAll),
    isLoading: query.isPending,
    isError: query.isError,
    error: query.error,
  };
}
