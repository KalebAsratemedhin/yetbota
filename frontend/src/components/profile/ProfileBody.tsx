"use client";

import { useCallback, useState } from "react";
import Link from "next/link";
import { toast } from "sonner";
import ContributionsPanel from "@/components/profile/ContributionsPanel";
import FollowListModal from "@/components/profile/FollowListModal";
import ProfileHeader from "@/components/profile/ProfileHeader";
import { BadgesCard, ReputationCard } from "@/components/profile/ReputationBadges";
import { useRequireSignIn } from "@/hooks/use-require-signin";
import {
  useFollowUserMutation,
  useListFollowersQuery,
  useListFollowingQuery,
  useUnfollowUserMutation,
} from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { UserInfo } from "@/lib/auth/types";
import { useAppSelector } from "@/lib/hooks";

type ListKind = "followers" | "following" | null;

export default function ProfileBody({
  user,
  mode,
  onOpenSettings,
}: {
  user: UserInfo;
  mode: "own" | "public";
  onOpenSettings?: () => void;
}) {
  const viewerId = useAppSelector((s) => s.auth.userId);
  const { isSignedIn, requireSignIn } = useRequireSignIn();
  const [listKind, setListKind] = useState<ListKind>(null);
  const [followOverride, setFollowOverride] = useState<boolean | null>(null);
  const [trackedUserId, setTrackedUserId] = useState(user.user_id);

  if (user.user_id !== trackedUserId) {
    setTrackedUserId(user.user_id);
    setFollowOverride(null);
  }

  // Followers/following lists stay auth-required on identity.
  const { data: followers } = useListFollowersQuery(
    { id: user.user_id, limit: 20 },
    { skip: !isSignedIn }
  );
  const { data: following } = useListFollowingQuery(
    { id: user.user_id, limit: 20 },
    { skip: !isSignedIn }
  );
  const { data: myFollowing } = useListFollowingQuery(
    { id: viewerId ?? "", limit: 50 },
    { skip: !viewerId || mode !== "public" }
  );

  const [followUser, { isLoading: followBusy }] = useFollowUserMutation();
  const [unfollowUser, { isLoading: unfollowBusy }] = useUnfollowUserMutation();

  const fromApi =
    Boolean(viewerId) &&
    (Boolean(myFollowing?.users.some((u) => u.user_id === user.user_id)) ||
      Boolean(followers?.users.some((u) => u.user_id === viewerId)));

  const isFollowing = followOverride ?? fromApi;

  const toggleFollow = useCallback(async () => {
    if (!requireSignIn("Sign in to follow")) return;
    setFollowOverride(!isFollowing);
    try {
      if (isFollowing) {
        await unfollowUser(user.user_id).unwrap();
        toast.success("Unfollowed", { description: `@${user.username}` });
      } else {
        await followUser(user.user_id).unwrap();
        toast.success("Following", { description: `@${user.username}` });
      }
    } catch (err) {
      setFollowOverride(null);
      toast.error(isFollowing ? "Could not unfollow" : "Could not follow", {
        description: getApiErrorMessage(err),
      });
    }
  }, [
    followUser,
    unfollowUser,
    isFollowing,
    requireSignIn,
    user.user_id,
    user.username,
  ]);

  return (
    <div className="bg-bg text-fg">
      <div className="mx-auto w-full max-w-[1152px] px-4 pt-4 sm:px-6 sm:pt-6">
        <ProfileHeader
          user={user}
          mode={mode}
          followerCount={isSignedIn ? followers?.total : undefined}
          followingCount={isSignedIn ? following?.total : undefined}
          onOpenFollowers={
            isSignedIn ? () => setListKind("followers") : undefined
          }
          onOpenFollowing={
            isSignedIn ? () => setListKind("following") : undefined
          }
          onOpenSettings={onOpenSettings}
          isFollowing={isFollowing}
          followBusy={followBusy || unfollowBusy}
          onToggleFollow={mode === "public" ? toggleFollow : undefined}
        />
      </div>

      <div
        className={`mx-auto w-full max-w-[1152px] px-4 pb-8 sm:px-6 sm:pb-10 ${
          mode === "own" ? "pt-8 sm:pt-10" : "pt-6 sm:pt-8"
        }`}
      >
        <div className="grid grid-cols-12 gap-6 sm:gap-8">
          <div className="col-span-12 lg:col-span-4">
            <ReputationCard
              score={user.reputation_score ?? 0}
              level={user.reputation_level || "Explorer"}
            />
          </div>
          <div className="col-span-12 lg:col-span-8">
            <BadgesCard badges={user.badges ?? []} />
          </div>
          <div className="col-span-12">
            <ContributionsPanel userId={user.user_id} mode={mode} />
          </div>
        </div>
      </div>

      {listKind ? (
        <FollowListModal
          open
          title={listKind === "following" ? "Following" : "Followers"}
          userId={user.user_id}
          kind={listKind}
          onClose={() => setListKind(null)}
        />
      ) : null}
    </div>
  );
}

export function ProfileLoadError({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 px-4">
      <p className="text-sm text-fg-muted">{message}</p>
      {onRetry ? (
        <button
          type="button"
          onClick={onRetry}
          className="text-sm font-semibold text-brand hover:underline"
        >
          Retry
        </button>
      ) : null}
      <Link href="/" className="text-sm text-fg-faint hover:text-fg">
        Home
      </Link>
    </div>
  );
}
