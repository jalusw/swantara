"use client";

import type { FieldValues, UseFormReturn } from "react-hook-form";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form } from "@/components/form";
import { cn } from "@/lib/utils";

export type EntityFormDialogProps<TFieldValues extends FieldValues = FieldValues> = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  badge?: React.ReactNode;
  form: UseFormReturn<TFieldValues>;
  onSubmit: (data: TFieldValues) => void | Promise<void>;
  isPending?: boolean;
  submitLabel?: string;
  cancelLabel?: string;
  submitVariant?: "default" | "destructive" | "secondary" | "outline" | "ghost" | "link";
  footerHint?: React.ReactNode;
  className?: string;
  children: React.ReactNode;
};

export function EntityFormDialog<TFieldValues extends FieldValues = FieldValues>({
  open,
  onOpenChange,
  title,
  description,
  badge,
  form,
  onSubmit,
  isPending = false,
  submitLabel = "Save",
  cancelLabel = "Cancel",
  submitVariant = "default",
  footerHint,
  className,
  children,
}: EntityFormDialogProps<TFieldValues>) {
  const isSubmitting = form.formState.isSubmitting || isPending;
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        data-slot="entity-form-dialog"
        className={cn("flex max-h-[85vh] flex-col overflow-hidden", className ?? "sm:max-w-2xl")}
      >
        <DialogHeader className="shrink-0">
          <div className="flex items-center gap-2 pr-8">
            <DialogTitle className="text-base font-semibold">{title}</DialogTitle>
            {badge}
          </div>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <Form form={form} onSubmit={onSubmit} className="flex min-h-0 flex-1 flex-col gap-0">
          <div className="-mx-4 min-h-0 flex-1 overflow-y-auto px-4 py-1">{children}</div>
          <DialogFooter className="items-center sm:items-center">
            {footerHint ? (
              <p className="mr-auto hidden text-xs text-muted-foreground sm:block">{footerHint}</p>
            ) : null}
            <Button
              type="button"
              variant="outline"
              disabled={isSubmitting}
              onClick={() => onOpenChange(false)}
            >
              {cancelLabel}
            </Button>
            <Button
              type="submit"
              variant={submitVariant}
              disabled={isSubmitting}
              aria-busy={isSubmitting || undefined}
            >
              {isSubmitting ? "Saving…" : submitLabel}
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
