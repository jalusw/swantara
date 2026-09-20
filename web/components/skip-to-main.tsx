import { cn } from "@/lib/utils";

export type SkipToMainProps = React.ComponentProps<"a"> & {
  label?: string;
  href?: string;
};

function SkipToMain({
  label = "Skip to main content",
  href = "#main",
  className,
  ...props
}: SkipToMainProps) {
  return (
    <a
      href={href}
      data-slot="skip-to-main"
      className={cn(
        "fixed top-3 left-1/2 z-skip flex min-h-11 min-w-11 -translate-x-1/2 items-center justify-center rounded-md bg-primary px-4 py-3 text-sm text-primary-foreground shadow-lg focus:-translate-x-1/2 focus:ring-2 focus:ring-ring focus:outline-hidden focus:outline-none [clip-path:inset(0_50%_50%_50%)] focus:[clip-path:inset(0)]",
        className,
      )}
      {...props}
    >
      {label}
    </a>
  );
}

export { SkipToMain };
