import { cn } from "@/lib/utils";

export type FieldProps = {
  children?: React.ReactNode;
  className?: string;
};

function Field({ children, className }: FieldProps) {
  return (
    <div data-slot="field" className={cn("flex flex-col gap-y-1.5", className)}>
      {children}
    </div>
  );
}

export { Field };
