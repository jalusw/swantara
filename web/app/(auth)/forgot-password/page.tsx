import { getTranslations } from "next-intl/server";
import AuthShell from "../_components/auth-shell";
import ForgotPasswordForm from "./_components/forgot-password-form";

export default async function ForgotPasswordPage() {
  const t = await getTranslations("Auth");
  return (
    <AuthShell title={t("forgotPasswordTitle")} subtitle={t("forgotPasswordSubtitle")}>
      <ForgotPasswordForm />
    </AuthShell>
  );
}
