"use client";

import { useRef } from "react";
import { Bell, Camera, MapPin, Settings } from "lucide-react";
import { toast } from "sonner";
import UserAvatar from "@/components/shared/UserAvatar";
import { useUploadAvatarMutation } from "@/lib/api/authApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { UserInfo } from "@/lib/auth/types";
import { cn } from "@/lib/utils";

const ACCEPTED = "image/jpeg,image/png,image/webp,image/gif";
const MAX_BYTES = 2 * 1024 * 1024;

function formatCount(n: number | undefined): string {
  if (n === undefined) return "—";
  return n.toLocaleString();
}

function validateAvatarFile(file: File): string | null {
  const okType = ["image/jpeg", "image/png", "image/webp", "image/gif"].includes(
    file.type
  );
  if (!okType) return "Use a JPEG, PNG, WebP, or GIF image.";
  if (file.size > MAX_BYTES) return "Image must be 2MB or smaller.";
  return null;
}

export default function ProfileHeader({
  user,
  mode,
  followerCount,
  followingCount,
  onOpenFollowers,
  onOpenFollowing,
  onOpenSettings,
  isFollowing,
  followBusy,
  onToggleFollow,
}: {
  user: UserInfo;
  mode: "own" | "public";
  followerCount?: number;
  followingCount?: number;
  onOpenFollowers?: () => void;
  onOpenFollowing?: () => void;
  onOpenSettings?: () => void;
  isFollowing?: boolean;
  followBusy?: boolean;
  onToggleFollow?: () => void;
}) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploadAvatar, { isLoading: uploading }] = useUploadAvatarMutation();

  const location =
    user.address.trim() || (mode === "own" ? "Add location in Settings" : "No location");
  const showFollow = mode === "public" && Boolean(onToggleFollow);

  async function onFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;
    const problem = validateAvatarFile(file);
    if (problem) {
      toast.error("Invalid image", { description: problem });
      return;
    }
    try {
      await uploadAvatar(file).unwrap();
      toast.success("Avatar updated");
    } catch (err) {
      toast.error("Could not upload avatar", {
        description: getApiErrorMessage(err),
      });
    }
  }

  return (
    <div className="relative shrink-0">
      <div className="relative h-44 w-full overflow-hidden rounded-2xl bg-linear-to-br from-[#123b2a] via-[#0f2f23] to-[#0a0a0a] sm:h-56 lg:h-64">
        <div
          className="absolute inset-0"
          style={{
            backgroundImage: `
              radial-gradient(ellipse at 50% 35%, rgba(255,255,255,0.10), transparent 55%),
              linear-gradient(90deg, rgba(255,255,255,0.08), transparent 55%),
              linear-gradient(180deg, rgba(0,0,0,0.05), rgba(0,0,0,0.65))
            `,
          }}
        />
        <div className="absolute inset-0 bg-linear-to-t from-[#0a0a0a]/90 via-[#0a0a0a]/35 to-transparent" />
      </div>

      <div className="relative z-10 -mt-14 flex flex-col items-start gap-4 px-4 sm:-mt-18 sm:px-6 lg:-mt-20 lg:flex-row lg:items-end lg:gap-6">
        <div className="relative shrink-0">
          <UserAvatar
            user={user}
            size="lg"
            className="border-4 border-bg shadow-[0px_25px_50px_-12px_rgba(0,0,0,0.35)]"
          />
          {mode === "own" ? (
            <>
              <input
                ref={fileRef}
                type="file"
                accept={ACCEPTED}
                className="sr-only"
                onChange={onFileChange}
              />
              <button
                type="button"
                disabled={uploading}
                onClick={() => fileRef.current?.click()}
                className="absolute right-1 bottom-1 inline-flex h-9 w-9 items-center justify-center rounded-xl bg-brand text-black shadow-md transition-colors hover:bg-brand-dark disabled:opacity-60 sm:h-10 sm:w-10"
                aria-label={uploading ? "Uploading avatar" : "Upload avatar"}
                title="Upload avatar"
              >
                <Camera className="h-4 w-4 sm:h-5 sm:w-5" />
              </button>
            </>
          ) : null}
        </div>

        <div className="w-full min-w-0 flex-1 pb-1 sm:pb-2">
          <div className="mb-1.5 flex flex-wrap items-center gap-2 sm:mb-2 sm:gap-3">
            <h1 className="max-w-full truncate text-xl leading-tight font-bold text-fg sm:text-2xl lg:text-[30px]">
              {user.full_name.trim() || user.username}
            </h1>
            <span className="text-sm text-fg-faint">@{user.username}</span>
            <span className="rounded-xl border border-brand/30 bg-brand/20 px-2.5 py-1 text-[10px] font-bold tracking-[0.12em] text-brand uppercase sm:px-3 sm:text-[12px]">
              {user.reputation_level || "Explorer"}
            </span>
          </div>
          <div className="flex flex-wrap items-center gap-4 sm:gap-6">
            <button
              type="button"
              onClick={onOpenFollowers}
              disabled={!onOpenFollowers}
              className="text-left text-sm transition-opacity hover:opacity-80 disabled:cursor-default sm:text-base"
            >
              <strong className="text-fg">{formatCount(followerCount)}</strong>{" "}
              <span className="text-fg-muted">Followers</span>
            </button>
            <button
              type="button"
              onClick={onOpenFollowing}
              disabled={!onOpenFollowing}
              className="text-left text-sm transition-opacity hover:opacity-80 disabled:cursor-default sm:text-base"
            >
              <strong className="text-fg">{formatCount(followingCount)}</strong>{" "}
              <span className="text-fg-muted">Following</span>
            </button>
            <span className="flex max-w-[18rem] items-center gap-1.5 sm:max-w-[24rem]">
              <MapPin className="h-3.5 w-3.5 shrink-0 text-fg-faint" />
              <span className="truncate text-sm text-fg-muted">{location}</span>
            </span>
          </div>
          {user.bio.trim() ? (
            <p className="mt-3 max-w-xl text-sm leading-relaxed text-fg-faint">{user.bio}</p>
          ) : null}
        </div>

        {mode === "own" ? (
          <div className="flex items-center gap-2 self-end lg:mb-2 lg:shrink-0 lg:self-auto">
            <button
              type="button"
              disabled
              title="Coming soon"
              className="inline-flex h-11 w-11 items-center justify-center rounded-xl border border-border-subtle bg-surface text-fg-muted opacity-60"
              aria-label="Notifications (coming soon)"
            >
              <Bell className="h-5 w-5" />
            </button>
            <button
              type="button"
              onClick={onOpenSettings}
              className="inline-flex h-11 w-11 items-center justify-center rounded-xl border border-border-subtle bg-surface text-fg-muted transition-colors hover:bg-surface-2 hover:text-fg"
              aria-label="Settings"
            >
              <Settings className="h-5 w-5" />
            </button>
          </div>
        ) : null}
      </div>

      {showFollow ? (
        <div className="mt-4 flex justify-end px-4 sm:px-6">
          <button
            type="button"
            onClick={onToggleFollow}
            disabled={followBusy}
            className={cn(
              "inline-flex h-10 min-w-[7.5rem] items-center justify-center rounded-full px-6 text-sm font-bold transition-colors disabled:opacity-60",
              isFollowing
                ? "border border-border-subtle bg-surface text-fg hover:bg-surface-2"
                : "bg-brand text-black hover:bg-brand-dark"
            )}
          >
            {followBusy ? "…" : isFollowing ? "Following" : "Follow"}
          </button>
        </div>
      ) : null}
    </div>
  );
}
