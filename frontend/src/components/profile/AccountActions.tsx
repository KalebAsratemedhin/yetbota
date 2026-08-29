"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import AuthInput from "@/components/auth/AuthInput";
import { ProfileSection } from "@/components/profile/ProfileSection";
import { useDeleteAccountMutation, useLogoutMutation } from "@/lib/api/authApi";
import { getApiErrorMessage } from "@/lib/auth/errors";

export function SessionPanel() {
  const router = useRouter();
  const [logout, { isLoading: isLoggingOut }] = useLogoutMutation();

  async function handleLogout() {
    try {
      await logout().unwrap();
      toast.success("Signed out", { description: "See you next time." });
    } catch {
      toast.success("Signed out", { description: "Local session cleared." });
    }
    router.replace("/");
  }

  return (
    <div className="space-y-5">
      <p className="text-sm text-fg-muted">
        Signing out revokes this device’s refresh token. Other devices stay signed in until they
        expire or you change your password elsewhere.
      </p>
      <Button
        type="button"
        onClick={handleLogout}
        disabled={isLoggingOut}
        className="h-11 rounded-xl border border-border-subtle bg-bg px-6 font-semibold text-fg hover:border-brand/40 hover:bg-overlay disabled:opacity-60"
      >
        {isLoggingOut ? "Signing out…" : "Log out"}
      </Button>
    </div>
  );
}

export function DangerZonePanel({ username }: { username: string }) {
  const router = useRouter();
  const [confirmText, setConfirmText] = useState("");
  const [deleteAccount, { isLoading: isDeleting }] = useDeleteAccountMutation();
  const canDelete = confirmText === username;

  async function handleDelete(e: React.FormEvent) {
    e.preventDefault();
    if (!canDelete) return;
    try {
      await deleteAccount().unwrap();
      toast.success("Account deleted", {
        description: "Your identity account has been removed.",
      });
      router.replace("/");
    } catch (err) {
      toast.error("Could not delete account", {
        description: getApiErrorMessage(err),
      });
    }
  }

  return (
    <form className="space-y-4" onSubmit={handleDelete}>
      <div className="rounded-2xl border border-red-500/25 bg-red-500/5 px-4 py-3 text-sm text-fg-muted">
        Type <span className="font-semibold text-fg">{username}</span> to confirm permanent deletion.
      </div>
      <AuthInput
        label="Confirm username"
        value={confirmText}
        onChange={(ev) => setConfirmText(ev.target.value)}
        autoComplete="off"
        disabled={isDeleting}
      />
      <Button
        type="submit"
        disabled={isDeleting || !canDelete}
        className="h-11 rounded-xl border border-red-500/40 bg-red-500/10 px-6 font-bold text-red-400 hover:bg-red-500/20 disabled:opacity-60"
      >
        {isDeleting ? "Deleting…" : "Delete account"}
      </Button>
    </form>
  );
}

export default function AccountActions({ username }: { username: string }) {
  return (
    <div className="space-y-6">
      <ProfileSection title="Session" description="End this device session.">
        <SessionPanel />
      </ProfileSection>
      <ProfileSection
        title="Danger zone"
        description="Permanently delete your account. This cannot be undone."
        className="border-red-500/20"
      >
        <DangerZonePanel username={username} />
      </ProfileSection>
    </div>
  );
}
