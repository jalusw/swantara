"use client";

import { useTranslations } from "next-intl";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/tabs";
import { CashSection } from "./cash-section";
import { DashboardOverview } from "./dashboard-overview-section";
import { FinanceOverview } from "./finance-overview";
import { InventoryHealthSection, InventoryOverview } from "./inventory-section";
import { PayrollSection, ProjectSection, SubscriptionSection } from "./operations-section";
import { ProcurementKpiCard } from "./procurement-card";
import { ManufacturingSection } from "./production-section";
import { PipelineSection, SalesSection } from "./sales-section";

const TABS = [
  "overview",
  "finance",
  "cash",
  "sales",
  "inventory",
  "procurement",
  "manufacturing",
  "operations",
] as const;

type TabId = (typeof TABS)[number];

export function DashboardTabs({ orgId }: { orgId: string }) {
  const t = useTranslations("Dashboard");
  const tabLabel = (tab: TabId) => (t as unknown as (k: string) => string)(`tab_${tab}`);
  return (
    <Tabs defaultValue="overview" className="gap-4 sm:gap-6">
      <TabsList variant="line">
        {TABS.map((tab) => (
          <TabsTrigger key={tab} value={tab}>
            {tabLabel(tab)}
          </TabsTrigger>
        ))}
      </TabsList>

      <TabsContent value="overview">
        <DashboardOverview orgId={orgId} />
      </TabsContent>

      <TabsContent value="finance">
        <FinanceTab />
      </TabsContent>

      <TabsContent value="cash">
        <CashTab />
      </TabsContent>

      <TabsContent value="sales">
        <SalesTab />
      </TabsContent>

      <TabsContent value="inventory">
        <InventoryTab />
      </TabsContent>

      <TabsContent value="procurement">
        <ProcurementTab />
      </TabsContent>

      <TabsContent value="manufacturing">
        <ManufacturingTab />
      </TabsContent>

      <TabsContent value="operations">
        <OperationsTab />
      </TabsContent>
    </Tabs>
  );
}

function TabSectionHead({ tab }: { tab: Exclude<TabId, "overview"> }) {
  const t = useTranslations("Dashboard");
  const dyn = t as unknown as (k: string) => string;
  return (
    <div className="space-y-1">
      <h2 className="text-sm font-medium uppercase tracking-widest text-muted-foreground">
        {dyn(`heading_${tab}`)}
      </h2>
      <p className="text-sm text-muted-foreground">{dyn(`headingDesc_${tab}`)}</p>
    </div>
  );
}

function FinanceTab() {
  return (
    <div className="flex flex-col gap-6">
      <TabSectionHead tab="finance" />
      <FinanceOverview />
    </div>
  );
}

function CashTab() {
  return (
    <div className="flex flex-col gap-6">
      <TabSectionHead tab="cash" />
      <CashSection />
    </div>
  );
}

function SalesTab() {
  return (
    <div className="flex flex-col gap-6">
      <TabSectionHead tab="sales" />
      <section className="grid gap-4 lg:grid-cols-12">
        <div className="lg:col-span-7">
          <SalesSection />
        </div>
        <div className="lg:col-span-5">
          <PipelineSection />
        </div>
      </section>
    </div>
  );
}

function InventoryTab() {
  return (
    <div className="flex flex-col gap-6">
      <TabSectionHead tab="inventory" />
      <section className="grid gap-4 lg:grid-cols-12">
        <div className="lg:col-span-7">
          <InventoryOverview />
        </div>
        <div className="lg:col-span-5">
          <InventoryHealthSection />
        </div>
      </section>
    </div>
  );
}

function ProcurementTab() {
  return (
    <div className="flex flex-col gap-6">
      <TabSectionHead tab="procurement" />
      <ProcurementKpiCard />
    </div>
  );
}

function ManufacturingTab() {
  return (
    <div className="flex flex-col gap-6">
      <TabSectionHead tab="manufacturing" />
      <ManufacturingSection />
    </div>
  );
}

function OperationsTab() {
  return (
    <div className="flex flex-col gap-6">
      <TabSectionHead tab="operations" />
      <section className="flex flex-col gap-8">
        <PayrollSection />
        <ProjectSection />
        <SubscriptionSection />
      </section>
    </div>
  );
}
