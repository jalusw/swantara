"use client";

import { Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { setActiveOrg } from "@/lib/server/active-org-actions";

export function AcceptOrganization({ orgId }: { orgId: number }) {
  const router = useRouter();
  const t = useTranslations("Onboarding");
  const tx = t as unknown as (key: string) => string;
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setActiveOrg(orgId).then(
      () => {
        if (cancelled) return;
        router.replace("/dashboard");
        router.refresh();
      },
      () => {
        if (!cancelled) setFailed(true);
      },
    );
    return () => {
      cancelled = true;
    };
  }, [orgId, router]);

  return (
    <div className="flex flex-col items-center gap-4 py-12 text-center">
      {failed ? (
        <p className="text-sm text-muted-foreground">{tx("selectOrgFailed")}</p>
      ) : (
        <>
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
          <p className="text-sm text-muted-foreground">{t("openingDashboard")}</p>
        </>
      )}
    </div>
  );
}
