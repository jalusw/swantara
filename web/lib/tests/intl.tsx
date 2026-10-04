import { type RenderOptions, render } from "@testing-library/react";
import { NextIntlClientProvider } from "next-intl";
import type { ReactNode } from "react";
import enMessages from "../../messages/en.json";
import idMessages from "../../messages/id.json";
import { defaultLocale } from "../i18n/config";

type TestLocale = "id" | "en";

export function renderWithIntl(
  ui: ReactNode,
  locale: TestLocale = defaultLocale,
  options?: Omit<RenderOptions, "wrapper">,
) {
  const messages = locale === "en" ? enMessages : idMessages;
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <NextIntlClientProvider locale={locale} messages={messages} timeZone="Asia/Jakarta">
        {children}
      </NextIntlClientProvider>
    );
  }
  return render(ui, { wrapper: Wrapper, ...options });
}

export const testMessages = {
  id: idMessages,
  en: enMessages,
};
