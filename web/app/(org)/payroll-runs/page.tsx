import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PayrollRunsSection } from "./_components/payroll-runs-section";

export default async function OrgPayrollRunsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Payroll");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <PayrollRunsSection orgId={id} />
    </div>
  );
}
