import type { ColumnDef } from "@tanstack/react-table";
import Link from "next/link";
import { ActiveBadge } from "@/components/active-badge";
import { Badge } from "@/components/badge";

type NameColumnOpts<T> = {
  basePath: string;
  header: string;
  displayNameAccessor?: (row: T) => string | null | undefined;
};

export function nameColumn<T extends { id: string | number; name: string }>(
  opts: NameColumnOpts<T>,
): ColumnDef<T> {
  const { basePath, header, displayNameAccessor } = opts;
  return {
    accessorKey: "name",
    header,
    cell: ({ row }) => (
      <Link
        href={`/${basePath}/${row.original.id}`}
        className={
          displayNameAccessor
            ? "flex flex-col rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
            : " rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        }
      >
        {displayNameAccessor ? (
          <>
            <span className="">{displayNameAccessor(row.original) || row.original.name}</span>
            <span className="text-xs text-muted-foreground">{row.original.name}</span>
          </>
        ) : (
          row.original.name
        )}
      </Link>
    ),
  };
}

type ActiveColumnOpts = {
  header: string;
  activeLabel: string;
  inactiveLabel: string;
};

export function activeColumn<T extends { active: boolean }>(opts: ActiveColumnOpts): ColumnDef<T> {
  const { header, activeLabel, inactiveLabel } = opts;
  return {
    accessorKey: "active",
    header,
    cell: ({ row }) => (
      <ActiveBadge active={row.original.active}>
        {row.original.active ? activeLabel : inactiveLabel}
      </ActiveBadge>
    ),
  };
}

export type BadgeVariant = "default" | "secondary" | "outline" | "destructive";

type StateColumnOpts = {
  header: string;
  label: (state: string) => string;
  variant: (state: string) => BadgeVariant;
};

export function stateColumn<T extends { state: string }>(opts: StateColumnOpts): ColumnDef<T> {
  const { header, label, variant } = opts;
  return {
    accessorKey: "state",
    header,
    cell: ({ row }) => (
      <Badge variant={variant(row.original.state)}>{label(row.original.state)}</Badge>
    ),
  };
}

export function makeStateVariant(
  successStates: string[],
  defaultVariant: BadgeVariant = "secondary",
): (state: string) => BadgeVariant {
  return (state) => {
    if (successStates.includes(state)) return "default";
    if (state === "cancelled") return "outline";
    return defaultVariant;
  };
}
