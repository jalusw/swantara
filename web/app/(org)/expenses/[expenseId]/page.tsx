import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ExpenseDetail } from "./_components/expense-detail-section";

export default async function ExpenseDetailPage({
  params,
}: {
  params: Promise<{ expenseId: string }>;
}) {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Expenses",
  );
  const { expenseId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/expenses"}>{t("backToExpenses")}</BackLink>
      <ExpenseDetail orgId={id} expenseId={expenseId} />
    </div>
  );
}
