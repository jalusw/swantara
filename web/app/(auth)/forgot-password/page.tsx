import AuthShell from "../_components/auth-shell";
import ForgotPasswordForm from "./_components/forgot-password-form";

export default async function ForgotPasswordPage() {
  return (
    <AuthShell
      title={"Reset your password"}
      subtitle={
        "Enter your phone number or email and we'll send you instructions to reset your password."
      }
    >
      <ForgotPasswordForm />
    </AuthShell>
  );
}
