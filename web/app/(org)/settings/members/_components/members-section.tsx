"use client";

import { ShieldCheck, UserCheck, UserPlus, UsersRound } from "lucide-react";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { MetricGrid } from "@/components/metric-grid";
import { StatCard } from "@/components/stat-card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Member } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { MembersInviteDialog } from "./members-invite-dialog";
import { MembersTable, toMemberRow } from "./members-table";

export function MembersSection() {
  const [inviteOpen, setInviteOpen] = useState(false);
  const query = useOrgListQuery<{ members: Member[] }, Record<string, never>>(
    "members",
    (organizationId) => getSwantaraService().members.list(organizationId),
  );

  const rows = (query.data?.members ?? []).map(toMemberRow);
  const total = rows.length;
  const active = rows.filter((row) => row.status === "active").length;
  const invited = rows.filter((row) => row.status === "invited").length;
  const admins = rows.filter((row) => row.role === "owner" || row.role === "admin").length;

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <MetricGrid>
        <StatCard label={"Total members"} value={String(total)} icon={UsersRound} />
        <StatCard label={"Active members"} value={String(active)} icon={UserCheck} />
        <StatCard label={"Invited"} value={String(invited)} icon={UserPlus} />
        <StatCard label={"Admins"} value={String(admins)} icon={ShieldCheck} />
      </MetricGrid>

      <Card>
        <CardHeader className="border-b">
          <CardTitle>{"All members"}</CardTitle>
          <CardDescription>{"People with access to this organization."}</CardDescription>
        </CardHeader>
        <CardContent>
          <MembersTable
            members={rows}
            onInviteClick={() => setInviteOpen(true)}
            status={
              query.isLoading
                ? { type: "loading" }
                : query.isError
                  ? {
                      type: "error",
                      message: query.error.message,
                      onRetry: () => void query.refetch(),
                    }
                  : undefined
            }
          />
        </CardContent>
      </Card>
      <MembersInviteDialog
        open={inviteOpen}
        onOpenChange={setInviteOpen}
        onInvited={() => void query.refetch()}
      />
    </div>
  );
}
