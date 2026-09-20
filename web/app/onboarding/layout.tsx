import { DigitalFingerPrintProvider } from "@/providers/digital-fingerprint";

export type OnboardingLayoutProps = {
  children: React.ReactNode;
};

export default async function OnboardingLayout({ children }: OnboardingLayoutProps) {
  return (
    <DigitalFingerPrintProvider>
      <main className="min-h-screen" id="main">
        {children}
      </main>
    </DigitalFingerPrintProvider>
  );
}
