import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";

import { createServerSwantaraService } from "@/lib/server/api-client";
import { SwantaraError } from "@/lib/services/swantara/errors";
import AuthShell from "../(auth)/_components/auth-shell";
import { AcceptOrganization } from "./_components/accept-organization";
import OnboardingForm from "./_components/onboarding-form";

export const dynamic = "force-dynamic";

export default async function OnboardingPage() {
  const t = await getTranslations("Onboarding");
  const service = await createServerSwantaraService();
  let organizations: Awaited<ReturnType<typeof service.me.organizations>>["organizations"];
  try {
    ({ organizations } = await service.me.organizations());
  } catch (error) {
    if (error instanceof SwantaraError && (error.status === 401 || error.status >= 500)) {
      redirect("/api/v1/auth/session?next=/onboarding");
    }
    throw error;
  }
  if (organizations.length > 0) {
    return (
      <AuthShell title={t("openingDashboard")} subtitle={t("choosingOrg")}>
        <AcceptOrganization orgId={organizations[0]!.id} />
      </AuthShell>
    );
  }

  return (
    <AuthShell title={t("setupCompany")} subtitle={t("setupSubtitle")}>
      <OnboardingForm />
    </AuthShell>
  );
}
