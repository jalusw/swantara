import { Suspense } from "react";
import { DigitalFingerPrintProvider } from "@/providers/digital-fingerprint";

export type AuthLayoutProps = {
  children: React.ReactNode;
};

export default async function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <DigitalFingerPrintProvider>
      <Suspense fallback={null}>
        <main className="min-h-screen" id="main">
          {children}
        </main>
      </Suspense>
    </DigitalFingerPrintProvider>
  );
}
