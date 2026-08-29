"use client";

import { useState } from "react";
import Link from "next/link";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import ContributionsPanel from "@/components/profile/ContributionsPanel";
import ProfileHeader from "@/components/profile/ProfileHeader";
import ProfileSettingsView from "@/components/profile/ProfileSettingsView";
import { BadgesCard, ReputationCard } from "@/components/profile/ReputationBadges";
import { useMeQuery } from "@/lib/api/authApi";

type View = "profile" | "settings";

function ProfileContent() {
  const [view, setView] = useState<View>("profile");
  const { data: user, isLoading, isError, refetch, isFetching } = useMeQuery();

  if (isLoading || (isFetching && !user)) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center text-sm text-fg-faint">
        Loading profile…
      </div>
    );
  }

  if (isError || !user) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 px-4">
        <p className="text-sm text-fg-muted">Could not load your profile.</p>
        <button
          type="button"
          onClick={() => void refetch()}
          className="text-sm font-semibold text-brand hover:underline"
        >
          Retry
        </button>
        <Link href="/" className="text-sm text-fg-faint hover:text-fg">
          Home
        </Link>
      </div>
    );
  }

  if (view === "settings") {
    return <ProfileSettingsView user={user} onBack={() => setView("profile")} />;
  }

  return (
    <div className="bg-bg text-fg">
      <div className="mx-auto w-full max-w-[1152px] px-4 pt-4 sm:px-6 sm:pt-6">
        <ProfileHeader user={user} onOpenSettings={() => setView("settings")} />
      </div>

      <div className="mx-auto w-full max-w-[1152px] px-4 pt-8 pb-8 sm:px-6 sm:pt-10 sm:pb-10">
        <div className="grid grid-cols-12 gap-6 sm:gap-8">
          <div className="col-span-12 lg:col-span-4">
            <ReputationCard />
          </div>
          <div className="col-span-12 lg:col-span-8">
            <BadgesCard />
          </div>
          <div className="col-span-12">
            <ContributionsPanel />
          </div>
        </div>
      </div>
    </div>
  );
}

export default function ProfilePage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg">
        <Navbar />
        <div className="pt-20">
          <ProfileContent />
        </div>
      </main>
    </RequireAuth>
  );
}
