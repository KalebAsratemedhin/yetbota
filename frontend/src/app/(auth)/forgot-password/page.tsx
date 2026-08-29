"use client";

import Link from "next/link";
import AuthCard from "@/components/auth/AuthCard";

export default function ForgotPasswordPage() {
  return (
    <AuthCard title="Forgot Password" backHref="/signin">
      <div className="mb-8 text-center">
        <h1 className="mb-2 text-2xl font-bold text-fg">Reset unavailable</h1>
        <p className="text-sm text-fg-faint">
          Password reset is not available yet. If you are signed in, change your password in
          Settings.
        </p>
      </div>
      <Link
        href="/signin"
        className="inline-flex h-12 w-full items-center justify-center rounded-xl bg-brand text-base font-bold text-black transition-colors hover:bg-brand-dark"
      >
        Back to sign in
      </Link>
    </AuthCard>
  );
}
