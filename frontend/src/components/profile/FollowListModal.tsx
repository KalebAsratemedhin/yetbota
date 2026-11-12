"use client";

import { useCallback, useMemo, useState, type MouseEvent } from "react";
import Link from "next/link";
import { X } from "lucide-react";
import { toast } from "sonner";
import UserAvatar from "@/components/shared/UserAvatar";
import { useRequireSignIn } from "@/hooks/use-require-signin";
import {
  useFollowUserMutation,
  useLazyListFollowersQuery,
  useLazyListFollowingQuery,
  useListFollowersQuery,
  useListFollowingQuery,
  useUnfollowUserMutation,
} from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { UserInfo } from "@/lib/auth/types";
import { useAppSelector } from "@/lib/hooks";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 20;
const MY_FOLLOWING_LIMIT = 100;

function FollowListRow({
  user,
  isFollowing,
  isSelf,
  onToggle,
  onClose,
}: {
  user: UserInfo;
  isFollowing: boolean;
  isSelf: boolean;
  onToggle: (user: UserInfo, currentlyFollowing: boolean) => Promise<void>;
  onClose: () => void;
}) {
  const [busy, setBusy] = useState(false);

  async function handleClick(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (isSelf || busy) return;
    setBusy(true);
    try {
      await onToggle(user, isFollowing);
    } finally {
      setBusy(false);
    }
  }

  return (
    <li className="flex items-center gap-3 rounded-xl px-3 py-2.5">
      <Link
        href={`/users/${user.user_id}`}
        onClick={onClose}
        className="flex min-w-0 flex-1 items-center gap-3 transition-opacity hover:opacity-90"
      >
        <UserAvatar user={user} size="sm" />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-semibold text-fg">
            {user.full_name.trim() || user.username}
          </p>
          <p className="truncate text-xs text-fg-muted">@{user.username}</p>
        </div>
      </Link>
      {!isSelf ? (
        <button
          type="button"
          disabled={busy}
          onClick={(e) => void handleClick(e)}
          className={cn(
            "shrink-0 rounded-full px-3.5 py-1.5 text-xs font-bold transition-colors disabled:opacity-60",
            isFollowing
              ? "border border-border-subtle bg-surface text-fg hover:bg-surface-2"
              : "bg-brand text-black hover:bg-brand-dark"
          )}
        >
          {busy ? "…" : isFollowing ? "Following" : "Follow"}
        </button>
      ) : null}
    </li>
  );
}

