"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { OrgModule } from "@/lib/services/swantara/identity";
import { getSwantaraService } from "@/lib/services/swantara/service";
import { useOrganizationId } from "./use-org-context";
import { orgListQueryKey } from "./use-org-query";

export function orgModulesQueryKey(organizationId: number) {
  return orgListQueryKey("modules", organizationId);
}

export function useOrgModules() {
  const organizationId = useOrganizationId();
  return useQuery({
    queryKey: orgModulesQueryKey(organizationId ?? 0),
    queryFn: () => getSwantaraService().organizations.listModules(organizationId as number),
    enabled: organizationId != null,
  });
}

export function useActiveModuleIds(): Set<string> | null {
  const { data } = useOrgModules();
  if (!data) return null;
  return new Set(data.modules.filter((module) => module.active).map((module) => module.moduleId));
}

export function useUpdateOrgModule() {
  const organizationId = useOrganizationId();
  const queryClient = useQueryClient();
  const queryKey = orgModulesQueryKey(organizationId ?? 0);
  return useMutation({
    mutationFn: ({ moduleId, active }: { moduleId: string; active: boolean }) =>
      getSwantaraService().organizations.updateModule(organizationId as number, {
        moduleId,
        active,
      }),
    onMutate: async ({ moduleId, active }) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<{ modules: OrgModule[] }>(queryKey);
      if (previous) {
        queryClient.setQueryData<{ modules: OrgModule[] }>(queryKey, {
          modules: previous.modules.map((module) =>
            module.moduleId === moduleId ? { ...module, active } : module,
          ),
        });
      }
      return { previous };
    },
    onError: (_error, _value, context) => {
      if (context?.previous) queryClient.setQueryData(queryKey, context.previous);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
}
