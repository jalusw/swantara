import { cookies } from "next/headers";
import { getRequestConfig } from "next-intl/server";
import { defaultLocale, isValidLocale, localeCookieName } from "@/lib/i18n/config";
import enMessages from "../messages/en.json";
import idMessages from "../messages/id.json";

const allMessages = {
  id: idMessages,
  en: enMessages,
} as const;

export default getRequestConfig(async () => {
  const store = await cookies();
  const raw = store.get(localeCookieName)?.value;
  const locale = isValidLocale(raw) ? raw : defaultLocale;

  return {
    locale,
    timeZone: "Asia/Jakarta",
    messages: allMessages[locale],
    // Graceful fallback for missing dynamic keys (e.g. sessionState_open):
    // return last segment humanized instead of throwing, so incremental
    // translation migration never crashes the UI.
    getMessageFallback({ namespace, key }: { namespace?: string; key: string }) {
      const last = key.split(".").pop() ?? key;
      // Strip common dynamic prefixes like sessionState_open -> Open
      const cleaned = last
        .replace(
          /^(sessionState|orderState|paymentMethod|giftCardState|invoiceState|paymentState|entryState|statementState|transferState|shipmentState|countState|expenseState|rmaState|rmaType|contractState|alertState|col|table|field)_?/,
          "",
        )
        .replace(/_/g, " ")
        .trim();
      return namespace ? cleaned || last : last;
    },
    onError(error) {
      // Missing messages are expected during incremental migration; log in dev only
      if (process.env.NODE_ENV === "development") {
        // eslint-disable-next-line no-console
        console.warn("[next-intl]", error);
      }
    },
  };
});
