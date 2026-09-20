import { cva, type VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";
import { cn } from "@/lib/utils";

const fieldFeedbackVariants = cva("inline", {
  variants: {
    intent: {
      success: "text-success",
      warning: "text-warning",
      danger: "text-destructive",
    },
    size: {
      xs: "text-xs",
      sm: "text-sm",
      md: "text-base",
      lg: "text-lg",
    },
  },
});

export type FieldFeedbackProps = {
  children?: React.ReactNode;
  className?: string;
  visible?: boolean;
  id?: string;
} & ComponentProps<"p"> &
  VariantProps<typeof fieldFeedbackVariants>;

function FieldFeedback({
  children,
  className,
  intent,
  visible,
  size = "sm",
  ...props
}: FieldFeedbackProps) {
  if (!children || !visible) {
    return null;
  }
  return (
    <p
      data-slot="field-feedback"
      className={cn(fieldFeedbackVariants({ intent, size }), className)}
      {...props}
    >
      {children}
    </p>
  );
}

export { FieldFeedback };
