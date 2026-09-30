"use client";

import { use, useCallback, useState } from "react";
import Link from "next/link";
import {
  AlertTriangle,
  ArrowLeft,
  Lightbulb,
  MapPin,
  RefreshCw,
} from "lucide-react";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import PostsGrid from "@/components/posts/PostsGrid";
import {
  useGetPlaceQuery,
  useLazyListPlaceEntriesQuery,
  useLazyListPlacePostsQuery,
  useListPlaceEntriesQuery,
  useListPlacePostsQuery,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { formatCoord, relativeTime } from "@/lib/content/helpers";
import type { PlaceEntry, PlaceEntryCategory } from "@/lib/content/types";
import { cn } from "@/lib/utils";

const ENTRY_TABS: {
  id: PlaceEntryCategory | "all";
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}[] = [
  { id: "all", label: "All", icon: MapPin },
  { id: "tip", label: "Tips", icon: Lightbulb },
  { id: "change", label: "Changes", icon: RefreshCw },
  { id: "important", label: "Important", icon: AlertTriangle },
];

function categoryStyles(category: PlaceEntryCategory) {
  switch (category) {
    case "tip":
      return "bg-brand/15 text-brand";
    case "change":
      return "bg-amber-500/15 text-amber-600 dark:text-amber-400";
    case "important":
      return "bg-rose-500/15 text-rose-500";
  }
}

function EntryRow({ entry }: { entry: PlaceEntry }) {
  return (
    <li className="rounded-2xl border border-border-subtle bg-surface px-4 py-4">
      <div className="flex flex-wrap items-center gap-2">
        <span
          className={cn(
            "rounded px-1.5 py-0.5 text-[10px] font-bold tracking-wide uppercase",
            categoryStyles(entry.category)
          )}
        >
          {entry.category}
        </span>
        <span className="text-[10px] font-medium tracking-wide text-fg-faint uppercase">
          {entry.source === "ai" ? "AI" : "Community"}
        </span>
        <span className="text-xs text-fg-faint">
          {relativeTime(entry.created_at)}
        </span>
      </div>
      <p className="mt-2 whitespace-pre-wrap text-sm leading-relaxed text-fg">
        {entry.body}
      </p>
      {entry.source_post_id ? (
        <Link
          href={`/posts/${entry.source_post_id}`}
          className="mt-2 inline-block text-xs font-semibold text-brand hover:underline"
        >
          View source post
        </Link>
      ) : null}
    </li>
  );
}

function PlaceDetail({ placeId }: { placeId: string }) {
  const placeQ = useGetPlaceQuery(placeId);
  const [entryTab, setEntryTab] =
    useState<PlaceEntryCategory | "all">("all");
  const category = entryTab === "all" ? undefined : entryTab;

  const entriesQ = useListPlaceEntriesQuery({
    placeId,
    limit: 20,
    ...(category ? { category } : {}),
  });
  const [fetchMoreEntries] = useLazyListPlaceEntriesQuery();
  const [extraEntries, setExtraEntries] = useState<PlaceEntry[]>([]);
  const [entryCursor, setEntryCursor] = useState<string | null | undefined>(
    undefined
  );
  const [loadingEntries, setLoadingEntries] = useState(false);

  const postsQ = useListPlacePostsQuery({ placeId, limit: 20 });
  const [fetchMorePosts] = useLazyListPlacePostsQuery();

  const firstEntries = entriesQ.data?.entries ?? [];
  const entries =
    extraEntries.length === 0
      ? firstEntries
      : (() => {
          const seen = new Set(firstEntries.map((e) => e.id));
          return [
            ...firstEntries,
            ...extraEntries.filter((e) => !seen.has(e.id)),
          ];
        })();

  const nextEntry =
    entryCursor === null
      ? undefined
      : entryCursor === undefined
        ? entriesQ.data?.next_cursor
        : entryCursor;

  const loadMoreEntries = useCallback(async () => {
    if (!nextEntry || loadingEntries) return;
    setLoadingEntries(true);
    try {
      const page = await fetchMoreEntries({
        placeId,
        limit: 20,
        cursor: nextEntry,
        ...(category ? { category } : {}),
      }).unwrap();
      setExtraEntries((prev) => [...prev, ...page.entries]);
      setEntryCursor(page.next_cursor ?? null);
    } finally {
      setLoadingEntries(false);
    }
  }, [nextEntry, loadingEntries, fetchMoreEntries, placeId, category]);

  function switchEntryTab(next: PlaceEntryCategory | "all") {
    setEntryTab(next);
    setExtraEntries([]);
    setEntryCursor(undefined);
  }

  if (placeQ.isLoading) {
    return (
      <p className="py-20 text-center text-sm text-fg-faint">Loading place…</p>
    );
  }

  if (placeQ.isError || !placeQ.data) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-20 text-center">
        <p className="text-sm text-fg-muted">
          {getApiErrorMessage(placeQ.error) || "Place not found."}
        </p>
        <Link
          href="/places"
          className="mt-3 inline-block text-sm font-semibold text-brand hover:underline"
        >
          Back to places
        </Link>
      </div>
    );
  }

  const place = placeQ.data;
  const summary = place.summary.trim();

  return (
    <div className="mx-auto w-full max-w-3xl px-4 pt-8 pb-16 sm:px-6 sm:pt-10">
      <Link
        href="/places"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-fg-muted hover:text-fg"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to places
      </Link>

      <header className="rounded-2xl border border-border-subtle bg-surface p-5 sm:p-6">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <h1 className="text-xl font-bold tracking-tight text-fg sm:text-2xl">
            {place.name}
          </h1>
          <span className="shrink-0 rounded-full bg-overlay px-2.5 py-1 text-xs font-bold tabular-nums text-fg-muted">
            {place.post_count} {place.post_count === 1 ? "post" : "posts"}
          </span>
        </div>
        <a
          href={`https://www.openstreetmap.org/?mlat=${place.lat}&mlon=${place.lng}#map=17/${place.lat}/${place.lng}`}
          target="_blank"
          rel="noreferrer"
          className="mt-3 inline-flex items-center gap-1.5 text-sm text-fg-muted hover:text-fg"
        >
          <MapPin className="h-4 w-4 text-brand" />
          {formatCoord(place.lat, place.lng)}
        </a>

        <section className="mt-5 border-t border-border-subtle pt-5">
          <p className="text-[10px] font-bold tracking-widest text-brand uppercase">
            AI summary
          </p>
          {summary ? (
            <p className="mt-2 text-sm leading-relaxed whitespace-pre-wrap text-fg-muted">
              {summary}
            </p>
          ) : (
            <p className="mt-2 text-sm text-fg-faint">
              The summary will appear once enough posts are attached.
            </p>
          )}
          {place.summary_updated_at ? (
            <p className="mt-2 text-xs text-fg-faint">
              Updated {relativeTime(place.summary_updated_at)}
            </p>
          ) : null}
        </section>
      </header>

      <section className="mt-10">
        <h2 className="mb-4 text-sm font-bold tracking-wide text-fg-faint uppercase">
          Directory
        </h2>
        <div className="mb-4 inline-flex flex-wrap rounded-xl bg-overlay p-1">
          {ENTRY_TABS.map((t) => {
            const Icon = t.icon;
            const active = entryTab === t.id;
            return (
              <button
                key={t.id}
                type="button"
                onClick={() => switchEntryTab(t.id)}
                className={cn(
                  "inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors sm:text-sm",
                  active
                    ? "bg-surface text-fg shadow-sm"
                    : "text-fg-muted hover:text-fg"
                )}
              >
                <Icon className="h-3.5 w-3.5" />
                {t.label}
              </button>
            );
          })}
        </div>

        {entriesQ.isLoading ? (
          <p className="py-8 text-sm text-fg-faint">Loading entries…</p>
        ) : entriesQ.isError ? (
          <p className="text-sm text-fg-muted">
            {getApiErrorMessage(entriesQ.error)}
          </p>
        ) : entries.length === 0 ? (
          <p className="rounded-xl border border-dashed border-border-subtle px-4 py-8 text-center text-sm text-fg-faint">
            No {entryTab === "all" ? "directory" : entryTab} entries yet.
          </p>
        ) : (
          <ul className="space-y-3">
            {entries.map((e) => (
              <EntryRow key={e.id} entry={e} />
            ))}
          </ul>
        )}

        {nextEntry ? (
          <div className="mt-4 flex justify-center">
            <button
              type="button"
              disabled={loadingEntries}
              onClick={() => void loadMoreEntries()}
              className="rounded-full border border-border-subtle bg-surface px-5 py-2 text-sm font-semibold text-fg hover:bg-surface-2 disabled:opacity-60"
            >
              {loadingEntries ? "Loading…" : "Load more"}
            </button>
          </div>
        ) : null}
      </section>

      <section className="mt-10">
        <h2 className="mb-4 text-sm font-bold tracking-wide text-fg-faint uppercase">
          Posts about this place
        </h2>
        {postsQ.isError ? (
          <p className="text-sm text-fg-muted">
            {getApiErrorMessage(postsQ.error)}
          </p>
        ) : (
          <PostsGrid
            initial={postsQ.data}
            emptyTitle="No posts attached yet"
            emptyBody="Community drops about this spot will show up here after AI attach."
            onLoadMore={async (cursor) =>
              fetchMorePosts({ placeId, limit: 20, cursor }).unwrap()
            }
          />
        )}
      </section>
    </div>
  );
}

export default function PlacePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);

  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <PlaceDetail placeId={id} />
        </div>
      </main>
    </RequireAuth>
  );
}
