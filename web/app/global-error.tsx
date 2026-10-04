"use client";

import { useEffect, useState } from "react";

import "@/styles/globals.css";

import { ErrorState } from "@/components/error-state";
import { type AppTheme, isAppTheme } from "@/components/theme";
import { cn } from "@/lib/utils";

function resolveStoredTheme(): AppTheme {
  try {
    const stored = localStorage.getItem("theme");
    if (stored && isAppTheme(stored)) {
      return stored;
    }
    if (stored === "system") {
      return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
    }
  } catch {}
  return "dark";
}

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const [theme, setTheme] = useState<AppTheme>("dark");
  const colorScheme = theme === "dark" ? "dark" : "light";
  useEffect(() => {
    setTheme(resolveStoredTheme());
  }, []);

  return (
    <html lang="id" className={cn("antialiased", theme)} style={{ colorScheme }}>
      <body>
        <main className="grid min-h-dvh place-items-center px-4 py-16">
          <ErrorState
            message={error.message}
            digest={error.digest}
            onReset={reset}
            className="animate-fade-up"
            labels={{
              title: "Terjadi kesalahan",
              description:
                "Terjadi kesalahan tak terduga. Silakan coba lagi atau hubungi dukungan jika masalah berlanjut.",
              retry: "Coba lagi",
              contactSupport: "Hubungi dukungan",
              backToHome: "Kembali ke beranda",
            }}
          />
        </main>
      </body>
    </html>
  );
}
