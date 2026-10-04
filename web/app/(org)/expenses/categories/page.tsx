import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ExpenseCategoriesSection } from "./_components/expense-categories-section";

export default async function ExpenseCategoriesPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Expenses",
  );
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("categoriesTitle")} description={t("categoriesDescription")} />
      <ExpenseCategoriesSection orgId={id} />
    </div>
  );
}
