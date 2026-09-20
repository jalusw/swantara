"use client";

import {
  keepPreviousData,
  type QueryKey,
  queryOptions,
  type UseQueryOptions,
  useQuery,
} from "@tanstack/react-query";
import { useOrganizationId } from "./use-org-context";

export function orgListQueryKey(
  resource: string,
  organizationId: number,
  params?: Record<string, unknown>,
): QueryKey {
  return [resource, organizationId, { ...params }];
}

export function orgDetailQueryKey(
  resource: string,
  organizationId: number,
  id: string | number,
): QueryKey {
  return [resource, organizationId, id];
}

export function orgListQueryOptions<
  TData,
  TParams extends Record<string, unknown>,
  TSelected = TData,
>(
  resource: string,
  organizationId: number | null,
  fetcher: (organizationId: number, params: TParams) => Promise<TData>,
  params?: TParams,
  options?: Omit<UseQueryOptions<TData, Error, TSelected>, "queryKey" | "queryFn" | "enabled">,
) {
  const resolvedParams = (params ?? {}) as TParams;
  return queryOptions({
    queryKey: orgListQueryKey(resource, organizationId ?? 0, resolvedParams),
    queryFn: () => fetcher(organizationId as number, resolvedParams),
    enabled: organizationId != null,
    placeholderData: keepPreviousData,
    ...options,
  });
}

export function useOrgListQuery<TData, TParams extends Record<string, unknown>, TSelected = TData>(
  resource: string,
  fetcher: (organizationId: number, params: TParams) => Promise<TData>,
  params?: TParams,
  options?: Omit<UseQueryOptions<TData, Error, TSelected>, "queryKey" | "queryFn" | "enabled">,
) {
  const organizationId = useOrganizationId();
  return useQuery(orgListQueryOptions(resource, organizationId, fetcher, params, options));
}

export function useOrgQuery<TData, TSelected = TData>(
  resource: string,
  id: string | number,
  fetcher: (organizationId: number) => Promise<TData>,
  options?: Omit<UseQueryOptions<TData, Error, TSelected>, "queryKey" | "queryFn" | "enabled">,
) {
  const organizationId = useOrganizationId();

  return useQuery({
    queryKey: orgDetailQueryKey(resource, organizationId ?? 0, id),
    queryFn: () => fetcher(organizationId as number),
    enabled: organizationId != null,
    ...options,
  });
}
