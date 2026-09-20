"use client";

import { useQuery } from "@tanstack/react-query";
import { getSwantaraService } from "@/lib/services/swantara";

export function useProductQuery(orgId: string, itemId: string) {
  return useQuery({
    queryKey: ["item", orgId, itemId],
    queryFn: () => getSwantaraService().products.get(Number(orgId), Number(itemId)),
    enabled: Boolean(orgId && itemId),
    select: (data) => data.item,
  });
}

export function useItemVariantsQuery(orgId: string, itemId: string) {
  return useQuery({
    queryKey: ["itemVariants", orgId, itemId],
    queryFn: () => getSwantaraService().products.variants.list(Number(orgId), Number(itemId)),
    enabled: Boolean(orgId && itemId),
    select: (data) => data.variants,
  });
}
