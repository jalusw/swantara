import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { Breadcrumb, type BreadcrumbItem } from "./breadcrumb";
import { PageHeader } from "./page-header";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./tabs";

export type RecordTab = {
  id: string;
  label: string;
  content: ReactNode;
};

export type RecordLayoutProps = {
  breadcrumbItems: BreadcrumbItem[];
  title: string;
  description?: string;
  status?: ReactNode;
  actions?: ReactNode;
  tabs?: RecordTab[];
  activeTab?: string;
  onTabChange?: (id: string) => void;
  defaultTab?: string;
  children?: ReactNode;
  className?: string;
};

export function RecordLayout({
  breadcrumbItems,
  title,
  description,
  status,
  actions,
  tabs,
  activeTab,
  onTabChange,
  defaultTab,
  children,
  className,
}: RecordLayoutProps) {
  return (
    <div data-slot="record-layout" className={cn("space-y-4", className)}>
      {breadcrumbItems.length > 0 ? <Breadcrumb items={breadcrumbItems} /> : null}

      <PageHeader
        title={title}
        description={description}
        actions={
          <>
            {status}
            {actions}
          </>
        }
      />

      {tabs ? (
        <Tabs
          value={activeTab}
          defaultValue={defaultTab ?? tabs[0]?.id}
          onValueChange={onTabChange}
          data-slot="record-tabs"
        >
          <TabsList variant="line">
            {tabs.map((tab) => (
              <TabsTrigger key={tab.id} value={tab.id}>
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
          {tabs.map((tab) => (
            <TabsContent key={tab.id} value={tab.id}>
              {tab.content}
            </TabsContent>
          ))}
        </Tabs>
      ) : (
        children
      )}
    </div>
  );
}
