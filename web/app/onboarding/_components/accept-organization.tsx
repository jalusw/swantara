"use client";

import { Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { setActiveOrg } from "@/lib/server/active-org-actions";

export function AcceptOrganization({ orgId }: { orgId: number }) {
  const router = useRouter();
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
        <p className="text-sm text-muted-foreground">
          {"Could not select your organization. Please try again."}
        </p>
      ) : (
        <>
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
          <p className="text-sm text-muted-foreground">{"Opening your dashboard…"}</p>
        </>
      )}
    </div>
  );
}
