"use client";

import { PageHeader } from "@/components/page-header";
import { ProfileForm } from "./_components/profile-form";

export default function OrgProfilePage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Profile"} description={"Manage your personal information"} />
      <ProfileForm />
    </div>
  );
}
