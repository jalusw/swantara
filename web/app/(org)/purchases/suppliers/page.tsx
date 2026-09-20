import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { PageHeader } from "@/components/page-header";

import { SuppliersStats } from "./_components/suppliers-stats-section";
import { SuppliersTable } from "./_components/suppliers-table-section";
import { SUPPLIERS_SPEND_TREND } from "./_utils";

export default async function OrgSuppliersPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Suppliers"}
        description={"Keep track of the suppliers and contacts your organization buys from."}
      />

      <SuppliersStats />

      <Card>
        <CardHeader>
          <CardTitle>{"Spend by month"}</CardTitle>
          <CardDescription>{"Total procurement spend over the last six months."}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={SUPPLIERS_SPEND_TREND}
            ariaLabel={"Spend by month"}
            valueFormatter={(value) => `$${value}k`}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="border-b">
          <CardTitle>{"All suppliers"}</CardTitle>
          <CardDescription>
            {"A list of everyone you purchase goods and services from."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <SuppliersTable />
        </CardContent>
      </Card>
    </div>
  );
}
