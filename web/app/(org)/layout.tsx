import { type DehydratedState, HydrationBoundary } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { Suspense } from "react";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { prefetchMeData, prefetchPermissionsData } from "@/lib/server/prefetch";
import { OrgActiveProvider } from "@/providers/org-active-provider";
import { OrgShell } from "./_components/org-shell";

export default async function OrganizationLayout({ children }: { children: ReactNode }) {
  const organizationId = await requireActiveOrgId();
  const [meState, permissionsState] = await Promise.allSettled([
    prefetchMeData(),
    prefetchPermissionsData(organizationId),
  ]);
  const queries = [
    ...(meState.status === "fulfilled" ? meState.value.queries : []),
    ...(permissionsState.status === "fulfilled" ? permissionsState.value.queries : []),
  ];
  const dehydratedState: DehydratedState = { queries, mutations: [] };

  return (
    <HydrationBoundary state={dehydratedState}>
      <OrgActiveProvider orgId={organizationId}>
        <Suspense fallback={null}>
          <OrgShell>{children}</OrgShell>
        </Suspense>
      </OrgActiveProvider>
    </HydrationBoundary>
  );
}
