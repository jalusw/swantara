"use client";

import { ErrorState } from "@/components/error-state";

export default function ReportsError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return <ErrorState message={error.message} digest={error.digest} onReset={reset} />;
}
