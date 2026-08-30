"use client";

import { useCallback, useState } from "react";
import PostCard from "@/components/posts/PostCard";
import type { PostsPage } from "@/lib/content/types";

export default function PostsGrid({
  initial,
  bookmarked,
  fromUser,
  onLoadMore,
  emptyTitle,
  emptyBody,
}: {
  initial: PostsPage | undefined;
  bookmarked?: boolean;
  fromUser?: boolean;
  onLoadMore?: (cursor: string) => Promise<PostsPage>;
  emptyTitle: string;
  emptyBody: string;
}) {
  const [extra, setExtra] = useState<PostsPage["posts"]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  const first = initial?.posts ?? [];
  const posts =
    extra.length === 0
      ? first
      : (() => {
          const seen = new Set(first.map((p) => p.id));
          return [...first, ...extra.filter((p) => !seen.has(p.id))];
        })();

  const next =
    cursor === null ? undefined : cursor === undefined ? initial?.next_cursor : cursor;

  const loadMore = useCallback(async () => {
    if (!next || !onLoadMore || loadingMore) return;
    setLoadingMore(true);
    try {
      const page = await onLoadMore(next);
      setExtra((prev) => [...prev, ...page.posts]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }, [next, onLoadMore, loadingMore]);

  if (!initial) {
    return (
      <div className="rounded-2xl border border-border-subtle bg-surface px-6 py-12 text-center text-sm text-fg-faint">
        Loading posts…
      </div>
    );
  }

  if (posts.length === 0) {
    return (
      <div className="rounded-2xl border border-dashed border-border-subtle bg-surface/60 px-6 py-14 text-center">
        <p className="text-sm font-semibold text-fg">{emptyTitle}</p>
        <p className="mx-auto mt-1 max-w-sm text-xs text-fg-faint">{emptyBody}</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {posts.map((post) => (
          <PostCard key={post.id} post={post} bookmarked={bookmarked} fromUser={fromUser} />
        ))}
      </div>
      {next && onLoadMore ? (
        <div className="flex justify-center pt-2">
          <button
            type="button"
            disabled={loadingMore}
            onClick={() => void loadMore()}
            className="rounded-full border border-border-subtle bg-surface px-5 py-2 text-sm font-semibold text-fg transition-colors hover:bg-surface-2 disabled:opacity-60"
          >
            {loadingMore ? "Loading…" : "Load more"}
          </button>
        </div>
      ) : null}
    </div>
  );
}
