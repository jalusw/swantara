import { z } from "zod";
import type { RegistrationStep } from "./register-form.types";

export function useRegisterFormSchema() {
  return z
    .object({
      firstName: z.string().min(1, {
        message: "First Name is required",
      }),
      lastName: z.string(),
      email: z
        .email({
          message: "Please enter a valid email address.",
        })
        .min(1, {
          message: "Email is required.",
        }),
      password: z.string().min(8, {
        message: "Password must contain at least 8 characters",
      }),
      passwordConfirmation: z.string().min(8, {
        message: "Password confirmation must contain at least 8 characters",
      }),
    })
    .refine((data) => data.password === data.passwordConfirmation, {
      message: "Password confirmation doesn't match",
      path: ["passwordConfirmation"],
    });
}

export type RegisterFormSchema = z.infer<ReturnType<typeof useRegisterFormSchema>>;

export const registerFormDefaultValues: RegisterFormSchema = {
  firstName: "",
  lastName: "",
  email: "",
  password: "",
  passwordConfirmation: "",
};

export const stepFields: Record<RegistrationStep, (keyof RegisterFormSchema)[]> = {
  name: ["firstName", "lastName"],
  email: ["email"],
  password: ["password", "passwordConfirmation"],
};
