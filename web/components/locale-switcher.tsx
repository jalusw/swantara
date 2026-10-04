"use client";

import { Check, Globe } from "lucide-react";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { useTransition } from "react";
import { Button } from "@/components/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/dropdown-menu";
import { setAppLocale } from "@/lib/i18n/actions";
import { type Locale, localeNames, locales } from "@/lib/i18n/config";
import { cn } from "@/lib/utils/style";

export type LocaleSwitcherProps = {
  className?: string;
  variant?: "ghost" | "outline" | "default";
};

export function LocaleSwitcher({ className, variant = "ghost" }: LocaleSwitcherProps) {
  const locale = useLocale() as Locale;
  const t = useTranslations("Locale");
  const router = useRouter();
  const [isPending, startTransition] = useTransition();

  function handleChange(next: Locale) {
    if (next === locale) return;
    startTransition(async () => {
      await setAppLocale(next);
      router.refresh();
    });
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant={variant}
          size="icon"
          className={cn(className)}
          aria-label={t("label")}
          disabled={isPending}
        >
          <Globe aria-hidden />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {locales.map((code) => (
          <DropdownMenuItem
            key={code}
            onSelect={() => handleChange(code)}
            className={cn(code === locale && "font-medium")}
          >
            <span className="flex w-full items-center justify-between gap-2">
              {localeNames[code]}
              {code === locale ? <Check aria-hidden className="size-4" /> : null}
            </span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
