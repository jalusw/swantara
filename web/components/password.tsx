"use client";

import { Eye, EyeOff } from "lucide-react";
import { type ComponentProps, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Input } from "./input";

export type PasswordProps = Omit<ComponentProps<"input">, "type">;

function Password({ className, disabled, ...props }: PasswordProps) {
  const [visible, setVisible] = useState(false);

  return (
    <div data-slot="password" className="relative flex w-full items-center">
      <Input
        type={visible ? "text" : "password"}
        disabled={disabled}
        placeholder="********"
        className={cn("pr-12", className)}
        {...props}
      />
      <Button
        type="button"
        variant="ghost"
        size="icon"
        disabled={disabled}
        aria-label={visible ? "Hide password" : "Show password"}
        aria-pressed={visible}
        onClick={() => setVisible((v) => !v)}
        className="absolute right-0 top-1/2 -translate-y-1/2 active:not-aria-[haspopup]:translate-y-[calc(-50%+1px)]"
      >
        {visible ? <EyeOff aria-hidden /> : <Eye aria-hidden />}
      </Button>
    </div>
  );
}

export { Password };
