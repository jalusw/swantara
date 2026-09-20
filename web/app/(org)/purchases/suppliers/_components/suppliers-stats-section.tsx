"use client";

import { FolderKanban, Store, TrendingUp, Truck } from "lucide-react";
import { MetricGrid } from "@/components/metric-grid";
import { StatCard } from "@/components/stat-card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

export function SuppliersStats() {
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const contacts = contactsQuery.data?.contacts ?? [];
  const activeContacts = contacts.filter((p) => p.active);

  const stats: Array<{ key: string; value: string; icon: typeof Truck }> = [
    { key: "total", value: formatNumber(activeContacts.length), icon: Truck },
    { key: "active", value: formatNumber(activeContacts.length), icon: Store },
    { key: "categories", value: "—", icon: FolderKanban },
    { key: "avgLeadTime", value: "—", icon: TrendingUp },
  ];

  return (
    <MetricGrid>
      {stats.map((stat) => (
        <StatCard
          key={stat.key}
          label={humanizeKey(String(stat.key))}
          value={stat.value}
          icon={stat.icon}
        />
      ))}
    </MetricGrid>
  );
}
