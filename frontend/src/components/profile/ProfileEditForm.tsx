"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import AuthInput from "@/components/auth/AuthInput";
import { ProfileSection } from "@/components/profile/ProfileSection";
import { useUpdateProfileMutation } from "@/lib/api/authApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { UserInfo } from "@/lib/auth/types";

export default function ProfileEditForm({
  user,
  framed = true,
}: {
  user: UserInfo;
  framed?: boolean;
}) {
  const [fullName, setFullName] = useState(user.full_name);
  const [phone, setPhone] = useState(user.phone);
  const [address, setAddress] = useState(user.address);
  const [bio, setBio] = useState(user.bio);
  const [updateProfile, { isLoading }] = useUpdateProfileMutation();

  useEffect(() => {
    setFullName(user.full_name);
    setPhone(user.phone);
    setAddress(user.address);
    setBio(user.bio);
  }, [user]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await updateProfile({
        full_name: fullName.trim(),
        phone: phone.trim(),
        address: address.trim(),
        bio: bio.trim(),
      }).unwrap();
      toast.success("Profile saved", { description: "Your details were updated." });
    } catch (err) {
      toast.error("Could not update profile", {
        description: getApiErrorMessage(err),
      });
    }
  }

  const body = (
    <form className="space-y-4" onSubmit={handleSubmit}>
      <AuthInput
        label="Full name"
        value={fullName}
        onChange={(ev) => setFullName(ev.target.value)}
        autoComplete="name"
        disabled={isLoading}
      />
      <div className="grid gap-4 sm:grid-cols-2">
        <AuthInput
          label="Phone"
          value={phone}
          onChange={(ev) => setPhone(ev.target.value)}
          autoComplete="tel"
          placeholder="Optional"
          disabled={isLoading}
        />
        <AuthInput
          label="Address"
          value={address}
          onChange={(ev) => setAddress(ev.target.value)}
          autoComplete="street-address"
          placeholder="Optional"
          disabled={isLoading}
        />
      </div>
      <div className="w-full">
        <label htmlFor="profile-bio" className="mb-1.5 block text-sm font-medium text-fg-muted">
          Bio
        </label>
        <textarea
          id="profile-bio"
          rows={4}
          value={bio}
          onChange={(ev) => setBio(ev.target.value)}
          disabled={isLoading}
          placeholder="A short intro for your community profile"
          className="w-full resize-y rounded-xl border border-border-subtle bg-surface-2 px-4 py-3 text-sm text-fg outline-none transition-all placeholder:text-gray-600 focus:border-brand/50 focus:ring-1 focus:ring-brand/20 disabled:opacity-60"
        />
      </div>

      <Button
        type="submit"
        disabled={isLoading}
        className="h-11 rounded-xl bg-brand px-6 font-bold text-black hover:bg-brand-dark disabled:opacity-60"
      >
        {isLoading ? "Saving…" : "Save changes"}
      </Button>
    </form>
  );

  if (!framed) return body;

  return (
    <ProfileSection
      title="Profile details"
      description="Name, contact, and bio shown on your public-facing profile."
    >
      {body}
    </ProfileSection>
  );
}
