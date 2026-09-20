import type { Metadata, Viewport } from "next";
import { Inter_Tight } from "next/font/google";
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

export const metadata: Metadata = {
  title: "Swantara",
  description: "Simplified and streamline your business workflow.",
  formatDetection: {
    telephone: false,
  },
};

export const viewport: Viewport = {
  themeColor: "#0f172a",
  width: "device-width",
  initialScale: 1,
};

export type RootLayoutProps = {
  children: React.ReactNode;
};

export default function RootLayout({ children }: RootLayoutProps) {
  return (
    <html
      lang="en"
      className={cn("antialiased", interTight.variable, "font-sans")}
      suppressHydrationWarning
      data-scroll-behavior="smooth"
    >
      <body>
        <SkipToMain label="Skip to main content" />
        <ProviderContainer>{children}</ProviderContainer>
        <ReportWebVitals />
      </body>
    </html>
  );
}
