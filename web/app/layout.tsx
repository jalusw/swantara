import type { Metadata, Viewport } from "next";
import { Inter_Tight } from "next/font/google";
import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages, getTranslations } from "next-intl/server";
import "@/styles/globals.css";

import { ReportWebVitals } from "@/app/_components/report-web-vitals";
import { SkipToMain } from "@/components/skip-to-main";
import { cn } from "@/lib/utils/style";
import { ProviderContainer } from "@/providers/container";

const interTight = Inter_Tight({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("Metadata");
  return {
    title: t("title"),
    description: t("description"),
    formatDetection: {
      telephone: false,
    },
  };
}

export const viewport: Viewport = {
  themeColor: "#0f172a",
  width: "device-width",
  initialScale: 1,
};

export type RootLayoutProps = {
  children: React.ReactNode;
};

export default async function RootLayout({ children }: RootLayoutProps) {
  const locale = await getLocale();
  const messages = await getMessages();
  const tCommon = await getTranslations("Common");
  return (
    <html
      lang={locale}
      className={cn("antialiased", interTight.variable, "font-sans")}
      suppressHydrationWarning
      data-scroll-behavior="smooth"
    >
      <body>
        <NextIntlClientProvider messages={messages}>
          <SkipToMain label={tCommon("skipToMain")} />
          <ProviderContainer>{children}</ProviderContainer>
          <ReportWebVitals />
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
