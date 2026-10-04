"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

export function QualitySubNav() {
  const t = useTranslations("Quality");
  return (
    <OrgSubNav
      label={t("title")}
      tabs={[
        {
          key: "quality",
          label: t("checksTitle"),
          href: "/quality",
        },
        { key: "qualityPoints", href: "/quality/points" },
        { key: "qualityAlerts", href: "/quality/alerts" },
      ]}
    />
  );
}
