import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CustomersAcquisition } from "./_components/customers-acquisition-section";
import { CustomersStats } from "./_components/customers-stats-section";
import { CustomersTable } from "./_components/customers-table-section";

export default async function OrgCustomersPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Customers"}
        description={"Manage your customer accounts, contacts, and outstanding balances."}
      />

      <section aria-label={"Customer overview"}>
        <CustomersStats />
      </section>

      <section aria-label={"All customers"}>
        <Card>
          <CardHeader className="border-b">
            <CardTitle>{"All customers"}</CardTitle>
            <CardDescription>
              {"A list of everyone who does business with your organization."}
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <CustomersTable orgId={id} />
          </CardContent>
        </Card>
      </section>

      <section aria-label={"Customer acquisition"}>
        <CustomersAcquisition />
      </section>
    </div>
  );
}
