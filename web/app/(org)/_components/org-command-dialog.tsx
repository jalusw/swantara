"use client";

import { useRouter } from "next/navigation";
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
  const { has } = usePermissions();

  const visibleItems = orgNavItems.filter(
    (item) => item.permission == null || has(item.permission),
  );

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"Search"}
      description={"Search across your organization"}
    >
      <CommandInput placeholder={"Search..."} />
      <CommandList>
        <CommandEmpty>{"No results found."}</CommandEmpty>
        <CommandGroup heading={"Pages"}>
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
