"use client";

import { useQuery } from "@tanstack/react-query";
import type { ItemVariant } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

export function useItemVariants(orgId: string, itemId: string | undefined) {
  return useQuery<{ variants: ItemVariant[] }, Error>({
    queryKey: ["itemVariants", orgId, itemId],
    queryFn: () => getSwantaraService().products.variants.list(Number(orgId), Number(itemId)),
    enabled: Boolean(orgId && itemId),
    staleTime: 5 * 60 * 1000,
  });
}
