"use client";

import { useQueryClient } from "@tanstack/react-query";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useTransition } from "react";
import { toast } from "sonner";
import { Button } from "@/components/button";
import { getCsrfToken } from "@/lib/constants/cookies";
import { clearActiveOrg } from "@/lib/server/active-org-actions";

export function LogoutButton() {
  const [isPending, startTransition] = useTransition();
  const router = useRouter();
  const queryClient = useQueryClient();
  const tCommon = useTranslations("Common");
  const tUserMenu = useTranslations("UserMenu");

  const handleLogout = () => {
    startTransition(async () => {
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
        toast.error(tUserMenu("logoutFailed"));
      }
    });
  };

  return (
    <Button
      variant="outline-destructive"
      size="lg"
      disabled={isPending}
      aria-busy={isPending}
      onClick={handleLogout}
    >
      <LogOut aria-hidden="true" />
      <span>{tCommon("logout")}</span>
    </Button>
  );
}
