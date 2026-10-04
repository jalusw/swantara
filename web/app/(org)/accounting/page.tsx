import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { AccountingStats } from "./_components/accounting-stats";

import { AccountsList } from "./_components/accounts-list";
import { CashFlowSummary } from "./_components/cash-flow-summary";
import { PerformanceChart } from "./_components/performance-chart";
import { RecentEntries } from "./_components/recent-entries";
import { getMockAccounts, getMockEntries } from "./_utils";

export default async function OrgAccountingPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const accounts = getMockAccounts();
  const entries = getMockEntries();

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("overviewTitle")} description={t("overviewDescription")} />

      <AccountingStats />

      <section className="grid gap-4 sm:gap-6 lg:grid-cols-3">
        <PerformanceChart />
        <CashFlowSummary />
      </section>

      <AccountsList accounts={accounts} />

      <RecentEntries entries={entries} />
    </div>
  );
}
