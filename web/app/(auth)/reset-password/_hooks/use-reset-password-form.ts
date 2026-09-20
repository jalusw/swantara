import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { axiosInstance } from "@/lib/client/axios";
import { isServerError } from "@/lib/client/error";

const REDIRECT_DELAY = 1200;

export function useResetPasswordFormSchema() {
  return z
    .object({
      password: z.string().min(8, { message: "Password must contain at least 8 characters" }),
      passwordConfirmation: z
        .string()
        .min(8, { message: "Password confirmation must contain at least 8 characters" }),
    })
    .refine((value) => value.password === value.passwordConfirmation, {
      path: ["passwordConfirmation"],
      message: "Password confirmation doesn't match",
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
  const resetPasswordFormSchema = useResetPasswordFormSchema();

  const form = useForm<ResetPasswordFormSchema>({
    resolver: zodResolver(resetPasswordFormSchema),
    defaultValues: resetPasswordFormDefaultValues,
  });

  const handleSubmit = form.handleSubmit((data: ResetPasswordFormSchema) => {
    if (!token) {
      toast.error("We couldn't reset your password. Please try again later.");
      return;
    }
    startTransition(async () => {
      try {
        await axiosInstance.post("/auth/password-reset", {
          token,
          password: data.password,
        });
        setIsSuccess(true);
        toast.success("Password updated!");
        setTimeout(() => router.push("/login"), REDIRECT_DELAY);
      } catch (error: unknown) {
        if (isServerError(error)) {
          toast.error("Something went wrong. Please try again later.");
          return;
        }
        toast.error("We couldn't reset your password. Please try again later.");
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
