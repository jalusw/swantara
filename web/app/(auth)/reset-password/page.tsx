import { getTranslations } from "next-intl/server";
import AuthShell from "../_components/auth-shell";
import ResetPasswordForm from "./_components/reset-password-form";

export type ResetPasswordPageProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function ResetPasswordPage({ searchParams }: ResetPasswordPageProps) {
  const { token } = await searchParams;
  const t = await getTranslations("Auth");

  return (
    <AuthShell title={t("resetPasswordTitle")} subtitle={t("resetPasswordSubtitle")}>
      <ResetPasswordForm token={token} />
    </AuthShell>
  );
}
