"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import AuthCard from "@/components/auth/AuthCard";
import AuthInput from "@/components/auth/AuthInput";
import { useSignUpMutation } from "@/lib/api/authApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { useAppSelector } from "@/lib/hooks";
import { markLocationPromptAfterAuth } from "@/lib/locationPrefs";

export default function SignUpPage() {
  const router = useRouter();
  const accessToken = useAppSelector((s) => s.auth.accessToken);
  const [fullName, setFullName] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [signUp, { isLoading }] = useSignUpMutation();

  useEffect(() => {
    if (accessToken) router.replace("/");
  }, [accessToken, router]);

  const canSubmit = fullName.trim() && username.trim() && password.length >= 8;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (password.length < 8) {
      toast.error("Invalid details", {
        description: "Password must be at least 8 characters",
      });
      return;
    }
    try {
      await signUp({
        full_name: fullName.trim(),
        username: username.trim(),
        password,
      }).unwrap();
      markLocationPromptAfterAuth();
      toast.success("Account created", { description: "Welcome to Yet Bota." });
      router.replace("/");
    } catch (err) {
      toast.error("Sign up failed", {
        description: getApiErrorMessage(err, "Could not create your account."),
      });
    }
  }

  return (
    <AuthCard title="Sign Up" backHref="/">
      <div className="mb-8 text-center">
        <h1 className="mb-2 text-2xl font-bold text-fg">Join Yet Bota.</h1>
        <p className="text-sm text-fg-faint">Start exploring your local community today.</p>
      </div>

      <form className="space-y-4" onSubmit={handleSubmit}>
        <AuthInput
          label="Full Name"
          type="text"
          placeholder="Enter your full name"
          autoComplete="name"
          value={fullName}
          onChange={(ev) => setFullName(ev.target.value)}
          disabled={isLoading}
        />

        <AuthInput
          label="Username"
          type="text"
          placeholder="Choose a username"
          autoComplete="username"
          value={username}
          onChange={(ev) => setUsername(ev.target.value)}
          disabled={isLoading}
        />

        <AuthInput
          label="Password"
          type="password"
          placeholder="Create a password"
          autoComplete="new-password"
          value={password}
          onChange={(ev) => setPassword(ev.target.value)}
          disabled={isLoading}
        />

        <Button
          type="submit"
          disabled={isLoading || !canSubmit}
          className="mt-2 h-12 w-full rounded-xl bg-brand text-base font-bold text-black hover:bg-brand-dark disabled:opacity-60"
        >
          {isLoading ? "Creating account…" : "Sign Up"}
        </Button>
      </form>

      <p className="mt-5 text-center text-xs leading-relaxed text-fg-faint">
        By signing up, you agree to our{" "}
        <Link
          href="/terms"
          className="text-fg-muted underline underline-offset-2 transition-colors hover:text-fg"
        >
          Terms of Service
        </Link>{" "}
        and{" "}
        <Link
          href="/privacy"
          className="text-fg-muted underline underline-offset-2 transition-colors hover:text-fg"
        >
          Privacy Policy
        </Link>
      </p>

      <p className="my-4 text-center text-xs text-fg-faint">
        Already have an account?{" "}
        <Link
          href="/signin"
          className="font-semibold text-brand transition-colors hover:text-brand-dark"
        >
          Sign In
        </Link>
      </p>
    </AuthCard>
  );
}
