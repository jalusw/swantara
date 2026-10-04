"use client";

import { initials } from "@dicebear/collection";
import { createAvatar } from "@dicebear/core";
import { Building2, Check, ChevronsUpDown, Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/dropdown";
import { useMeOrganizationsQuery } from "@/lib/hooks/use-me-query";
import { useOrganizationId } from "@/lib/hooks/use-org-context";
import { setActiveOrg } from "@/lib/server/active-org-actions";
import { cn } from "@/lib/utils";
import { CreateOrganizationDialog } from "./create-organization-dialog";

function OrgLogo({
  logo,
  name,
  className,
}: {
  logo?: string | null;
  name?: string;
  className?: string;
}) {
  const tOrg = useTranslations("Organization" as unknown as "Common");
  const tx = tOrg as unknown as (key: string, values?: Record<string, string>) => string;
  const fallbackSrc = useMemo(() => {
    if (!name) return null;
    return createAvatar(initials, {
      seed: name,
      backgroundColor: ["4f46e5"],
      textColor: ["ffffff"],
      size: 128,
    }).toDataUri();
  }, [name]);

  const src = logo || fallbackSrc;

  if (src) {
    return (
      // biome-ignore lint/performance/noImgElement: org logo from API or dicebear
      <img
        src={src}
        alt={name ? tx("logoAltWithName", { name }) : tx("logoAlt")}
        className={cn("size-8 shrink-0 rounded-lg object-cover", className)}
      />
    );
  }
  return (
    <span
      className={cn(
        "grid size-8 shrink-0 place-items-center rounded-lg bg-primary/15 text-primary ring-1 ring-primary/30",
        className,
      )}
    >
      <Building2 className="size-4" aria-hidden />
    </span>
  );
}

export function OrgBrand() {
  const router = useRouter();
  const tOrg = useTranslations("Organization" as unknown as "Common");
  const tx = tOrg as unknown as (key: string) => string;
  const [createOpen, setCreateOpen] = useState(false);
  const activeOrgId = useOrganizationId();
  const { data, isLoading } = useMeOrganizationsQuery();
  const organizations = data?.organizations ?? [];

  const current = organizations.find((org) => org.id === activeOrgId) ?? organizations[0];

  const handleSelect = async (id: number) => {
    await setActiveOrg(id);
    router.push("/dashboard");
    router.refresh();
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button
            type="button"
            className="group/brand flex w-full items-center gap-2.5 rounded-md p-2 text-left outline-none transition-colors hover:bg-sidebar-accent/60 focus-visible:ring-2 ring-offset-2 ring-offset-sidebar focus-visible:ring-sidebar-ring"
          >
            <OrgLogo logo={current?.logo} name={current?.name} />
            <span className="flex min-w-0 flex-1 flex-col gap-0.5">
              <span className="truncate text-sm">
                {current?.name ?? (isLoading ? "…" : tx("selectOrganization"))}
              </span>
              <span className="truncate text-xs text-muted-foreground">{tx("label")}</span>
            </span>
            <ChevronsUpDown className="size-4 shrink-0 text-muted-foreground" aria-hidden />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="min-w-56" align="start">
          <DropdownMenuGroup>
            <DropdownMenuLabel>{tx("label")}</DropdownMenuLabel>
            {organizations.length === 0 ? (
              <DropdownMenuItem disabled>
                {isLoading ? tx("loading") : tx("empty")}
              </DropdownMenuItem>
            ) : (
              organizations.map((org) => {
                const isActive = org.id === current?.id;
                return (
                  <DropdownMenuItem
                    key={org.id}
                    onClick={() => handleSelect(org.id)}
                    data-active={isActive}
                  >
                    <OrgLogo
                      logo={org.logo}
                      name={org.name}
                      className={cn("size-6", isActive && "ring-1 ring-primary/30")}
                    />
                    <span className="flex-1 truncate">{org.name}</span>
                    {isActive ? <Check className="text-primary" /> : null}
                  </DropdownMenuItem>
                );
              })
            )}
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => {
              setCreateOpen(true);
            }}
          >
            <Plus />
            {tx("create")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <CreateOrganizationDialog open={createOpen} onOpenChange={setCreateOpen} />
    </>
  );
}
