import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { BankStatementsSection } from "./_components/bank-statements-section";

export default async function OrgBankStatementsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("bankStatementsTitle")} description={t("bankStatementsDescription")} />
      <BankStatementsSection orgId={id} />
    </div>
  );
}
