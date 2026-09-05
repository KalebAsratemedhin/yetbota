"use client";

import { useState } from "react";
import Link from "next/link";
import {
  AlertTriangle,
  ArrowLeft,
  KeyRound,
  LogOut,
  MapPin,
  Palette,
  UserRound,
} from "lucide-react";
import { DangerZonePanel, SessionPanel } from "@/components/profile/AccountActions";
import AppearanceSettings from "@/components/profile/AppearanceSettings";
import ChangePasswordForm from "@/components/profile/ChangePasswordForm";
import LocationSettings from "@/components/profile/LocationSettings";
import ProfileEditForm from "@/components/profile/ProfileEditForm";
import UserAvatar from "@/components/shared/UserAvatar";
import type { UserInfo } from "@/lib/auth/types";
import { cn } from "@/lib/utils";

type SettingsSection =
  | "appearance"
  | "location"
  | "profile"
  | "security"
  | "session"
  | "danger";

const NAV: {
  id: SettingsSection;
  label: string;
  hint: string;
  icon: React.ComponentType<{ className?: string }>;
  tone?: "danger";
}[] = [
  { id: "appearance", label: "Appearance", hint: "Theme on this device", icon: Palette },
  { id: "location", label: "Location", hint: "Nearby feed ranking", icon: MapPin },
  { id: "profile", label: "Profile", hint: "Name, contact, bio", icon: UserRound },
  { id: "security", label: "Security", hint: "Change password", icon: KeyRound },
  { id: "session", label: "Session", hint: "Sign out here", icon: LogOut },
  {
    id: "danger",
    label: "Danger zone",
    hint: "Delete forever",
    icon: AlertTriangle,
    tone: "danger",
  },
];

const PANEL: Record<
  SettingsSection,
  { eyebrow: string; title: string; description: string }
> = {
  appearance: {
    eyebrow: "Look & feel",
    title: "Appearance",
    description: "Pick light or dark. Preference stays on this browser.",
  },
  location: {
    eyebrow: "Feed",
    title: "Location",
    description:
      "Control whether the home feed uses your device location for nearby places.",
  },
  profile: {
    eyebrow: "Public identity",
    title: "Profile details",
    description: "What neighbors see when they open your page.",
  },
  security: {
    eyebrow: "Access",
    title: "Security",
    description: "Rotate your password. You’ll stay signed in on this device.",
  },
  session: {
    eyebrow: "This device",
    title: "Session",
    description: "End the current session without deleting your account.",
  },
  danger: {
    eyebrow: "Irreversible",
    title: "Danger zone",
    description:
      "Wipe the account and revoke tokens. Contributions may remain anonymized later.",
  },
};

