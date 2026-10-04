import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SalaryRulesSection } from "./_components/salary-rules-section";

export default async function OrgSalaryRulesPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Payroll");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("salaryRulesTitle")} description={t("salaryRulesSubtitle")} />
      <SalaryRulesSection orgId={id} />
    </div>
  );
}
