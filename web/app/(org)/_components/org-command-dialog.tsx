"use client";

import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/command";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { navLabel, orgNavItems } from "./nav-items";

export type OrgCommandDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function OrgCommandDialog({ open, onOpenChange }: OrgCommandDialogProps) {
  const router = useRouter();
  const t = useTranslations("Search");
  const tx = t as unknown as (key: string) => string;
  const { has } = usePermissions();

  const visibleItems = orgNavItems.filter(
    (item) => item.permission == null || has(item.permission),
  );

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title={tx("title")}
      description={tx("description")}
    >
      <CommandInput placeholder={t("placeholder")} />
      <CommandList>
        <CommandEmpty>{tx("noResults")}</CommandEmpty>
        <CommandGroup heading={tx("pages")}>
          {visibleItems.map((item) => {
            const Icon = item.icon;
            const label = navLabel(item.key);
            return (
              <CommandItem
                key={item.key}
                value={label}
                onSelect={() => {
                  onOpenChange(false);
                  router.push(item.href);
                }}
              >
                <Icon />
                <span>{label}</span>
              </CommandItem>
            );
          })}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}
