"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import AuthCard from "@/components/auth/AuthCard";
import AuthInput from "@/components/auth/AuthInput";
import { useSignInMutation } from "@/lib/api/authApi";
import { getApiErrorMessage, safeNextPath } from "@/lib/auth/errors";
import { useAppSelector } from "@/lib/hooks";
import { markLocationPromptAfterAuth } from "@/lib/locationPrefs";

export default function SignInPage() {
  const router = useRouter();
  const accessToken = useAppSelector((s) => s.auth.accessToken);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [signIn, { isLoading }] = useSignInMutation();

  useEffect(() => {
    if (!accessToken) return;
    const next =
      typeof window !== "undefined"
        ? safeNextPath(new URLSearchParams(window.location.search).get("next"))
        : "/";
    router.replace(next);
  }, [accessToken, router]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await signIn({ username: username.trim(), password }).unwrap();
      markLocationPromptAfterAuth();
      toast.success("Signed in", { description: "Welcome back." });
      const next = safeNextPath(new URLSearchParams(window.location.search).get("next"));
      router.replace(next);
    } catch (err) {
      toast.error("Sign in failed", {
        description: getApiErrorMessage(err, "Check your username and password."),
      });
    }
  }

  return (
    <AuthCard title="Sign In" backHref="/">
      <div className="mb-8 text-center">
        <h1 className="mb-2 text-2xl font-bold text-fg">Welcome Back.</h1>
        <p className="text-sm text-fg-faint">Stay connected with your local community.</p>
      </div>

      <form className="space-y-4" onSubmit={handleSubmit}>
        <AuthInput
          label="Username"
          type="text"
          placeholder="Enter your username"
          autoComplete="username"
          value={username}
          onChange={(ev) => setUsername(ev.target.value)}
          disabled={isLoading}
        />

        <AuthInput
          label="Password"
          type="password"
          placeholder="········"
          autoComplete="current-password"
          value={password}
          onChange={(ev) => setPassword(ev.target.value)}
          disabled={isLoading}
          rightLabel={
            <Link
              href="/forgot-password"
              className="text-xs font-medium text-brand transition-colors hover:text-brand-dark"
            >
              Forgot Password?
            </Link>
          }
        />

        <Button
          type="submit"
          disabled={isLoading || !username.trim() || !password}
          className="mt-2 h-12 w-full rounded-xl bg-brand text-base font-bold text-black hover:bg-brand-dark disabled:opacity-60"
        >
          {isLoading ? "Signing in…" : "Sign In"}
        </Button>
      </form>

      <p className="mt-8 text-center text-xs text-fg-faint">
        Don&apos;t have an account?{" "}
        <Link
          href="/signup"
          className="font-semibold text-brand transition-colors hover:text-brand-dark"
        >
          Sign Up
        </Link>
      </p>
    </AuthCard>
  );
}
