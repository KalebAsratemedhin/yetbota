"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import AuthInput from "@/components/auth/AuthInput";
import { ProfileSection } from "@/components/profile/ProfileSection";
import { useChangePasswordMutation } from "@/lib/api/authApi";
import { getApiErrorMessage } from "@/lib/auth/errors";

export default function ChangePasswordForm({ framed = true }: { framed?: boolean }) {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [changePassword, { isLoading }] = useChangePasswordMutation();

  const canSubmit = currentPassword.length > 0 && newPassword.length >= 8;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (newPassword.length < 8) {
      toast.error("Invalid password", {
        description: "New password must be at least 8 characters",
      });
      return;
    }
    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      }).unwrap();
      setCurrentPassword("");
      setNewPassword("");
      toast.success("Password updated", {
        description: "Use your new password next time you sign in.",
      });
    } catch (err) {
      toast.error("Could not change password", {
        description: getApiErrorMessage(err),
      });
    }
  }

  const body = (
    <form className="max-w-md space-y-4" onSubmit={handleSubmit}>
      <AuthInput
        label="Current password"
        type="password"
        autoComplete="current-password"
        value={currentPassword}
        onChange={(ev) => setCurrentPassword(ev.target.value)}
        disabled={isLoading}
      />
      <AuthInput
        label="New password"
        type="password"
        autoComplete="new-password"
        placeholder="At least 8 characters"
        value={newPassword}
        onChange={(ev) => setNewPassword(ev.target.value)}
        disabled={isLoading}
      />

      <Button
        type="submit"
        disabled={isLoading || !canSubmit}
        className="h-11 rounded-xl bg-brand px-6 font-bold text-black hover:bg-brand-dark disabled:opacity-60"
      >
        {isLoading ? "Updating…" : "Update password"}
      </Button>
    </form>
  );

  if (!framed) return body;

  return (
    <ProfileSection title="Security" description="Change the password for this account.">
      {body}
    </ProfileSection>
  );
}
