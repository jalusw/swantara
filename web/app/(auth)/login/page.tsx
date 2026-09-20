import type { Metadata } from "next";
import Link from "next/link";
import AuthShell from "../_components/auth-shell";
import LoginForm from "./_components/login-form";

export const metadata: Metadata = {
  robots: {
    index: false,
    follow: false,
  },
};

export default async function LoginPage() {
  return (
    <AuthShell title="Welcome back" subtitle="Enter your credentials to get started.">
      <LoginForm />
      <p className="text-sm text-muted-foreground">
        {"New here?"}{" "}
        <Link
          className="inline-flex min-h-11 items-center rounded-md px-1 font-medium text-primary underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          href="/register"
        >
          {"Register a new account"}
        </Link>
      </p>
    </AuthShell>
  );
}
