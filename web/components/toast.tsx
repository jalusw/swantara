"use client";

import { Toaster as SonnerToaster } from "sonner";

type ToasterProps = React.ComponentProps<typeof SonnerToaster> & {
  theme?: "light" | "dark";
};

function Toaster({ theme = "light", ...props }: ToasterProps) {
  const sonnerTheme = theme === "dark" ? "dark" : "light";

  return (
    <SonnerToaster
      data-slot="toaster"
      theme={sonnerTheme}
      className="toaster group"
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-popover group-[.toaster]:text-popover-foreground group-[.toaster]:border-border group-[.toaster]:shadow-md group-[.toaster]:rounded-lg group-[.toaster]:ring-1 group-[.toaster]:ring-foreground/10",
          description: "group-[.toast]:text-muted-foreground text-sm",
          actionButton:
            "group-[.toast]:bg-primary group-[.toast]:text-primary-foreground group-[.toast]:rounded-md group-[.toast]:text-sm group-[.toast]: group-[.toast]:px-3 group-[.toast]:py-1 group-[.toast]:h-auto",
          cancelButton:
            "group-[.toast]:bg-transparent group-[.toast]:text-muted-foreground group-[.toast]:rounded-md group-[.toast]:text-sm group-[.toast]:px-3 group-[.toast]:py-1 group-[.toast]:h-auto group-[.toast]:border group-[.toast]:border-border",
          closeButton:
            "group-[.toast]:text-muted-foreground group-[.toast]:opacity-50 group-[.toast]:hover:opacity-100 group-[.toast]:transition-opacity",
          icon: "group-[.toast]:size-4",
          error: "group-[.toaster]:border-destructive/30 group-[.toaster]:ring-destructive/20",
          success: "group-[.toaster]:border-success/30 group-[.toaster]:ring-success/20",
          warning: "group-[.toaster]:border-warning/30 group-[.toaster]:ring-warning/20",
          info: "group-[.toaster]:border-info/30 group-[.toaster]:ring-info/20",
        },
      }}
      {...props}
    />
  );
}

export { Toaster };
export { toast } from "sonner";
