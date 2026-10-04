"use client";

import { useTranslations } from "next-intl";
import { PageHeader } from "@/components/page-header";
import { ProfileForm } from "./_components/profile-form";

export default function OrgProfilePage() {
  const t = useTranslations("Profile");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <ProfileForm />
    </div>
  );
}
