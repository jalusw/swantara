"use client";

import { useQueryClient } from "@tanstack/react-query";
import { Check, LogOut, Moon, Settings, Sun, UserRound } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Avatar, AvatarFallback } from "@/components/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/dropdown";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/tooltip";
import { getCsrfToken } from "@/lib/constants/cookies";
import { clearActiveOrg } from "@/lib/server/active-org-actions";
import { useThemeStore } from "@/stores/theme.store";

const currentUser = {
  name: "John Doe",
  email: "john@acme.com",
};

export function OrgUserMenu() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const t = useTranslations("UserMenu");
  const initials = currentUser.name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .toUpperCase();

  const handleLogout = async () => {
    try {
      const response = await fetch("/api/v1/auth/logout", {
        method: "POST",
        headers: {
          "x-csrf-token": getCsrfToken() ?? "",
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({}),
      });
      if (!response.ok) throw new Error("logout failed");
      queryClient.clear();
      await clearActiveOrg();
      router.push("/login");
    } catch {
      toast.error(t("logoutFailed"));
    }
  };

  return (
    <DropdownMenu>
      <Tooltip>
        <TooltipTrigger
          render={
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                aria-label={t("accountMenu")}
                className="rounded-full outline-none ring-ring transition-shadow focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring"
              >
                <Avatar size="default">
                  <AvatarFallback>{initials}</AvatarFallback>
                </Avatar>
              </button>
            </DropdownMenuTrigger>
          }
        />
        <TooltipContent>{currentUser.name}</TooltipContent>
      </Tooltip>
      <DropdownMenuContent className="min-w-60" align="end">
        <DropdownMenuGroup>
          <DropdownMenuLabel>
            <span className="flex flex-col gap-0.5">
              <span className="text-sm">{currentUser.name}</span>
              <span className="text-xs font-normal text-muted-foreground">{currentUser.email}</span>
            </span>
          </DropdownMenuLabel>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem render={<Link href={"/profile"} />}>
          <UserRound />
          {t("profile")}
        </DropdownMenuItem>
        <DropdownMenuItem render={<Link href={"/settings"} />}>
          <Settings />
          {t("settings")}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuSub>
          <DropdownMenuSubTrigger>
            <Sun />
            {t("theme")}
          </DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            <DropdownMenuItem onClick={() => setTheme("light")}>
              <Sun />
              <span>{t("light")}</span>
              {theme === "light" && <Check className="ml-auto" />}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setTheme("dark")}>
              <Moon />
              <span>{t("dark")}</span>
              {theme === "dark" && <Check className="ml-auto" />}
            </DropdownMenuItem>
          </DropdownMenuSubContent>
        </DropdownMenuSub>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive" onClick={handleLogout}>
          <LogOut />
          {t("logout")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
