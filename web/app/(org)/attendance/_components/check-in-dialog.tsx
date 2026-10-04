"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useId, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import type { Employee } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

export function CheckInDialog({
  open,
  onOpenChange,
  orgId,
  employees,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  employees: Employee[];
  onSave: () => void;
}) {
  const employeeId = useId();
  const timeId = useId();
  const [isPending, setIsPending] = useState(false);
  const t = useTranslations("Attendance");
  const tCommon = useTranslations("Common");

  const schema = z.object({
    employeeId: z.string().min(1, t("validationEmployeeRequired")),
    checkIn: z.string().min(1, t("validationCheckInRequired")),
  });

  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      employeeId: "",
      checkIn: toLocalDatetime(new Date()),
    },
  });

  function toLocalDatetime(date: Date): string {
    const pad = (n: number) => String(n).padStart(2, "0");
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
  }

  function handleSubmit(values: Values) {
    setIsPending(true);
    void getSwantaraService()
      .attendances.checkIn(Number(orgId), {
        employeeId: Number(values.employeeId),
        checkIn: new Date(values.checkIn).toISOString(),
      })
      .then(() => {
        form.reset({ employeeId: "", checkIn: toLocalDatetime(new Date()) });
        onSave();
      })
      .catch(() => toast.error(t("saveFailed")))
      .finally(() => setIsPending(false));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("checkInTitle")}</DialogTitle>
          <DialogDescription>{t("checkInDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="employeeId" label={t("fieldEmployee")} className="flex flex-col gap-2">
              {({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={employeeId} aria-label={t("fieldEmployee")}>
                    <SelectValue placeholder={t("selectEmployee")} />
                  </SelectTrigger>
                  <SelectContent>
                    {employees.map((emp) => (
                      <SelectItem key={emp.id} value={String(emp.id)}>
                        {emp.employeeNumber}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="checkIn" label={t("fieldCheckInTime")} className="flex flex-col gap-2">
              {({ field }) => <Input id={timeId} type="datetime-local" {...field} />}
            </FormField>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <Button type="submit" disabled={isPending}>
              {t("confirmCheckIn")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
