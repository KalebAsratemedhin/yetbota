"use client";

import { useCallback } from "react";
import FeedShell from "@/components/feed/FeedShell";
import SnapFeed from "@/components/feed/SnapFeed";
import {
  useLazyListRelatedPostsQuery,
  useListRelatedPostsQuery,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { Post } from "@/lib/content/types";

/**
 * Snap feed starting at `seed`, then related posts — same shell as the home feed.
 */
export default function RelatedPostFeed({
  seed,
  query,
  initialCommentsOpen = false,
}: {
  seed: Post;
  /** Original search query when opened from search results. */
  query?: string;
  initialCommentsOpen?: boolean;
}) {
  const q = query?.trim() || undefined;
  const relatedQ = useListRelatedPostsQuery({
    id: seed.id,
    ...(q ? { q } : {}),
    limit: 12,
  });
  const [fetchMore] = useLazyListRelatedPostsQuery();

  const onLoadMore = useCallback(
    async (cursor: string) => {
      return fetchMore({
        id: seed.id,
        ...(q ? { q } : {}),
        limit: 12,
        cursor,
      }).unwrap();
    },
    [fetchMore, q, seed.id]
  );

  if (relatedQ.isError && !relatedQ.data) {
    return (
      <FeedShell>
        <div className="flex min-h-[calc(100dvh-5rem)] flex-col items-center justify-center gap-3 px-4 text-center">
          <p className="text-sm text-fg-muted">
            {getApiErrorMessage(relatedQ.error) || "Could not load related posts."}
          </p>
          <button
            type="button"
            onClick={() => void relatedQ.refetch()}
            className="text-sm font-semibold text-brand hover:underline"
          >
            Retry
          </button>
        </div>
      </FeedShell>
    );
  }

  const rest = (relatedQ.data?.posts ?? []).filter((p) => p.id !== seed.id);
  const initialPosts = [seed, ...rest];

  return (
    <FeedShell>
      <SnapFeed
        key={seed.id}
        embedded
        initialPosts={initialPosts}
        nextCursor={relatedQ.data?.next_cursor}
        onLoadMore={onLoadMore}
        initialCommentsOpen={initialCommentsOpen}
      />
    </FeedShell>
  );
}
