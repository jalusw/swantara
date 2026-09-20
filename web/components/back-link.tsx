"use client";

import { ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";

export function BackLink({ href, children }: { href: string; children: React.ReactNode }) {
  const router = useRouter();
  return (
    <button
      data-slot="back-link"
      type="button"
      onClick={() => router.push(href)}
      className="inline-flex min-h-11 min-w-11 items-center gap-1.5 px-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
    >
      <ArrowLeft className="size-4" />
      {children}
    </button>
  );
}
