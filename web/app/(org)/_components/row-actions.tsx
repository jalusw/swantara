"use client";

import { PencilIcon, Trash2Icon } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/button";
import { ConfirmDialog } from "@/components/confirm-dialog";

export type RowActionsProps = {
  editLabel: string;
  deleteLabel: string;
  confirmTitle: string;
  confirmDescription: string;
  confirmLabel?: string;
  onEdit: () => void;
  onDelete?: () => void;
};

export function RowActions({
  editLabel,
  deleteLabel,
  confirmTitle,
  confirmDescription,
  confirmLabel,
  onEdit,
  onDelete,
}: RowActionsProps) {
  const t = useTranslations("Common");
  return (
    <div className="flex items-center justify-end gap-1">
      <Button variant="ghost" size="icon-sm" onClick={onEdit} aria-label={editLabel}>
        <PencilIcon />
      </Button>
      {onDelete ? (
        <ConfirmDialog
          title={confirmTitle}
          description={confirmDescription}
          confirmLabel={confirmLabel ?? t("delete")}
          onConfirm={onDelete}
          trigger={
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={deleteLabel}
              className="text-destructive"
            >
              <Trash2Icon />
            </Button>
          }
        />
      ) : null}
    </div>
  );
}
