import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ExpensesSection } from "./_components/expenses-section";

export default async function ExpensesPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Expenses",
  );
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("expensesTitle")} description={t("expensesDescription")} />
      <ExpensesSection orgId={id} />
    </div>
  );
}
