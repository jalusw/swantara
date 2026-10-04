import Link from "next/link";
import { getTranslations } from "next-intl/server";
import AuthShell from "../_components/auth-shell";
import RegisterForm from "./_components/register-form";

export default async function RegisterPage() {
  const t = await getTranslations("Auth");
  return (
    <AuthShell title={t("registerTitle")} subtitle={t("registerSubtitle")}>
      <RegisterForm />
      <p className="text-sm text-muted-foreground">
        {t("haveAccount")}{" "}
        <Link
          className="inline-flex min-h-11 items-center rounded-md px-1 font-medium text-primary underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          href="/login"
        >
          {t("loginButton")}
        </Link>
      </p>
    </AuthShell>
  );
}
