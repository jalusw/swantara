"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/tabs";
import { CurrenciesSection } from "./currencies-section";
import { DimensionsSection } from "./dimensions-section";
import { FxRatesSection } from "./fx-rates-section";
import { PaymentTermsSection } from "./payment-terms-section";
import { UnitsSection } from "./units-section";

const tabs = [
  { id: "currencies", component: CurrenciesSection },
  { id: "fxRates", component: FxRatesSection },
  { id: "units", component: UnitsSection },
  { id: "dimensions", component: DimensionsSection },
  { id: "paymentTerms", component: PaymentTermsSection },
] as const;

export function ReferenceDataTabs({ orgId }: { orgId: string }) {
  return (
    <Tabs defaultValue="currencies">
      <TabsList variant="line">
        {tabs.map((tab) => (
          <TabsTrigger key={tab.id} value={tab.id}>
            {String(tab.id)}
          </TabsTrigger>
        ))}
      </TabsList>
      {tabs.map((tab) => {
        const Section = tab.component;
        return (
          <TabsContent key={tab.id} value={tab.id} className="pt-4">
            <Section orgId={orgId} />
          </TabsContent>
        );
      })}
    </Tabs>
  );
}
