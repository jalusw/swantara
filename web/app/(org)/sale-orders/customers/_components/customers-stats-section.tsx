"use client";

import { UserCheck, UserPlus, Users } from "lucide-react";
import { MetricGrid } from "@/components/metric-grid";
import { StatCard } from "@/components/stat-card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

function countNewThisMonth(contacts: Contact[]) {
  const now = new Date();
  return contacts.filter((contact) => {
    if (!contact.createdAt) return false;
    const created = new Date(contact.createdAt);
    return created.getFullYear() === now.getFullYear() && created.getMonth() === now.getMonth();
  }).length;
}

export function CustomersStats() {
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  if (contactsQuery.isLoading) {
    return (
      <output aria-busy="true" aria-label={"Customer overview"}>
        <MetricGrid>
          {[0, 1, 2].map((key) => (
            <div key={key} className="h-28 animate-pulse rounded-xl bg-muted" aria-hidden />
          ))}
        </MetricGrid>
      </output>
    );
  }

  if (contactsQuery.isError) {
    return <p className="text-sm text-muted-foreground">{"Something went wrong."}</p>;
  }

  const contacts = contactsQuery.data?.contacts ?? [];

  const stats: Array<{ key: string; label: string; value: string; icon: typeof Users }> = [
    { key: "total", label: "Total customers", value: formatNumber(contacts.length), icon: Users },
    {
      key: "active",
      label: "Active",
      value: formatNumber(contacts.filter((contact) => contact.active).length),
      icon: UserCheck,
    },
    {
      key: "newThisMonth",
      label: "New this month",
      value: formatNumber(countNewThisMonth(contacts)),
      icon: UserPlus,
    },
  ];

  return (
    <MetricGrid>
      {stats.map((stat) => (
        <StatCard key={stat.key} label={stat.label} value={stat.value} icon={stat.icon} />
      ))}
    </MetricGrid>
  );
}
