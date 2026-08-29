"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import AuthInput from "@/components/auth/AuthInput";
import { ProfileSection } from "@/components/profile/ProfileSection";
import { useUpdateProfileMutation, useDeleteAvatarMutation } from "@/lib/api/authApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { UserInfo } from "@/lib/auth/types";
import UserAvatar from "@/components/shared/UserAvatar";
import {
  isValidOptionalPhone,
  normalizePhone,
  PHONE_INVALID_MESSAGE,
} from "@/lib/validation/phone";

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
  const [phoneError, setPhoneError] = useState<string | undefined>();
  const [updateProfile, { isLoading }] = useUpdateProfileMutation();
  const [deleteAvatar, { isLoading: deletingAvatar }] = useDeleteAvatarMutation();

  function validatePhoneField(value: string): boolean {
    if (isValidOptionalPhone(value)) {
      setPhoneError(undefined);
      return true;
    }
    setPhoneError(PHONE_INVALID_MESSAGE);
    return false;
  }

  async function handleRemoveAvatar() {
    try {
      await deleteAvatar().unwrap();
      toast.success("Avatar removed");
    } catch (err) {
      toast.error("Could not remove avatar", {
        description: getApiErrorMessage(err),
      });
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!validatePhoneField(phone)) {
      toast.error("Invalid phone number", { description: PHONE_INVALID_MESSAGE });
      return;
    }
    try {
      await updateProfile({
        full_name: fullName.trim(),
        phone: normalizePhone(phone),
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

  const hasAvatar = Boolean(user.avatar_url?.trim());
  const busy = isLoading || deletingAvatar;

  const body = (
    <form className="space-y-4" onSubmit={handleSubmit}>
      <div className="flex items-center gap-4 rounded-xl border border-border-subtle bg-surface-2/50 p-3">
        <UserAvatar user={user} size="md" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-fg">Profile photo</p>
          <p className="text-xs text-fg-faint">
            Change it from the camera on your profile. JPEG, PNG, WebP, or GIF · max 2MB.
          </p>
          {hasAvatar ? (
            <button
              type="button"
              onClick={() => void handleRemoveAvatar()}
              disabled={busy}
              className="mt-2 text-xs font-semibold text-red-400 hover:underline disabled:opacity-60"
            >
              {deletingAvatar ? "Removing…" : "Remove photo"}
            </button>
          ) : null}
        </div>
      </div>

      <AuthInput
        label="Full name"
        value={fullName}
        onChange={(ev) => setFullName(ev.target.value)}
        autoComplete="name"
        disabled={busy}
      />
      <div className="grid gap-4 sm:grid-cols-2">
        <AuthInput
          label="Phone"
          type="tel"
          inputMode="tel"
          value={phone}
          onChange={(ev) => {
            setPhone(ev.target.value);
            if (phoneError) validatePhoneField(ev.target.value);
          }}
          onBlur={() => validatePhoneField(phone)}
          autoComplete="tel"
          placeholder="+251911234567"
          disabled={busy}
          error={phoneError}
        />
        <AuthInput
          label="Address"
          value={address}
          onChange={(ev) => setAddress(ev.target.value)}
          autoComplete="street-address"
          placeholder="Optional"
          disabled={busy}
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
          disabled={busy}
          placeholder="A short intro for your community profile"
          className="w-full resize-y rounded-xl border border-border-subtle bg-surface-2 px-4 py-3 text-sm text-fg outline-none transition-all placeholder:text-gray-600 focus:border-brand/50 focus:ring-1 focus:ring-brand/20 disabled:opacity-60"
        />
      </div>

      <Button
        type="submit"
        disabled={busy || Boolean(phoneError)}
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
