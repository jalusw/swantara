"use client";

import { FolderKanban, Store, TrendingUp, Truck } from "lucide-react";
import { useTranslations } from "next-intl";
import { MetricGrid } from "@/components/metric-grid";
import { StatCard } from "@/components/stat-card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

export function SuppliersStats() {
  const t = useTranslations("Purchases");
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const contacts = contactsQuery.data?.contacts ?? [];
  const activeContacts = contacts.filter((p) => p.active);

  function statLabel(key: string): string {
    try {
      return (t as unknown as (k: string) => string)(`supplierStat.${key}`);
    } catch {
      return key;
    }
  }

  const stats: Array<{ key: string; value: string; icon: typeof Truck }> = [
    { key: "total", value: formatNumber(activeContacts.length), icon: Truck },
    { key: "active", value: formatNumber(activeContacts.length), icon: Store },
    { key: "categories", value: "—", icon: FolderKanban },
    { key: "avgLeadTime", value: "—", icon: TrendingUp },
  ];

  return (
    <MetricGrid>
      {stats.map((stat) => (
        <StatCard key={stat.key} label={statLabel(stat.key)} value={stat.value} icon={stat.icon} />
      ))}
    </MetricGrid>
  );
}
