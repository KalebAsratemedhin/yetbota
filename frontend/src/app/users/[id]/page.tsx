"use client";

import { use } from "react";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import ProfileBody, { ProfileLoadError } from "@/components/profile/ProfileBody";
import { useGetUserProfileQuery } from "@/lib/api/socialApi";
import { useAppSelector } from "@/lib/hooks";

function PublicProfileContent({ userId }: { userId: string }) {
  const router = useRouter();
  const viewerId = useAppSelector((s) => s.auth.userId);
  const { data: user, isLoading, isError, isFetching, refetch, error } =
    useGetUserProfileQuery(userId, { skip: !userId });

  useEffect(() => {
    if (viewerId && userId && viewerId === userId) {
      router.replace("/profile");
    }
  }, [viewerId, userId, router]);

  if (viewerId && viewerId === userId) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center text-sm text-fg-faint">
        Redirecting to your profile…
      </div>
    );
  }

  if (isLoading || (isFetching && !user)) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center text-sm text-fg-faint">
        Loading profile…
      </div>
    );
  }

  if (isError || !user) {
    const status =
      error && typeof error === "object" && "status" in error
        ? (error as { status: number | string }).status
        : undefined;
    const message =
      status === 404 ? "User not found." : "Could not load this profile.";
    return <ProfileLoadError message={message} onRetry={() => void refetch()} />;
  }

  return <ProfileBody user={user} mode="public" />;
}

export default function UserProfilePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);

  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg">
        <Navbar />
        <div className="pt-20">
          <PublicProfileContent userId={id} />
        </div>
      </main>
    </RequireAuth>
  );
}
