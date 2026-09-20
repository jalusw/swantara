import { queryOptions } from "@tanstack/react-query";
import type { SwantaraService } from "@/lib/services/swantara";

export const ME_QUERY_KEY = ["me"] as const;
export const ME_ORGANIZATIONS_QUERY_KEY = ["me", "organizations"] as const;

export function permissionsQueryKey(organizationId: number) {
  return ["permissions", organizationId, {}] as const;
}

export function meQueryOptions(service: SwantaraService) {
  return queryOptions({
    queryKey: ME_QUERY_KEY,
    queryFn: () => service.me.me(),
  });
}

export function meOrganizationsQueryOptions(service: SwantaraService) {
  return queryOptions({
    queryKey: ME_ORGANIZATIONS_QUERY_KEY,
    queryFn: () => service.me.organizations(),
  });
}

export function permissionsQueryOptions(service: SwantaraService, organizationId: number) {
  return queryOptions({
    queryKey: permissionsQueryKey(organizationId),
    queryFn: () => service.me.permissions(organizationId),
  });
}
