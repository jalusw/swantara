"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { UserPlus } from "lucide-react";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Avatar, AvatarFallback } from "@/components/avatar";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import type { DataTableStatus } from "@/components/tanstack-table";
import type { Member } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

export type MemberRow = {
  id: string;
  name: string;
  email: string;
  role: "owner" | "admin" | "member";
  status: "active" | "invited" | "inactive";
  addedAt: string;
};

export function memberStatus(user: Member["user"]): MemberRow["status"] {
  if (!user) return "invited";
  return user.active ? "active" : "inactive";
}

export function toMemberRow(member: Member): MemberRow {
  const user = member.user;
  const name = user ? `${user.firstName}${user.lastName ? ` ${user.lastName}` : ""}` : "—";
  const roleCode = member.roles?.[0]?.code.toLowerCase() ?? "";
  const role = roleCode.includes("owner")
    ? "owner"
    : roleCode.includes("admin")
      ? "admin"
      : "member";
  const status = memberStatus(user);

  return {
    id: String(member.id),
    name,
    email: user?.email ?? "—",
    role,
    status,
    addedAt: formatDate(member.createdAt),
  };
}

function initials(name: string) {
  return name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2);
}

export function MembersTable({
  members,
  status,
  onInviteClick,
}: {
  members: MemberRow[];
  status?: DataTableStatus;
  onInviteClick?: () => void;
}) {
  const t = useTranslations("Settings");
  const memberRole = (role: string) => (t as unknown as (k: string) => string)(`role_${role}`);
  const memberStatus = (st: string) =>
    (t as unknown as (k: string) => string)(`memberStatus_${st}`);
  const columns: ColumnDef<MemberRow>[] = [
    {
      accessorKey: "name",
      header: t("colMember"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <Avatar size="sm">
            <AvatarFallback>{initials(row.original.name)}</AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <span className="">{row.original.name}</span>
            <span className="text-xs text-muted-foreground">{row.original.email}</span>
          </div>
        </div>
      ),
    },
    {
      accessorKey: "role",
      header: t("colRole"),
      cell: ({ row }) => (
        <Badge variant={row.original.role === "owner" ? "default" : "secondary"}>
          {memberRole(row.original.role)}
        </Badge>
      ),
    },
    {
      accessorKey: "status",
      header: t("colStatus"),
      cell: ({ row }) => {
        const toneMap: Record<MemberRow["status"], "success" | "info" | "neutral"> = {
          active: "success",
          invited: "info",
          inactive: "neutral",
        };
        return (
          <StateBadge
            tone={toneMap[row.original.status]}
            label={memberStatus(row.original.status)}
          />
        );
      },
    },
    {
      accessorKey: "addedAt",
      header: t("colAdded"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.addedAt}</span>,
    },
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={members}
      getRowId={(row) => row.id}
      searchKeys={["name", "email"]}
      statusKey="status"
      statusOptions={[
        { value: "active", label: memberStatus("active") },
        { value: "invited", label: memberStatus("invited") },
        { value: "inactive", label: memberStatus("inactive") },
      ]}
      searchPlaceholder={t("membersSearchPlaceholder")}
      filterLabel={t("filterByStatus")}
      allLabel={t("allStatuses")}
      ariaLabel={t("allMembers")}
      status={status}
      exportFileName={t("allMembers")}
      actions={
        <Button size="sm" onClick={onInviteClick}>
          <UserPlus />
          <span>{t("invite")}</span>
        </Button>
      }
    />
  );
}
