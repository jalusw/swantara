"use client";

import { ArrowLeft, Headset, LifeBuoy, RefreshCcw } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "./button";

export type ErrorStateLabels = {
  title?: string;
  description?: string;
  retry?: string;
  contactSupport?: string;
  backToHome?: string;
  errorReference?: string;
};

export type ErrorStateProps = {
  message?: string;
  digest?: string;
  onReset: () => void;
  className?: string;
  labels?: ErrorStateLabels;
  supportHref?: string;
  homeHref?: string;
};

export function ErrorState({
  message,
  digest,
  onReset,
  className,
  labels,
  supportHref = "/support",
  homeHref = "/login",
}: ErrorStateProps) {
  const allLabels = {
    title: "Terjadi kesalahan",
    description:
      "Terjadi kesalahan tak terduga. Silakan coba lagi atau hubungi dukungan jika masalah berlanjut.",
    retry: "Coba lagi",
    contactSupport: "Hubungi dukungan",
    backToHome: "Kembali ke beranda",
    errorReference: digest ? `Referensi kesalahan: ${digest}` : undefined,
    ...labels,
  };

  return (
    <div
      data-slot="error-state"
      className={cn("flex w-full flex-col items-center text-center", className)}
    >
      <div className="mt-10 flex flex-col items-center gap-4 sm:mt-12">
        <LifeBuoy className="size-12 text-primary" aria-hidden />
        <h1 className="max-w-2xl font-heading text-4xl leading-tight tracking-tight text-balance sm:text-5xl">
          {allLabels.title}
        </h1>
        <p className="max-w-xl text-base leading-relaxed text-muted-foreground text-pretty sm:text-lg">
          {allLabels.description}
        </p>
      </div>

      <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row">
        <Button size="lg" onClick={onReset}>
          <RefreshCcw aria-hidden />
          {allLabels.retry}
        </Button>
        <Button size="lg" variant="outline" asChild>
          <a href={supportHref}>
            <Headset aria-hidden />
            {allLabels.contactSupport}
          </a>
        </Button>
      </div>

      <a
        href={homeHref}
        className="mt-8 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden />
        {allLabels.backToHome}
      </a>

      {message || digest ? (
        <div className="mt-10 w-full max-w-md">
          {message ? (
            <code className="block w-full truncate rounded-md bg-muted px-3 py-2 text-left text-xs text-muted-foreground">
              {message}
            </code>
          ) : null}
          {digest && allLabels.errorReference ? (
            <p className="mt-3 text-xs text-muted-foreground">
              {allLabels.errorReference.replace("{digest}", digest)}
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
