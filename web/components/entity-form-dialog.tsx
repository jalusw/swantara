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

export type EntityFormDialogProps<TFieldValues extends FieldValues = FieldValues> = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  form: UseFormReturn<TFieldValues>;
  onSubmit: (data: TFieldValues) => void | Promise<void>;
  isPending?: boolean;
  className?: string;
  children: React.ReactNode;
};

export function EntityFormDialog<TFieldValues extends FieldValues = FieldValues>({
  open,
  onOpenChange,
  title,
  description,
  form,
  onSubmit,
  isPending = false,
  className,
  children,
}: EntityFormDialogProps<TFieldValues>) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-slot="entity-form-dialog" className={className ?? "sm:max-w-2xl"}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <Form form={form} onSubmit={onSubmit}>
          {children}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={isPending}
              onClick={() => onOpenChange(false)}
            >
              {"Cancel"}
            </Button>
            <Button type="submit" disabled={isPending} aria-busy={isPending || undefined}>
              {"Save"}
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
