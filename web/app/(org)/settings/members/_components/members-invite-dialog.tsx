"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Avatar, AvatarFallback } from "@/components/avatar";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { useDebouncedValue } from "@/lib/hooks/use-debounced-value";
import { useOrganizationId } from "@/lib/hooks/use-org-context";
import type { PublicUser } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

export type MembersInviteDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onInvited?: () => void;
};

function isBlockedQuery(value: string): boolean {
  const trimmed = value.trim();
  if (trimmed.includes("@")) return true;
  return /^[+]?[0-9][0-9\s\-()]{5,}$/.test(trimmed);
}

function displayName(user: PublicUser): string {
  return user.lastName ? `${user.firstName} ${user.lastName}` : user.firstName;
}

function initialsFor(name: string): string {
  return name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

export function MembersInviteDialog({ open, onOpenChange, onInvited }: MembersInviteDialogProps) {
  const t = useTranslations("Settings");
  const organizationId = useOrganizationId();
  const queryClient = useQueryClient();
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<PublicUser | null>(null);
  const debounced = useDebouncedValue(query, 300);
  const trimmed = debounced.trim();
  const blocked = trimmed.length > 0 && isBlockedQuery(trimmed);
  const enabled = open && trimmed.length >= 2 && !blocked;

  const searchQuery = useQuery({
    queryKey: ["users", "search", trimmed],
    queryFn: () => getSwantaraService().users.search(trimmed),
    enabled,
  });

  const suggestions = searchQuery.data?.users ?? [];

  const inviteMutation = useMutation({
    mutationFn: (userId: number) =>
      getSwantaraService().members.create(organizationId as number, {
        userId,
        organizationId: organizationId as number,
      }),
    onSuccess: () => {
      toast.success(t("memberInvited"));
      void queryClient.invalidateQueries({ queryKey: ["members"] });
      setQuery("");
      setSelected(null);
      onOpenChange(false);
      onInvited?.();
    },
    onError: () => {
      toast.error(t("inviteFailed"));
    },
  });

  function handleOpenChange(next: boolean) {
    if (!next) {
      setQuery("");
      setSelected(null);
    }
    onOpenChange(next);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("inviteTitle")}</DialogTitle>
          <DialogDescription>{t("inviteDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-3">
          <Input
            placeholder={t("inviteSearchPlaceholder")}
            value={query}
            onChange={(event) => {
              setQuery(event.target.value);
              setSelected(null);
            }}
            aria-label={t("searchUsers")}
          />
          {trimmed.length > 0 && trimmed.length < 2 && (
            <p className="text-sm text-muted-foreground">{t("minChars")}</p>
          )}
          {blocked && <p className="text-sm text-muted-foreground">{t("blockedQueryHint")}</p>}
          {enabled && searchQuery.isPending && (
            <p className="text-sm text-muted-foreground">{t("loadingUsers")}</p>
          )}
          {enabled && searchQuery.isError && (
            <p className="text-sm text-muted-foreground">{t("loadFailed")}</p>
          )}
          {enabled &&
            !searchQuery.isPending &&
            !searchQuery.isError &&
            suggestions.length === 0 && (
              <p className="text-sm text-muted-foreground">{t("noUsersFound")}</p>
            )}
          {suggestions.length > 0 && (
            <ul className="flex max-h-64 flex-col gap-1 overflow-auto">
              {suggestions.map((user) => {
                const name = displayName(user);
                const active = selected?.id === user.id;
                return (
                  <li key={user.id}>
                    <button
                      type="button"
                      onClick={() => setSelected(user)}
                      aria-pressed={active}
                      className={`flex min-h-[44px] w-full items-center gap-3 rounded-lg p-2 text-left ${
                        active ? "bg-accent" : ""
                      }`}
                    >
                      <Avatar size="sm">
                        <AvatarFallback>{initialsFor(name)}</AvatarFallback>
                      </Avatar>
                      <span className="flex flex-col">
                        <span>{name}</span>
                        <span className="text-xs text-muted-foreground">@{user.username}</span>
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
        <DialogFooter>
          <Button
            disabled={!selected || !organizationId || inviteMutation.isPending}
            onClick={() => {
              if (selected) inviteMutation.mutate(selected.id);
            }}
          >
            {inviteMutation.isPending ? t("inviting") : t("invite")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