export default function FollowListModal({
  open,
  title,
  userId,
  kind,
  onClose,
}: {
  open: boolean;
  title: string;
  userId: string;
  kind: "followers" | "following";
  onClose: () => void;
}) {
  const viewerId = useAppSelector((s) => s.auth.userId);
  const { requireSignIn } = useRequireSignIn();
  const listKey = `${kind}:${userId}`;
  const [activeKey, setActiveKey] = useState(listKey);
  const [extraUsers, setExtraUsers] = useState<UserInfo[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);
  /** Optimistic follow overrides keyed by user id. */
  const [followOverrides, setFollowOverrides] = useState<Record<string, boolean>>({});

  if (open && listKey !== activeKey) {
    setActiveKey(listKey);
    setExtraUsers([]);
    setNextCursor(undefined);
    setFollowOverrides({});
  }

  const followersQuery = useListFollowersQuery(
    { id: userId, limit: PAGE_SIZE },
    { skip: !open || kind !== "followers" }
  );
  const followingQuery = useListFollowingQuery(
    { id: userId, limit: PAGE_SIZE },
    { skip: !open || kind !== "following" }
  );
  const { data: myFollowing } = useListFollowingQuery(
    { id: viewerId ?? "", limit: MY_FOLLOWING_LIMIT },
    { skip: !open || !viewerId }
  );

  const page = kind === "followers" ? followersQuery.data : followingQuery.data;
  const isLoading = kind === "followers" ? followersQuery.isLoading : followingQuery.isLoading;
  const isError = kind === "followers" ? followersQuery.isError : followingQuery.isError;
  const refetch = kind === "followers" ? followersQuery.refetch : followingQuery.refetch;

  const [fetchFollowers] = useLazyListFollowersQuery();
  const [fetchFollowing] = useLazyListFollowingQuery();
  const [followUser] = useFollowUserMutation();
  const [unfollowUser] = useUnfollowUserMutation();

  const firstUsers = page?.users ?? [];
  const users =
    extraUsers.length === 0
      ? firstUsers
      : (() => {
          const seen = new Set(firstUsers.map((u) => u.user_id));
          return [...firstUsers, ...extraUsers.filter((u) => !seen.has(u.user_id))];
        })();

  const followingIds = useMemo(() => {
    const ids = new Set((myFollowing?.users ?? []).map((u) => u.user_id));
    // Own "Following" list: everyone listed is followed by the viewer.
    if (viewerId && userId === viewerId && kind === "following") {
      for (const u of users) ids.add(u.user_id);
    }
    return ids;
  }, [myFollowing?.users, viewerId, userId, kind, users]);

  function isFollowingUser(id: string): boolean {
    if (id in followOverrides) return followOverrides[id]!;
    return followingIds.has(id);
  }

  // null = exhausted; undefined = use first-page cursor; string = next page cursor
  const cursorForMore =
    nextCursor === null ? undefined : (nextCursor ?? page?.next_cursor);

  const total = page?.total;

  const loadMore = useCallback(async () => {
    if (!cursorForMore || loadingMore) return;
    setLoadingMore(true);
    try {
      const fetchList = kind === "followers" ? fetchFollowers : fetchFollowing;
      const result = await fetchList({
        id: userId,
        limit: PAGE_SIZE,
        cursor: cursorForMore,
      });
      if (result.data) {
        setExtraUsers((prev) => [...prev, ...result.data!.users]);
        setNextCursor(result.data.next_cursor ?? null);
      }
    } finally {
      setLoadingMore(false);
    }
  }, [cursorForMore, loadingMore, kind, fetchFollowers, fetchFollowing, userId]);

  const onToggle = useCallback(
    async (target: UserInfo, currentlyFollowing: boolean) => {
      if (!requireSignIn("Sign in to follow")) return;
      setFollowOverrides((prev) => ({ ...prev, [target.user_id]: !currentlyFollowing }));
      try {
        if (currentlyFollowing) {
          await unfollowUser(target.user_id).unwrap();
          toast.success("Unfollowed", { description: `@${target.username}` });
        } else {
          await followUser(target.user_id).unwrap();
          toast.success("Following", { description: `@${target.username}` });
        }
      } catch (err) {
        setFollowOverrides((prev) => {
          const next = { ...prev };
          delete next[target.user_id];
          return next;
        });
        toast.error(currentlyFollowing ? "Could not unfollow" : "Could not follow", {
          description: getApiErrorMessage(err),
        });
      }
    },
    [followUser, unfollowUser, requireSignIn]
  );

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center">
      <button
        type="button"
        className="absolute inset-0 bg-black/60"
        aria-label="Close"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="follow-list-title"
        className="relative z-10 flex max-h-[70vh] w-full max-w-md flex-col rounded-t-2xl border border-border-subtle bg-surface sm:rounded-2xl"
      >
        <div className="flex items-center justify-between border-b border-border-subtle px-4 py-3">
          <h2 id="follow-list-title" className="text-base font-bold text-fg">
            {title}
            {typeof total === "number" ? (
              <span className="ml-2 text-sm font-normal text-fg-faint">({total})</span>
            ) : null}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-fg-muted hover:bg-surface-2 hover:text-fg"
            aria-label="Close"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="overflow-y-auto px-2 py-2">
          {isLoading ? (
            <p className="px-3 py-8 text-center text-sm text-fg-faint">Loading…</p>
          ) : isError && users.length === 0 ? (
            <div className="flex flex-col items-center gap-3 px-3 py-8">
              <p className="text-sm text-fg-muted">Could not load list.</p>
              <button
                type="button"
                onClick={() => void refetch()}
                className="text-sm font-semibold text-brand hover:underline"
              >
                Retry
              </button>
            </div>
          ) : users.length === 0 ? (
            <p className="px-3 py-8 text-center text-sm text-fg-faint">No users yet.</p>
          ) : (
            <>
              <ul className="space-y-0.5">
                {users.map((u) => (
                  <FollowListRow
                    key={u.user_id}
                    user={u}
                    isFollowing={isFollowingUser(u.user_id)}
                    isSelf={Boolean(viewerId && u.user_id === viewerId)}
                    onToggle={onToggle}
                    onClose={onClose}
                  />
                ))}
              </ul>
              {cursorForMore ? (
                <div className="px-3 py-3">
                  <button
                    type="button"
                    onClick={() => void loadMore()}
                    disabled={loadingMore}
                    className="w-full rounded-xl border border-border-subtle bg-surface-2 py-2 text-sm font-semibold text-fg transition-colors hover:bg-overlay disabled:opacity-60"
                  >
                    {loadingMore ? "Loading…" : "Load more"}
                  </button>
                </div>
              ) : null}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
