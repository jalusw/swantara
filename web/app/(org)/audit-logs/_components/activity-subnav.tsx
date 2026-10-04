"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

export function ActivitySubNav() {
  const t = useTranslations("AuditLogs");
  return (
    <OrgSubNav
      label={t("layoutTitle")}
      tabs={[
        { key: "activity", label: t("title"), href: "/audit-logs" },
        {
          key: "integrationEvents",
          href: "/audit-logs/integration-events",
        },
      ]}
    />
  );
}
