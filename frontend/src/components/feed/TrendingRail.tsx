"use client";

import { useCallback, useEffect, useState, type ComponentType, type ReactNode } from "react";
import Link from "next/link";
import {
  ArrowUpRight,
  Flame,
  Heart,
  MapPin,
  Sparkles,
  Users,
  X,
} from "lucide-react";
import {
  useLazyListTrendingPostsQuery,
  useListTrendingCreatorsQuery,
  useListTrendingPostsQuery,
} from "@/lib/api/contentApi";
import {
  useGetUserProfileQuery,
  useListFollowersQuery,
} from "@/lib/api/socialApi";
import UserAvatar from "@/components/shared/UserAvatar";
import type { Post } from "@/lib/content/types";
import { placeTitle } from "@/lib/content/helpers";
import { cn } from "@/lib/utils";

const RAIL_LIMIT = 5;
/** Creators API is top-N only (no cursor); peek one past rail to decide “More”. */
const CREATORS_PEEK = RAIL_LIMIT + 1;
const CREATORS_MODAL_LIMIT = 50;
const POSTS_PAGE = 20;

function formatCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, "")}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, "")}k`;
  return String(n);
}

function postThumb(post: Post): string | null {
  const m = post.media.find((x) => x.kind === "image") ?? post.media[0];
  return m?.url?.trim() || null;
}

function SkeletonPulse({ className }: { className?: string }) {
  return (
    <div
      className={cn("animate-pulse rounded-md bg-overlay", className)}
      aria-hidden
    />
  );
}

function CreatorsSkeleton({ count = 5 }: { count?: number }) {
  return (
    <ul className="space-y-2.5" aria-hidden>
      {Array.from({ length: count }, (_, i) => (
        <li key={i} className="flex items-center gap-2.5">
          <SkeletonPulse className="h-9 w-9 !rounded-full" />
          <div className="min-w-0 flex-1 space-y-1.5">
            <SkeletonPulse className="h-3 w-24" />
            <SkeletonPulse className="h-2.5 w-28" />
          </div>
        </li>
      ))}
    </ul>
  );
}

function PostsSkeleton({ count = 5 }: { count?: number }) {
  return (
    <ul className="space-y-3" aria-hidden>
      {Array.from({ length: count }, (_, i) => (
        <li key={i} className="flex gap-3">
          <SkeletonPulse className="h-16 w-16 shrink-0 rounded-lg" />
          <div className="min-w-0 flex-1 space-y-2 pt-1">
            <SkeletonPulse className="h-3 w-full" />
            <SkeletonPulse className="h-3 w-2/3" />
            <SkeletonPulse className="h-2.5 w-12" />
          </div>
        </li>
      ))}
    </ul>
  );
}

function TrendingCreatorRow({
  userId,
  dense = true,
  onNavigate,
}: {
  userId: string;
  dense?: boolean;
  onNavigate?: () => void;
}) {
  const { data: user, isLoading } = useGetUserProfileQuery(userId);
  const { data: followers } = useListFollowersQuery(
    { id: userId, limit: 1 },
    { skip: !userId }
  );

  if (isLoading && !user) {
    return (
      <div className={cn("flex items-center gap-2.5", dense ? "py-1" : "px-3 py-2.5")}>
        <SkeletonPulse className={cn("!rounded-full", dense ? "h-9 w-9" : "h-10 w-10")} />
        <div className="min-w-0 flex-1 space-y-1.5">
          <SkeletonPulse className="h-3 w-24" />
          <SkeletonPulse className="h-2.5 w-28" />
        </div>
      </div>
    );
  }

  const level = user?.reputation_level?.trim() || "Explorer";
  const followerTotal = followers?.total;

  return (
    <Link
      href={`/users/${userId}`}
      onClick={onNavigate}
      className={cn(
        "group flex items-center gap-2.5 rounded-xl transition-colors hover:bg-brand/[0.06]",
        dense ? "py-1.5" : "px-3 py-2.5"
      )}
    >
      <UserAvatar
        user={
          user ?? {
            full_name: "",
            username: "…",
            avatar_url: "",
          }
        }
        size="xs"
        className={dense ? "!h-9 !w-9 !text-[10px]" : "!h-10 !w-10 !text-xs"}
      />
      <div className="min-w-0 flex-1">
        <p className="truncate text-[13px] font-semibold text-fg group-hover:text-brand">
          {user?.full_name.trim() || user?.username || "Creator"}
        </p>
        <div className="mt-0.5 flex min-w-0 items-center gap-1.5 text-[10px] text-fg-faint">
          <span className="shrink-0 rounded bg-brand/15 px-1.5 py-0.5 font-bold tracking-wide text-brand uppercase">
            {level}
          </span>
          <span className="inline-flex min-w-0 items-center gap-0.5 truncate">
            <Users className="h-2.5 w-2.5 shrink-0 opacity-70" aria-hidden />
            <span className="tabular-nums">
              {followerTotal != null ? formatCount(followerTotal) : "—"}
            </span>
          </span>
        </div>
      </div>
    </Link>
  );
}

function TrendingPostRow({
  post,
  dense = true,
  onNavigate,
}: {
  post: Post;
  dense?: boolean;
  onNavigate?: () => void;
}) {
  const thumb = postThumb(post);
  const caption = placeTitle(post);

  return (
    <Link
      href={`/posts/${post.id}`}
      onClick={onNavigate}
      className={cn(
        "group flex gap-3 rounded-xl transition-colors hover:bg-brand/[0.06]",
        dense ? "py-1.5" : "px-3 py-2.5"
      )}
    >
      <div
        className={cn(
          "relative shrink-0 overflow-hidden rounded-lg bg-surface-2",
          dense ? "h-16 w-16" : "h-[4.5rem] w-[4.5rem]"
        )}
      >
        {thumb ? (
          // eslint-disable-next-line @next/next/no-img-element -- Cloudinary URLs
          <img
            src={thumb}
            alt=""
            className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.04]"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <MapPin className="h-5 w-5 text-brand/55" />
          </div>
        )}
      </div>
      <div className="min-w-0 flex-1 py-0.5">
        <p className="line-clamp-2 text-[13px] leading-snug font-medium text-fg group-hover:text-brand">
          {caption}
        </p>
        <div className="mt-1 flex items-center gap-2 text-[11px] text-fg-faint">
          <span className="inline-flex items-center gap-1 tabular-nums">
            <Heart className="h-3 w-3 text-rose-400/90" aria-hidden />
            {formatCount(post.like_count)}
          </span>
          {post.hashtags[0] ? (
            <span className="truncate font-semibold tracking-wide text-brand/85 uppercase">
              #{post.hashtags[0]}
            </span>
          ) : null}
        </div>
      </div>
    </Link>
  );
}

function SectionHeading({
  children,
  icon: Icon = Flame,
}: {
  children: string;
  icon?: ComponentType<{ className?: string; "aria-hidden"?: boolean }>;
}) {
  return (
    <div className="mb-3.5 flex items-center gap-2">
      <span className="inline-flex h-6 w-6 items-center justify-center rounded-md bg-brand/12 text-brand">
        <Icon className="h-3.5 w-3.5" aria-hidden />
      </span>
      <h2 className="text-[11px] font-bold tracking-[0.16em] text-fg uppercase">
        {children}
      </h2>
    </div>
  );
}

function MoreButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="mt-2.5 group inline-flex w-full items-center gap-1 rounded-lg px-1 py-1.5 text-left text-[12px] font-semibold text-brand transition-colors hover:bg-brand/10"
    >
      More
      <ArrowUpRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
    </button>
  );
}

function TrendingEmpty({
  kind,
}: {
  kind: "creators" | "places";
}) {
  const isCreators = kind === "creators";

  return (
    <div className="overflow-hidden rounded-2xl border border-dashed border-border-subtle/90 bg-gradient-to-br from-brand/[0.07] via-transparent to-overlay px-3.5 py-4">
      {isCreators ? (
        <div className="mb-2.5 flex items-center" aria-hidden>
          {[0, 1, 2].map((i) => (
            <span
              key={i}
              className={cn(
                "flex h-7 w-7 items-center justify-center rounded-full border border-border-subtle bg-surface text-fg-faint",
                i > 0 && "-ml-2"
              )}
            >
              <Users className="h-3 w-3 opacity-70" />
            </span>
          ))}
        </div>
      ) : (
        <div className="mb-2.5 grid grid-cols-3 gap-1" aria-hidden>
          {[0, 1, 2].map((i) => (
            <span
              key={i}
              className="flex aspect-square items-center justify-center rounded-md border border-border-subtle/80 bg-surface/80"
            >
              <MapPin
                className={cn(
                  "h-3 w-3",
                  i === 1 ? "text-brand" : "text-fg-faint/70"
                )}
              />
            </span>
          ))}
        </div>
      )}

      <p className="text-[12px] text-fg-muted">
        {isCreators ? "No creators trending yet." : "No places trending yet."}
      </p>
      <Link
        href="/compose"
        className="mt-2 inline-flex items-center gap-1 text-[11px] font-semibold text-brand hover:opacity-80"
      >
        <Sparkles className="h-3 w-3" aria-hidden />
        Create a post
      </Link>
    </div>
  );
}

function TrendingModalShell({
  open,
  title,
  onClose,
  children,
}: {
  open: boolean;
  title: string;
  onClose: () => void;
  children: ReactNode;
}) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

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
        aria-labelledby="trending-modal-title"
        className="relative z-10 flex max-h-[75vh] w-full max-w-md flex-col rounded-t-2xl border border-border-subtle bg-surface sm:rounded-2xl"
      >
        <div className="flex items-center justify-between border-b border-border-subtle px-4 py-3">
          <h2 id="trending-modal-title" className="text-base font-bold text-fg">
            {title}
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
        <div className="overflow-y-auto px-2 py-2">{children}</div>
      </div>
    </div>
  );
}

function TrendingCreatorsModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const q = useListTrendingCreatorsQuery(
    { limit: CREATORS_MODAL_LIMIT },
    { skip: !open }
  );
  const creators = q.data?.creators ?? [];

  return (
    <TrendingModalShell open={open} title="Trending creators" onClose={onClose}>
      {q.isLoading ? (
        <div className="px-2 py-3">
          <CreatorsSkeleton count={8} />
        </div>
      ) : q.isError ? (
        <div className="flex flex-col items-center gap-3 px-3 py-8">
          <p className="text-sm text-fg-muted">Could not load creators.</p>
          <button
            type="button"
            onClick={() => void q.refetch()}
            className="text-sm font-semibold text-brand hover:underline"
          >
            Retry
          </button>
        </div>
      ) : creators.length === 0 ? (
        <div className="px-2 py-3">
          <TrendingEmpty kind="creators" />
        </div>
      ) : (
        <ul className="space-y-0.5">
          {creators.map((c) => (
            <li key={c.user_id}>
              <TrendingCreatorRow
                userId={c.user_id}
                dense={false}
                onNavigate={onClose}
              />
            </li>
          ))}
        </ul>
      )}
    </TrendingModalShell>
  );
}

function TrendingPostsModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const first = useListTrendingPostsQuery(
    { limit: POSTS_PAGE },
    { skip: !open }
  );
  const [fetchMore] = useLazyListTrendingPostsQuery();
  const [extra, setExtra] = useState<Post[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  useEffect(() => {
    if (!open) {
      setExtra([]);
      setCursor(undefined);
    }
  }, [open]);

  const posts =
    extra.length === 0
      ? (first.data?.posts ?? [])
      : (() => {
          const seen = new Set((first.data?.posts ?? []).map((p) => p.id));
          return [
            ...(first.data?.posts ?? []),
            ...extra.filter((p) => !seen.has(p.id)),
          ];
        })();

  const next =
    cursor === null
      ? undefined
      : cursor === undefined
        ? first.data?.next_cursor
        : cursor;

  const loadMore = useCallback(async () => {
    if (!next || loadingMore) return;
    setLoadingMore(true);
    try {
      const page = await fetchMore({ limit: POSTS_PAGE, cursor: next }).unwrap();
      setExtra((prev) => [...prev, ...page.posts]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }, [next, loadingMore, fetchMore]);

  return (
    <TrendingModalShell open={open} title="Trending places" onClose={onClose}>
      {first.isLoading ? (
        <div className="px-2 py-3">
          <PostsSkeleton count={6} />
        </div>
      ) : first.isError && posts.length === 0 ? (
        <div className="flex flex-col items-center gap-3 px-3 py-8">
          <p className="text-sm text-fg-muted">Could not load places.</p>
          <button
            type="button"
            onClick={() => void first.refetch()}
            className="text-sm font-semibold text-brand hover:underline"
          >
            Retry
          </button>
        </div>
      ) : posts.length === 0 ? (
        <div className="px-2 py-3">
          <TrendingEmpty kind="places" />
        </div>
      ) : (
        <>
          <ul className="space-y-0.5">
            {posts.map((p) => (
              <li key={p.id}>
                <TrendingPostRow post={p} dense={false} onNavigate={onClose} />
              </li>
            ))}
          </ul>
          {next ? (
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
    </TrendingModalShell>
  );
}

/** Thin, low-contrast scrollbar for the sticky rail. */
const railScrollClass =
  "[scrollbar-width:thin] [scrollbar-color:color-mix(in_oklab,var(--fg)_18%,transparent)_transparent] " +
  "[&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent " +
  "[&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-fg/15 " +
  "hover:[&::-webkit-scrollbar-thumb]:bg-fg/28";

export default function TrendingRail({ className }: { className?: string }) {
  const [creatorsOpen, setCreatorsOpen] = useState(false);
  const [postsOpen, setPostsOpen] = useState(false);

  const creatorsQ = useListTrendingCreatorsQuery({ limit: CREATORS_PEEK });
  const postsQ = useListTrendingPostsQuery({ limit: RAIL_LIMIT });

  const creatorsAll = creatorsQ.data?.creators ?? [];
  const creators = creatorsAll.slice(0, RAIL_LIMIT);
  const posts = postsQ.data?.posts ?? [];

  const showMoreCreators = creatorsAll.length > RAIL_LIMIT;
  const showMorePosts = Boolean(postsQ.data?.next_cursor);

  return (
    <>
      <aside
        className={cn(
          "flex max-h-[calc(100dvh-6.5rem)] w-60 shrink-0 flex-col gap-7 self-start overflow-y-auto overscroll-contain pr-2",
          railScrollClass,
          className
        )}
      >
        <section>
          <SectionHeading icon={Users}>Trending creators</SectionHeading>
          {creatorsQ.isError ? (
            <div className="rounded-2xl border border-border-subtle/80 bg-overlay/40 px-3.5 py-4">
              <p className="text-[12px] text-fg-muted">Couldn’t load</p>
              <button
                type="button"
                onClick={() => void creatorsQ.refetch()}
                className="mt-1.5 text-[11px] font-semibold text-brand hover:opacity-80"
              >
                Retry
              </button>
            </div>
          ) : creatorsQ.isLoading ? (
            <CreatorsSkeleton />
          ) : creators.length === 0 ? (
            <TrendingEmpty kind="creators" />
          ) : (
            <>
              <ul className="space-y-0.5">
                {creators.map((c) => (
                  <li key={c.user_id}>
                    <TrendingCreatorRow userId={c.user_id} />
                  </li>
                ))}
              </ul>
              {showMoreCreators ? (
                <MoreButton onClick={() => setCreatorsOpen(true)} />
              ) : null}
            </>
          )}
        </section>

        <section>
          <SectionHeading icon={Flame}>Trending places</SectionHeading>
          {postsQ.isError ? (
            <div className="rounded-2xl border border-border-subtle/80 bg-overlay/40 px-3.5 py-4">
              <p className="text-[12px] text-fg-muted">Couldn’t load</p>
              <button
                type="button"
                onClick={() => void postsQ.refetch()}
                className="mt-1.5 text-[11px] font-semibold text-brand hover:opacity-80"
              >
                Retry
              </button>
            </div>
          ) : postsQ.isLoading ? (
            <PostsSkeleton />
          ) : posts.length === 0 ? (
            <TrendingEmpty kind="places" />
          ) : (
            <>
              <ul className="space-y-1">
                {posts.map((p) => (
                  <li key={p.id}>
                    <TrendingPostRow post={p} />
                  </li>
                ))}
              </ul>
              {showMorePosts ? (
                <MoreButton onClick={() => setPostsOpen(true)} />
              ) : null}
            </>
          )}
        </section>
      </aside>

      <TrendingCreatorsModal
        open={creatorsOpen}
        onClose={() => setCreatorsOpen(false)}
      />
      <TrendingPostsModal open={postsOpen} onClose={() => setPostsOpen(false)} />
    </>
  );
}
