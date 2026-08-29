"use client";

import { useState } from "react";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import ProfileBody, { ProfileLoadError } from "@/components/profile/ProfileBody";
import ProfileSettingsView from "@/components/profile/ProfileSettingsView";
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
      <ProfileLoadError
        message="Could not load your profile."
        onRetry={() => void refetch()}
      />
    );
  }

  if (view === "settings") {
    return <ProfileSettingsView user={user} onBack={() => setView("profile")} />;
  }

  return (
    <ProfileBody
      user={user}
      mode="own"
      onOpenSettings={() => setView("settings")}
    />
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
