import { z } from "zod";
import type { RegistrationStep } from "./register-form.types";

export type RegisterFormMessages = {
  firstNameRequired: string;
  invalidEmail: string;
  emailRequired: string;
  passwordMinLength: string;
  confirmPasswordMinLength: string;
  passwordMismatch: string;
};

const registerFormDefaultMessages: RegisterFormMessages = {
  firstNameRequired: "Nama depan wajib diisi",
  invalidEmail: "Masukkan alamat email yang valid.",
  emailRequired: "Email wajib diisi.",
  passwordMinLength: "Kata sandi minimal 8 karakter",
  confirmPasswordMinLength: "Konfirmasi kata sandi minimal 8 karakter",
  passwordMismatch: "Konfirmasi kata sandi tidak cocok",
};

export function useRegisterFormSchema(
  messages: RegisterFormMessages = registerFormDefaultMessages,
) {
  return z
    .object({
      firstName: z.string().min(1, {
        message: messages.firstNameRequired,
      }),
      lastName: z.string(),
      email: z
        .email({
          message: messages.invalidEmail,
        })
        .min(1, {
          message: messages.emailRequired,
        }),
      password: z.string().min(8, {
        message: messages.passwordMinLength,
      }),
      passwordConfirmation: z.string().min(8, {
        message: messages.confirmPasswordMinLength,
      }),
    })
    .refine((data) => data.password === data.passwordConfirmation, {
      message: messages.passwordMismatch,
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
