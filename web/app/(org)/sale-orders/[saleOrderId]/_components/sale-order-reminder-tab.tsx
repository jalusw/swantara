"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import type { ReminderAction } from "./sale-order-detail-hooks";

type SaleOrderReminderTabProps = {
  reminders: ReminderAction[];
  onGenerate: () => void;
};

export function SaleOrderReminderTab({ reminders, onGenerate }: SaleOrderReminderTabProps) {
  const t = useTranslations("Sales");
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{t("tabReminder")}</CardTitle>
        <p className="text-sm text-muted-foreground">{t("reminderTabDescription")}</p>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <Button size="sm" variant="outline" onClick={onGenerate}>
          {t("generateReminder")}
        </Button>
        {reminders.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("noReminders")}</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-muted-foreground">
                  <th className="pb-2 text-left">{t("tableCustomer")}</th>
                  <th className="pb-2 text-left">{t("tableInvoice")}</th>
                  <th className="pb-2 text-left">{t("tableLevel")}</th>
                  <th className="pb-2 text-left">{t("daysOverdue")}</th>
                </tr>
              </thead>
              <tbody>
                {reminders.map((reminder) => (
                  <tr key={reminder.id} className="border-b">
                    <td className="py-2">{reminder.contactId}</td>
                    <td className="py-2">#{reminder.invoiceId}</td>
                    <td className="py-2">{reminder.levelId}</td>
                    <td className="py-2">—</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
