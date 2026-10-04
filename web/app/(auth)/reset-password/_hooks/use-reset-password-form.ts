import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { axiosInstance } from "@/lib/client/axios";
import { isServerError } from "@/lib/client/error";

const REDIRECT_DELAY = 1200;

export type ResetPasswordFormMessages = {
  passwordMinLength: string;
  confirmPasswordMinLength: string;
  passwordMismatch: string;
};

const resetPasswordFormDefaultMessages: ResetPasswordFormMessages = {
  passwordMinLength: "Password must contain at least 8 characters",
  confirmPasswordMinLength: "Password confirmation must contain at least 8 characters",
  passwordMismatch: "Password confirmation doesn't match",
};

export function useResetPasswordFormSchema(
  messages: ResetPasswordFormMessages = resetPasswordFormDefaultMessages,
) {
  return z
    .object({
      password: z.string().min(8, { message: messages.passwordMinLength }),
      passwordConfirmation: z.string().min(8, { message: messages.confirmPasswordMinLength }),
    })
    .refine((value) => value.password === value.passwordConfirmation, {
      path: ["passwordConfirmation"],
      message: messages.passwordMismatch,
    });
}

export type ResetPasswordFormSchema = z.infer<ReturnType<typeof useResetPasswordFormSchema>>;

const resetPasswordFormDefaultValues: ResetPasswordFormSchema = {
  password: "",
  passwordConfirmation: "",
};

export function useResetPasswordForm({ token }: { token?: string }) {
  const [isPending, startTransition] = useTransition();
  const [isSuccess, setIsSuccess] = useState(false);
  const router = useRouter();
  const t = useTranslations("Auth");
  const tCommon = useTranslations("Common");
  const tx = t as unknown as (key: string) => string;
  const txCommon = tCommon as unknown as (key: string) => string;
  const resetPasswordFormSchema = useResetPasswordFormSchema({
    passwordMinLength: tx("passwordMinLength"),
    confirmPasswordMinLength: tx("confirmPasswordMinLength"),
    passwordMismatch: tx("passwordMismatch"),
  });

  const form = useForm<ResetPasswordFormSchema>({
    resolver: zodResolver(resetPasswordFormSchema),
    defaultValues: resetPasswordFormDefaultValues,
  });

  const handleSubmit = form.handleSubmit((data: ResetPasswordFormSchema) => {
    if (!token) {
      toast.error(tx("resetFailed"));
      return;
    }
    startTransition(async () => {
      try {
        await axiosInstance.post("/auth/password-reset", {
          token,
          password: data.password,
        });
        setIsSuccess(true);
        toast.success(tx("passwordUpdatedTitle"));
        setTimeout(() => router.push("/login"), REDIRECT_DELAY);
      } catch (error: unknown) {
        if (isServerError(error)) {
          toast.error(txCommon("serverError"));
          return;
        }
        toast.error(tx("resetFailed"));
      }
    });
  });

  return {
    form,
    handleSubmit,
    isPending,
    isSuccess,
  };
}
