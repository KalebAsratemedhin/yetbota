"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Loader2 } from "lucide-react";
import CommentsPanel from "@/components/posts/CommentsPanel";
import PostDetailCard from "@/components/posts/PostDetailCard";
import PostMapPanel from "@/components/posts/PostMapPanel";
import PostQuestionsPanel from "@/components/posts/PostQuestionsPanel";
import { usePostSidePanel } from "@/hooks/use-post-side-panel";
import { prefetchPostMedia } from "@/lib/media/prefetch";
import type { Post } from "@/lib/content/types";
import { cn } from "@/lib/utils";

/** Which post owns A/V: last one whose top has crossed mid-viewport. */
export function pickFocusedPostId(orderedIds: string[]): string | null {
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

export default function SnapFeed({
  initialPosts,
  nextCursor,
  onLoadMore,
  initialCommentsOpen = false,
  embedded = false,
  fromUser = false,
}: {
  initialPosts: Post[];
  nextCursor?: string;
  onLoadMore: (cursor: string) => Promise<{ posts: Post[]; next_cursor?: string }>;
  initialCommentsOpen?: boolean;
  /** When true, parent owns page width (e.g. home with trending rail). */
  embedded?: boolean;
  /** Profile → post continuation; keeps `from=user` comment links. */
  fromUser?: boolean;
}) {
  const [posts, setPosts] = useState(initialPosts);
  const [cursor, setCursor] = useState<string | null | undefined>(nextCursor);
  const [loadingMore, setLoadingMore] = useState(false);
  const [focusedId, setFocusedId] = useState(initialPosts[0]?.id ?? "");
  const {
    commentsOpen,
    mapOpen,
    qaOpen,
    sideOpen,
    panelPostId,
    onCommentsClick,
    onMapClick,
    onQaClick,
    closeMobile,
    followFocusedPost,
  } = usePostSidePanel({
    initialMode: initialCommentsOpen ? "comments" : "none",
    initialPostId: initialPosts[0]?.id ?? "",
  });
  const sentinelRef = useRef<HTMLDivElement>(null);
  const loadingRef = useRef(false);
  const idsRef = useRef(initialPosts.map((p) => p.id));

  // Parent may grow the list (e.g. seed post → full feed) without remounting.
  useEffect(() => {
    setPosts((prev) => {
      const ordered: Post[] = [];
      const used = new Set<string>();
      for (const p of initialPosts) {
        ordered.push(p);
        used.add(p.id);
      }
      for (const p of prev) {
        if (!used.has(p.id)) ordered.push(p);
      }
      if (
        ordered.length === prev.length &&
        ordered.every((p, i) => p.id === prev[i]?.id)
      ) {
        return prev;
      }
      return ordered;
    });
  }, [initialPosts]);

  useEffect(() => {
    if (nextCursor === undefined) return;
    setCursor((current) => {
      if (current === null) return null;
      if (current === undefined) return nextCursor ?? null;
      return current;
    });
  }, [nextCursor]);

  const activeSidePost =
    posts.find((p) => p.id === (panelPostId || focusedId)) ?? posts[0];

  useEffect(() => {
    idsRef.current = posts.map((p) => p.id);
    for (const p of posts.slice(0, 3)) prefetchPostMedia(p);
  }, [posts]);

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
    if (!cursor || loadingRef.current) return;
    loadingRef.current = true;
    setLoadingMore(true);
    try {
      const page = await onLoadMore(cursor);
      setPosts((prev) => {
        const seen = new Set(prev.map((p) => p.id));
        const next = page.posts.filter((p) => !seen.has(p.id));
        return next.length ? [...prev, ...next] : prev;
      });
      setCursor(page.next_cursor ?? null);
    } catch {
      setCursor(null);
    } finally {
      loadingRef.current = false;
      setLoadingMore(false);
    }
  }, [cursor, onLoadMore]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || !cursor) return;
    const obs = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) void loadMore();
      },
      { root: null, rootMargin: "200% 0px", threshold: 0 }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [loadMore, cursor]);

  if (posts.length === 0) {
    return (
      <div className="flex min-h-[calc(100dvh-5rem)] flex-col items-center justify-center gap-3 px-4 text-center">
        <p className="text-sm font-semibold text-fg">Your feed is quiet</p>
        <p className="max-w-sm text-xs text-fg-faint">
          Follow people and drop places — posts from your network (and nearby, when
          location is on) will show up here.
        </p>
        <Link
          href="/compose"
          className="mt-2 rounded-full bg-brand px-4 py-2 text-xs font-bold text-black hover:bg-brand-dark"
        >
          Drop a place
        </Link>
      </div>
    );
  }

  return (
    <div
      className={cn(
        "flex w-full gap-0",
        embedded
          ? cn(
              "lg:gap-8",
              sideOpen
                ? "justify-start px-3 sm:px-5 lg:pl-6 lg:pr-4"
                : "justify-center px-3 sm:px-4"
            )
          : "mx-auto max-w-6xl justify-center lg:gap-8 lg:px-6",
        sideOpen && "pb-[min(52vh,28rem)] lg:pb-0"
      )}
    >
      <div
        className={cn(
          "w-full max-w-xl",
          sideOpen && embedded && "shrink-0 lg:max-w-[36rem]"
        )}
      >
        {posts.map((p) => (
          <section
            key={p.id}
            className="scroll-mt-20 box-border min-h-[calc(100dvh-5rem)] w-full"
          >
            <PostDetailCard
              post={p}
              active={p.id === focusedId}
              layout="snap"
              fromUser={fromUser}
              onOpenComments={() => onCommentsClick(p.id)}
              onOpenMap={() => onMapClick(p.id)}
              onOpenQa={() => onQaClick(p.id)}
            />
          </section>
        ))}
        <div ref={sentinelRef} className="h-px w-full" aria-hidden />
        {loadingMore ? (
          <div className="flex items-center justify-center gap-2 py-6 text-sm text-fg-faint">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading more…
          </div>
        ) : null}
        {cursor === null ? (
          <p className="py-8 text-center text-xs text-fg-faint">You’re caught up</p>
        ) : null}
      </div>

      {commentsOpen ? (
        <CommentsPanel
          postId={panelPostId || focusedId}
          mobileOpen={commentsOpen}
          onMobileClose={closeMobile}
          desktopOpen={commentsOpen}
        />
      ) : null}
      {mapOpen && activeSidePost ? (
        <PostMapPanel
          lat={activeSidePost.lat}
          lng={activeSidePost.lng}
          mobileOpen={mapOpen}
          onMobileClose={closeMobile}
          desktopOpen={mapOpen}
        />
      ) : null}
      {qaOpen && activeSidePost ? (
        <PostQuestionsPanel
          postId={activeSidePost.id}
          lat={activeSidePost.lat}
          lng={activeSidePost.lng}
          mobileOpen={qaOpen}
          onMobileClose={closeMobile}
          desktopOpen={qaOpen}
        />
      ) : null}
    </div>
  );
}
