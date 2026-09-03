"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import CommentsPanel from "@/components/posts/CommentsPanel";
import PostDetailCard from "@/components/posts/PostDetailCard";
import PostMapPanel from "@/components/posts/PostMapPanel";
import { usePostSidePanel } from "@/hooks/use-post-side-panel";
import { useGetFeedQuery, useLazyGetFeedQuery } from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { prefetchPostMedia } from "@/lib/media/prefetch";
import type { Post } from "@/lib/content/types";
import { cn } from "@/lib/utils";

type GeoState =
  | { status: "pending" }
  | { status: "ready"; lat: number; lng: number }
  | { status: "unavailable" };

function readGeo(): Promise<{ lat: number; lng: number }> {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error("Geolocation is not available in this browser."));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) =>
        resolve({
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
        }),
      (err) => {
        reject(
          new Error(
            err.code === err.PERMISSION_DENIED
              ? "Location permission denied"
              : "Could not read location"
          )
        );
      },
      { enableHighAccuracy: true, timeout: 15000, maximumAge: 60_000 }
    );
  });
}

/** Which post owns A/V: last one whose top has crossed mid-viewport. */
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

function SnapFeed({
  initialPosts,
  nextCursor,
  onLoadMore,
}: {
  initialPosts: Post[];
  nextCursor?: string;
  onLoadMore: (cursor: string) => Promise<{ posts: Post[]; next_cursor?: string }>;
}) {
  const [posts, setPosts] = useState(initialPosts);
  const [cursor, setCursor] = useState<string | null | undefined>(nextCursor);
  const [loadingMore, setLoadingMore] = useState(false);
  const [focusedId, setFocusedId] = useState(initialPosts[0]?.id ?? "");
  const {
    commentsOpen,
    mapOpen,
    sideOpen,
    panelPostId,
    onCommentsClick,
    onMapClick,
    closeMobile,
    followFocusedPost,
  } = usePostSidePanel();
  const sentinelRef = useRef<HTMLDivElement>(null);
  const loadingRef = useRef(false);
  const idsRef = useRef(initialPosts.map((p) => p.id));

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
        "mx-auto flex w-full max-w-6xl justify-center gap-0 lg:gap-6 lg:px-4",
        sideOpen && "pb-[min(52vh,28rem)] lg:pb-0"
      )}
    >
      <div className="w-full max-w-xl">
        {posts.map((p) => (
          <section
            key={p.id}
            className="scroll-mt-20 box-border min-h-[calc(100dvh-5rem)] w-full"
          >
            <PostDetailCard
              post={p}
              active={p.id === focusedId}
              layout="snap"
              onOpenComments={() => onCommentsClick(p.id)}
              onOpenMap={() => onMapClick(p.id)}
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
    </div>
  );
}

export default function HomeFeed() {
  const [geo, setGeo] = useState<GeoState>({ status: "pending" });
  const [geoTick, setGeoTick] = useState(0);

  useEffect(() => {
    let cancelled = false;
    void readGeo()
      .then((coords) => {
        if (!cancelled) setGeo({ status: "ready", ...coords });
      })
      .catch(() => {
        if (!cancelled) setGeo({ status: "unavailable" });
      });
    return () => {
      cancelled = true;
    };
  }, [geoTick]);

  useEffect(() => {
    if (geo.status !== "unavailable") return;
    toast.message("Location off", {
      id: "feed-geo-unavailable",
      description: "Showing following / recent posts. Enable location for nearby.",
      duration: 5000,
      action: {
        label: "Enable",
        onClick: () => {
          setGeo({ status: "pending" });
          setGeoTick((n) => n + 1);
        },
      },
    });
  }, [geo.status]);

  const feedArgs = useMemo(
    () =>
      geo.status === "ready"
        ? { lat: geo.lat, lng: geo.lng, limit: 12 }
        : { limit: 12 },
    [geo]
  );

  const feedQ = useGetFeedQuery(feedArgs);
  const [fetchMore] = useLazyGetFeedQuery();

  const onLoadMore = useCallback(
    async (cursor: string) => {
      return fetchMore({
        ...feedArgs,
        cursor,
      }).unwrap();
    },
    [fetchMore, feedArgs]
  );

  if (feedQ.isError && !feedQ.data) {
    return (
      <div className="flex min-h-[calc(100dvh-5rem)] flex-col items-center justify-center gap-3 px-4 text-center">
        <p className="text-sm text-fg-muted">
          {getApiErrorMessage(feedQ.error) || "Could not load feed."}
        </p>
        <button
          type="button"
          onClick={() => void feedQ.refetch()}
          className="text-sm font-semibold text-brand hover:underline"
        >
          Retry
        </button>
      </div>
    );
  }

  if (feedQ.isLoading || !feedQ.data) {
    return (
      <div className="flex min-h-[calc(100dvh-5rem)] flex-col items-center justify-center gap-3 px-4">
        <Loader2 className="h-6 w-6 animate-spin text-brand" />
        <p className="text-sm text-fg-muted">Loading your feed…</p>
      </div>
    );
  }

  const feedKey =
    geo.status === "ready"
      ? `${geo.lat.toFixed(3)}:${geo.lng.toFixed(3)}`
      : "no-geo";

  return (
    <SnapFeed
      key={feedKey}
      initialPosts={feedQ.data.posts}
      nextCursor={feedQ.data.next_cursor}
      onLoadMore={onLoadMore}
    />
  );
}
