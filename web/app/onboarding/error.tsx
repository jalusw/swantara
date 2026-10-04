"use client";

import { useTranslations } from "next-intl";
import { ErrorState } from "@/components/error-state";

export default function OnboardingError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const t = useTranslations("Error");
  return (
    <main className="grid min-h-dvh place-items-center px-4 py-16">
      <ErrorState
        message={error.message}
        digest={error.digest}
        onReset={reset}
        labels={{
          title: t("title"),
          description: t("description"),
          retry: t("retry"),
          contactSupport: t("contactSupport"),
          backToHome: t("backToHome"),
          errorReference: error.digest ? t("reference", { digest: error.digest }) : undefined,
        }}
      />
    </main>
  );
}
