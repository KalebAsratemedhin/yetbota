"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import CommentsPanel from "@/components/posts/CommentsPanel";
import { commentsGridClass } from "@/components/posts/commentsLayout";
import PostDetailCard from "@/components/posts/PostDetailCard";
import { useCommentsPanel } from "@/hooks/use-comments-panel";
import { useLazyListUserPostsQuery } from "@/lib/api/contentApi";
import { encodeContentCursor } from "@/lib/content/cursor";
import { prefetchPostMedia } from "@/lib/media/prefetch";
import type { Post } from "@/lib/content/types";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 3;

/**
 * Which post should own A/V: the last one whose top edge has crossed mid-screen.
 * When the next post's top hits mid-viewport, the previous one must stop.
 */
function pickFocusedPostId(orderedIds: string[]): string | null {
  if (orderedIds.length === 0) return null;
  const mid = window.innerHeight * 0.5;
  let focused = orderedIds[0]!;
  for (const id of orderedIds) {
    const el = document.querySelector<HTMLElement>(`[data-post-id="${id}"]`);
    if (!el) continue;
    if (el.getBoundingClientRect().top <= mid) focused = id;
    else break;
  }
  return focused;
}

function AuthorPostFeedSession({
  seed,
  initialMobileComments,
}: {
  seed: Post;
  initialMobileComments?: boolean;
}) {
  const initialCursor = (() => {
    try {
      return encodeContentCursor(seed.created_at, seed.id);
    } catch {
      return null;
    }
  })();

  const [posts, setPosts] = useState<Post[]>([seed]);
  const [cursor, setCursor] = useState<string | null>(initialCursor);
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(initialCursor == null);
  const [focusedId, setFocusedId] = useState(seed.id);
  const {
    mobileOpen: mobileCommentsOpen,
    desktopOpen: desktopCommentsOpen,
    panelPostId,
    onCommentsClick,
    closeMobile,
    followFocusedPost,
  } = useCommentsPanel({
    initialMobileOpen: Boolean(initialMobileComments),
    initialDesktopOpen: Boolean(initialMobileComments),
  });
  const [fetchPage] = useLazyListUserPostsQuery();
  const sentinelRef = useRef<HTMLDivElement>(null);
  const loadingRef = useRef(false);
  const idsRef = useRef<string[]>([seed.id]);

  useEffect(() => {
    idsRef.current = posts.map((p) => p.id);
    prefetchPostMedia(seed);
    for (const p of posts.slice(0, 3)) prefetchPostMedia(p);
  }, [posts, seed]);

  const refreshFocus = useCallback(() => {
    const next = pickFocusedPostId(idsRef.current);
    if (next) setFocusedId(next);
  }, []);

  useEffect(() => {
    refreshFocus();
    window.addEventListener("scroll", refreshFocus, { passive: true });
    window.addEventListener("resize", refreshFocus);
    return () => {
      window.removeEventListener("scroll", refreshFocus);
      window.removeEventListener("resize", refreshFocus);
    };
  }, [refreshFocus, posts.length]);

  useEffect(() => {
    const idx = posts.findIndex((p) => p.id === focusedId);
    if (idx < 0) return;
    for (const p of posts.slice(Math.max(0, idx - 1), idx + 2)) {
      prefetchPostMedia(p);
    }
  }, [focusedId, posts]);

  useEffect(() => {
    followFocusedPost(focusedId);
  }, [focusedId, followFocusedPost]);

  const loadMore = useCallback(async () => {
    if (!cursor || loadingRef.current || done) return;
    loadingRef.current = true;
    setLoading(true);
    try {
      const page = await fetchPage({
        id: seed.author_id,
        limit: PAGE_SIZE,
        cursor,
      }).unwrap();
      setPosts((prev) => {
        const seen = new Set(prev.map((p) => p.id));
        const next = page.posts.filter((p) => !seen.has(p.id));
        return next.length ? [...prev, ...next] : prev;
      });
      if (!page.next_cursor || page.posts.length === 0) {
        setDone(true);
        setCursor(null);
      } else {
        setCursor(page.next_cursor);
      }
    } catch {
      setDone(true);
    } finally {
      loadingRef.current = false;
      setLoading(false);
    }
  }, [cursor, done, fetchPage, seed.author_id]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || done) return;
    const obs = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) void loadMore();
      },
      { root: null, rootMargin: "800px 0px", threshold: 0 }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [loadMore, done]);

  return (
    <div
      className={cn(
        commentsGridClass(desktopCommentsOpen),
        mobileCommentsOpen && "pb-[min(52vh,28rem)] lg:pb-0"
      )}
    >
      <div className="mx-auto w-full max-w-xl lg:mx-0">
        {posts.map((p) => (
          <PostDetailCard
            key={p.id}
            post={p}
            active={p.id === focusedId}
            fromUser
            onOpenComments={() => onCommentsClick(p.id)}
          />
        ))}
        <div ref={sentinelRef} className="h-10" aria-hidden />
        {loading ? (
          <p className="py-6 text-center text-sm text-fg-faint">Loading more…</p>
        ) : null}
        {done && posts.length > 1 ? (
          <p className="py-8 text-center text-xs text-fg-faint">No more posts</p>
        ) : null}
      </div>

      <CommentsPanel
        postId={panelPostId || focusedId}
        mobileOpen={mobileCommentsOpen}
        onMobileClose={closeMobile}
        desktopOpen={desktopCommentsOpen}
      />
    </div>
  );
}

/**
 * Continues an author's posts below the opened one (profile → post flow).
 * Remounts when seed changes (keyed by id).
 */
export default function AuthorPostFeed({
  seed,
  initialMobileComments,
}: {
  seed: Post;
  initialMobileComments?: boolean;
}) {
  return (
    <AuthorPostFeedSession
      key={seed.id}
      seed={seed}
      initialMobileComments={initialMobileComments}
    />
  );
}
