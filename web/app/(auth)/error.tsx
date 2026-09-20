"use client";

import { ErrorState } from "@/components/error-state";

export default function AuthError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="grid min-h-dvh place-items-center px-4 py-16">
      <ErrorState message={error.message} digest={error.digest} onReset={reset} />
    </main>
  );
}
