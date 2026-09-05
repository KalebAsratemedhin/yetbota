"use client";

import { useCallback, useMemo } from "react";
import FeedShell from "@/components/feed/FeedShell";
import SnapFeed from "@/components/feed/SnapFeed";
import { useLazyListUserPostsQuery } from "@/lib/api/contentApi";
import { encodeContentCursor } from "@/lib/content/cursor";
import type { Post } from "@/lib/content/types";

const PAGE_SIZE = 6;

/**
 * Profile → post: same shell + snap UI as the home feed, continuing the author's posts.
 */
export default function AuthorPostFeed({
  seed,
  initialMobileComments = false,
}: {
  seed: Post;
  initialMobileComments?: boolean;
}) {
  const [fetchPage] = useLazyListUserPostsQuery();

  const initialCursor = useMemo(() => {
    try {
      return encodeContentCursor(seed.created_at, seed.id);
    } catch {
      return undefined;
    }
  }, [seed.created_at, seed.id]);

  const onLoadMore = useCallback(
    async (cursor: string) => {
      return fetchPage({
        id: seed.author_id,
        limit: PAGE_SIZE,
        cursor,
      }).unwrap();
    },
    [fetchPage, seed.author_id]
  );

  return (
    <FeedShell>
      <SnapFeed
        key={seed.id}
        embedded
        fromUser
        initialPosts={[seed]}
        nextCursor={initialCursor}
        onLoadMore={onLoadMore}
        initialCommentsOpen={initialMobileComments}
      />
    </FeedShell>
  );
}
