import { type DehydratedState, dehydrate, QueryClient } from "@tanstack/react-query";

import { serverQueryStaleTime } from "@/lib/constants/query";
import {
  meOrganizationsQueryOptions,
  meQueryOptions,
  permissionsQueryOptions,
} from "@/lib/queries/me";
import { createServerSwantaraService } from "@/lib/server/api-client";

type Prefetch = (queryClient: QueryClient) => Promise<unknown>[];

async function dehydrateQueries(staleTime: number, prefetch: Prefetch) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { staleTime } },
  });

  await Promise.allSettled(prefetch(queryClient));

  return dehydrate(queryClient);
}

export async function prefetchMeData(): Promise<DehydratedState> {
  const service = await createServerSwantaraService();

  return dehydrateQueries(serverQueryStaleTime, (queryClient) => [
    queryClient.query(meQueryOptions(service)),
    queryClient.query(meOrganizationsQueryOptions(service)),
  ]);
}

export async function prefetchPermissionsData(organizationId: number): Promise<DehydratedState> {
  const service = await createServerSwantaraService();

  return dehydrateQueries(serverQueryStaleTime, (queryClient) => [
    queryClient.query(permissionsQueryOptions(service, organizationId)),
  ]);
}
