"use client";

import { Bell, MapPin, Settings } from "lucide-react";
import type { UserInfo } from "@/lib/auth/types";

function initialsFrom(user: UserInfo): string {
  const fromName = user.full_name
    .split(" ")
    .filter(Boolean)
    .map((n) => n[0])
    .join("");
  return (fromName || user.username || "YB").slice(0, 2).toUpperCase();
}

export default function ProfileHeader({
  user,
  onOpenSettings,
}: {
  user: UserInfo;
  onOpenSettings?: () => void;
}) {
  const location = user.address.trim() || "Add location in Settings";

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
          <div className="flex h-24 w-24 items-center justify-center overflow-hidden rounded-3xl border-4 border-bg bg-[#e8b89a] shadow-[0px_25px_50px_-12px_rgba(0,0,0,0.35)] sm:h-32 sm:w-32 lg:h-40 lg:w-40">
            <span className="text-xl font-bold text-fg/60 sm:text-2xl lg:text-3xl">
              {initialsFrom(user)}
            </span>
          </div>
        </div>

        <div className="w-full min-w-0 flex-1 pb-1 sm:pb-2">
          <div className="mb-1.5 flex flex-wrap items-center gap-2 sm:mb-2 sm:gap-3">
            <h1 className="max-w-full truncate text-xl leading-tight font-bold text-fg sm:text-2xl lg:text-[30px]">
              {user.full_name.trim() || user.username}
            </h1>
            <span className="rounded-xl border border-brand/30 bg-brand/20 px-2.5 py-1 text-[10px] font-bold tracking-[0.12em] text-brand uppercase sm:px-3 sm:text-[12px]">
              Explorer
            </span>
          </div>
          <div className="flex flex-wrap items-center gap-4 sm:gap-6">
            <span className="text-sm sm:text-base">
              <strong className="text-fg">—</strong>{" "}
              <span className="text-fg-muted">Followers</span>
            </span>
            <span className="text-sm sm:text-base">
              <strong className="text-fg">—</strong>{" "}
              <span className="text-fg-muted">Following</span>
            </span>
            <span className="flex max-w-[18rem] items-center gap-1.5 sm:max-w-[24rem]">
              <MapPin className="h-3.5 w-3.5 shrink-0 text-fg-faint" />
              <span className="truncate text-sm text-fg-muted">{location}</span>
            </span>
          </div>
          {user.bio.trim() ? (
            <p className="mt-3 max-w-xl text-sm leading-relaxed text-fg-faint">{user.bio}</p>
          ) : null}
        </div>

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
      </div>
    </div>
  );
}
