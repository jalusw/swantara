"use client";

import { TriangleAlertIcon } from "lucide-react";
import { isValidElement, type ReactNode } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "./alert-dialog";

type ConfirmVariant = "destructive" | "warning" | "neutral";

const variantTone: Record<ConfirmVariant, string> = {
  destructive: "bg-destructive/10 text-destructive",
  warning: "bg-warning/10 text-warning",
  neutral: "bg-primary/10 text-primary",
};

const confirmButtonVariant: Record<
  ConfirmVariant,
  "destructive" | "outline-destructive" | "default"
> = {
  destructive: "destructive",
  warning: "outline-destructive",
  neutral: "default",
};

export type ConfirmDialogProps = {
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: ConfirmVariant;
  loading?: boolean;
  disabled?: boolean;
  onConfirm: () => void;
  trigger?: ReactNode;
  className?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
};

/**
 * Helper over AlertDialog for destructive/irreversible confirmations.
 * Pass `trigger` as a button element for uncontrolled mode, or pass
 * `open`/`onOpenChange` to control the dialog from outside.
 */
export function ConfirmDialog({
  title,
  description,
  confirmLabel,
  cancelLabel,
  variant = "destructive",
  loading = false,
  disabled = false,
  onConfirm,
  trigger,
  className,
  open,
  onOpenChange,
}: ConfirmDialogProps) {
  const resolvedConfirmLabel = confirmLabel ?? "Confirm";
  const resolvedCancelLabel = cancelLabel ?? "Cancel";
  const idle = loading || disabled;
  const isControlled = open !== undefined;
  return (
    <AlertDialog data-slot="confirm-dialog" open={open} onOpenChange={onOpenChange}>
      {!isControlled && trigger && isValidElement(trigger) ? (
        <AlertDialogTrigger render={trigger} />
      ) : null}
      <AlertDialogContent size="sm" className={className}>
        <AlertDialogHeader>
          <AlertDialogMedia className={variantTone[variant]}>
            <TriangleAlertIcon aria-hidden />
          </AlertDialogMedia>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          {description ? <AlertDialogDescription>{description}</AlertDialogDescription> : null}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={idle}>{resolvedCancelLabel}</AlertDialogCancel>
          <AlertDialogAction
            variant={confirmButtonVariant[variant]}
            disabled={idle}
            aria-busy={loading || undefined}
            onClick={onConfirm}
          >
            {resolvedConfirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
