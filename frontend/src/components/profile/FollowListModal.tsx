"use client";

import { useCallback, useState } from "react";
import Link from "next/link";
import { X } from "lucide-react";
import UserAvatar from "@/components/shared/UserAvatar";
import {
  useLazyListFollowersQuery,
  useLazyListFollowingQuery,
  useListFollowersQuery,
  useListFollowingQuery,
} from "@/lib/api/socialApi";
import type { UserInfo } from "@/lib/auth/types";

const PAGE_SIZE = 20;

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
  const listKey = `${kind}:${userId}`;
  const [activeKey, setActiveKey] = useState(listKey);
  const [extraUsers, setExtraUsers] = useState<UserInfo[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  if (open && listKey !== activeKey) {
    setActiveKey(listKey);
    setExtraUsers([]);
    setNextCursor(undefined);
  }

  const followersQuery = useListFollowersQuery(
    { id: userId, limit: PAGE_SIZE },
    { skip: !open || kind !== "followers" }
  );
  const followingQuery = useListFollowingQuery(
    { id: userId, limit: PAGE_SIZE },
    { skip: !open || kind !== "following" }
  );

  const page = kind === "followers" ? followersQuery.data : followingQuery.data;
  const isLoading = kind === "followers" ? followersQuery.isLoading : followingQuery.isLoading;
  const isError = kind === "followers" ? followersQuery.isError : followingQuery.isError;
  const refetch = kind === "followers" ? followersQuery.refetch : followingQuery.refetch;

  const [fetchFollowers] = useLazyListFollowersQuery();
  const [fetchFollowing] = useLazyListFollowingQuery();

  const firstUsers = page?.users ?? [];
  const users =
    extraUsers.length === 0
      ? firstUsers
      : (() => {
          const seen = new Set(firstUsers.map((u) => u.user_id));
          return [...firstUsers, ...extraUsers.filter((u) => !seen.has(u.user_id))];
        })();

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
                  <li key={u.user_id}>
                    <Link
                      href={`/users/${u.user_id}`}
                      onClick={onClose}
                      className="flex items-center gap-3 rounded-xl px-3 py-2.5 transition-colors hover:bg-surface-2"
                    >
                      <UserAvatar user={u} size="sm" />
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-semibold text-fg">
                          {u.full_name.trim() || u.username}
                        </p>
                        <p className="truncate text-xs text-fg-muted">@{u.username}</p>
                      </div>
                    </Link>
                  </li>
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