export default function ProfileSettingsView({
  user,
  onBack,
}: {
  user: UserInfo;
  onBack: () => void;
}) {
  const [section, setSection] = useState<SettingsSection>("appearance");
  const meta = PANEL[section];

  return (
    <div className="mx-auto w-full max-w-[1152px] px-4 pt-10 pb-20 sm:px-6 sm:pt-14">
      <div className="mb-10 flex flex-wrap items-end justify-between gap-4 border-b border-border-subtle pb-8">
        <div className="min-w-0">
          <button
            type="button"
            onClick={onBack}
            className="mb-4 inline-flex items-center gap-2 text-sm text-fg-muted transition-colors hover:text-fg"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to profile
          </button>
          <p className="text-[10px] font-bold tracking-[0.22em] text-brand uppercase">
            Account control
          </p>
          <h1 className="mt-2 text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
            Settings
          </h1>
          <p className="mt-2 max-w-md text-sm text-fg-faint">
            Tune how Yet Bota looks and how your identity is stored — one lane at a time.
          </p>
        </div>

        <div className="flex items-center gap-3 rounded-2xl border border-border-subtle bg-surface px-3 py-2.5">
          <UserAvatar user={user} size="sm" />
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-fg">
              {user.full_name.trim() || user.username}
            </p>
            <p className="truncate text-xs text-fg-faint">@{user.username}</p>
          </div>
          <Link
            href="/"
            className="ml-2 hidden text-xs font-semibold text-fg-muted transition-colors hover:text-brand sm:inline"
          >
            Home
          </Link>
        </div>
      </div>

      <div className="mb-6 -mx-1 flex gap-2 overflow-x-auto px-1 pb-1 md:hidden [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {NAV.map((item) => {
          const Icon = item.icon;
          const active = section === item.id;
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => setSection(item.id)}
              className={cn(
                "inline-flex shrink-0 items-center gap-2 rounded-full px-3.5 py-2 text-xs font-semibold transition-colors",
                active
                  ? item.tone === "danger"
                    ? "bg-red-500/15 text-red-400 ring-1 ring-red-500/30"
                    : "bg-brand text-black"
                  : "bg-surface text-fg-muted ring-1 ring-border-subtle"
              )}
            >
              <Icon className="h-3.5 w-3.5" />
              {item.label}
            </button>
          );
        })}
      </div>

      <div className="grid gap-8 md:grid-cols-[240px_minmax(0,1fr)] lg:grid-cols-[260px_minmax(0,1fr)] lg:gap-10">
        <aside className="hidden md:block">
          <nav
            aria-label="Settings sections"
            className="sticky top-28 space-y-1 rounded-3xl border border-border-subtle bg-surface p-2 shadow-[0_20px_50px_-40px_rgba(0,0,0,0.5)]"
          >
            {NAV.map((item) => {
              const Icon = item.icon;
              const active = section === item.id;
              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => setSection(item.id)}
                  className={cn(
                    "group relative flex w-full items-start gap-3 rounded-2xl px-3 py-3 text-left transition-all",
                    active
                      ? item.tone === "danger"
                        ? "bg-red-500/10"
                        : "bg-brand/10"
                      : "hover:bg-overlay"
                  )}
                >
                  <span
                    className={cn(
                      "absolute top-3 bottom-3 left-0 w-1 rounded-full transition-opacity",
                      active
                        ? item.tone === "danger"
                          ? "bg-red-400 opacity-100"
                          : "bg-brand opacity-100"
                        : "opacity-0"
                    )}
                  />
                  <span
                    className={cn(
                      "mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border transition-colors",
                      active
                        ? item.tone === "danger"
                          ? "border-red-500/30 bg-red-500/15 text-red-400"
                          : "border-brand/30 bg-brand text-black"
                        : "border-border-subtle bg-bg text-fg-muted group-hover:text-fg"
                    )}
                  >
                    <Icon className="h-4 w-4" />
                  </span>
                  <span className="min-w-0 pt-0.5">
                    <span
                      className={cn(
                        "block text-sm font-semibold",
                        active
                          ? item.tone === "danger"
                            ? "text-red-400"
                            : "text-fg"
                          : "text-fg-muted group-hover:text-fg"
                      )}
                    >
                      {item.label}
                    </span>
                    <span className="mt-0.5 block text-[11px] leading-snug text-fg-faint">
                      {item.hint}
                    </span>
                  </span>
                </button>
              );
            })}
          </nav>
        </aside>

        <div
          key={section}
          className="rounded-3xl border border-border-subtle bg-surface p-6 sm:p-8"
        >
          <div className="mb-8 border-b border-border-subtle pb-6">
            <p
              className={cn(
                "text-[10px] font-bold tracking-[0.2em] uppercase",
                section === "danger" ? "text-red-400" : "text-brand"
              )}
            >
              {meta.eyebrow}
            </p>
            <h2 className="mt-2 text-2xl font-extrabold tracking-tight text-fg">{meta.title}</h2>
            <p className="mt-2 max-w-lg text-sm leading-relaxed text-fg-faint">
              {meta.description}
            </p>
          </div>

          {section === "appearance" && <AppearanceSettings framed={false} />}
          {section === "location" && <LocationSettings framed={false} />}
          {section === "profile" && (
            <ProfileEditForm
              key={`${user.user_id}-${user.avatar_url}-${user.full_name}-${user.phone}-${user.address}-${user.bio}`}
              user={user}
              framed={false}
            />
          )}
          {section === "security" && <ChangePasswordForm framed={false} />}
          {section === "session" && <SessionPanel />}
          {section === "danger" && <DangerZonePanel username={user.username} />}
        </div>
      </div>
    </div>
  );
}
