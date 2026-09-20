"use client";

import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import type { ReminderAction } from "./sale-order-detail-hooks";

type SaleOrderReminderTabProps = {
  reminders: ReminderAction[];
  onGenerate: () => void;
};

export function SaleOrderReminderTab({ reminders, onGenerate }: SaleOrderReminderTabProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{"Reminder"}</CardTitle>
        <p className="text-sm text-muted-foreground">
          {"Overdue invoices and reminder actions by days overdue."}
        </p>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <Button size="sm" variant="outline" onClick={onGenerate}>
          {"Generate reminder"}
        </Button>
        {reminders.length === 0 ? (
          <p className="text-sm text-muted-foreground">{"No reminder actions."}</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-muted-foreground">
                  <th className="pb-2 text-left">{"Customer"}</th>
                  <th className="pb-2 text-left">{"Invoice"}</th>
                  <th className="pb-2 text-left">{"Level"}</th>
                  <th className="pb-2 text-left">{"Days overdue"}</th>
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
