import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { TaxYearsSection } from "./_components/tax-years-section";

export default async function OrgTaxYearsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("taxYearsTitle")} description={t("taxYearsDescription")} />
      <TaxYearsSection orgId={id} />
    </div>
  );
}
