"use client";

import { useEffect, useState } from "react";
import { Check, Search, X } from "lucide-react";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import {
  useGetPostQuery,
  useLazySearchPostsQuery,
  useSearchPostsQuery,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { coverUrl, placeTitle, relativeTime } from "@/lib/content/helpers";
import type { Post } from "@/lib/content/types";
import { cn } from "@/lib/utils";

const MIN_QUERY = 2;

function PostPickRow({
  post,
  selected,
  onSelect,
}: {
  post: Post;
  selected: boolean;
  onSelect: (post: Post) => void;
}) {
  const cover = coverUrl(post);
  return (
    <button
      type="button"
      onClick={() => onSelect(post)}
      className={cn(
        "flex w-full items-center gap-3 rounded-xl border px-3 py-2.5 text-left transition-colors",
        selected
          ? "border-brand/50 bg-brand/10"
          : "border-border-subtle bg-surface hover:border-brand/30 hover:bg-surface-2"
      )}
    >
      <div className="relative h-12 w-12 shrink-0 overflow-hidden rounded-lg bg-overlay">
        {cover ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={cover} alt="" className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full items-center justify-center text-[10px] text-fg-faint">
            —
          </div>
        )}
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-semibold text-fg">{placeTitle(post)}</p>
        <p className="truncate text-xs text-fg-faint">
          {post.caption.trim() || relativeTime(post.created_at)}
        </p>
      </div>
      {selected ? <Check className="h-4 w-4 shrink-0 text-brand" /> : null}
    </button>
  );
}

export default function PostSearchPicker({
  selectedPostId,
  onSelect,
  onClear,
}: {
  selectedPostId: string | null;
  onSelect: (post: Post) => void;
  onClear: () => void;
}) {
  const [query, setQuery] = useState("");
  const debounced = useDebouncedValue(query.trim(), 300);
  const canSearch = debounced.length >= MIN_QUERY;
  const selectedQ = useGetPostQuery(selectedPostId ?? "", {
    skip: !selectedPostId,
  });
  const postsQ = useSearchPostsQuery(
    { q: debounced, limit: 20 },
    { skip: !canSearch }
  );
  const [fetchMore] = useLazySearchPostsQuery();
  const [extra, setExtra] = useState<Post[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  useEffect(() => {
    setExtra([]);
    setCursor(undefined);
  }, [debounced]);

  const first = postsQ.data?.posts ?? [];
  const posts =
    extra.length === 0
      ? first
      : (() => {
          const seen = new Set(first.map((p) => p.id));
          return [...first, ...extra.filter((p) => !seen.has(p.id))];
        })();

  const next =
    cursor === null
      ? undefined
      : cursor === undefined
        ? postsQ.data?.next_cursor
        : cursor;

  async function loadMore() {
    if (!next || loadingMore) return;
    setLoadingMore(true);
    try {
      const page = await fetchMore({ q: debounced, limit: 20, cursor: next }).unwrap();
      setExtra((prev) => [...prev, ...page.posts]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }

  const selected = selectedQ.data;

  return (
    <div className="flex h-full min-h-0 flex-col bg-surface">
      <div className="shrink-0 border-b border-border-subtle px-4 py-3">
        <p className="text-sm font-semibold text-fg">Link a place post</p>
        <p className="mt-0.5 text-xs text-fg-muted">
          Search places already on YetBota.
        </p>
        <label className="relative mt-3 block">
          <span className="sr-only">Search posts</span>
          <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-fg-faint" />
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoComplete="off"
            placeholder="Search places…"
            className="w-full rounded-xl border border-border-subtle bg-bg py-2.5 pr-3 pl-9 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
          />
        </label>
      </div>

      {selectedPostId ? (
        <div className="shrink-0 border-b border-border-subtle px-4 py-3">
          <div className="mb-2 flex items-center justify-between gap-2">
            <p className="text-[11px] font-bold tracking-wide text-brand uppercase">
              Selected
            </p>
            <button
              type="button"
              onClick={onClear}
              className="inline-flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-semibold text-fg-muted hover:bg-surface-2 hover:text-fg"
            >
              <X className="h-3 w-3" />
              Clear
            </button>
          </div>
          {selected ? (
            <PostPickRow post={selected} selected onSelect={onSelect} />
          ) : (
            <p className="text-sm text-fg-faint">Loading place…</p>
          )}
        </div>
      ) : null}

      <div className="yb-scrollbar min-h-0 flex-1 overflow-y-auto px-4 py-3">
        {!canSearch ? (
          <p className="py-10 text-center text-sm text-fg-faint">
            {query.trim()
              ? "Keep typing — need at least 2 characters."
              : "Type to find a place post."}
          </p>
        ) : postsQ.isFetching && !postsQ.data ? (
          <p className="py-10 text-center text-sm text-fg-faint">Searching…</p>
        ) : postsQ.isError ? (
          <p className="py-10 text-center text-sm text-fg-muted">
            {getApiErrorMessage(postsQ.error)}
          </p>
        ) : posts.length === 0 ? (
          <p className="py-10 text-center text-sm text-fg-faint">
            No places match “{debounced}”.
          </p>
        ) : (
          <ul className="space-y-2">
            {posts.map((post) => (
              <li key={post.id}>
                <PostPickRow
                  post={post}
                  selected={post.id === selectedPostId}
                  onSelect={onSelect}
                />
              </li>
            ))}
          </ul>
        )}
        {next ? (
          <button
            type="button"
            disabled={loadingMore}
            onClick={() => void loadMore()}
            className="mt-3 w-full rounded-xl border border-border-subtle bg-bg py-2 text-sm font-semibold text-fg hover:bg-overlay disabled:opacity-60"
          >
            {loadingMore ? "Loading…" : "Load more"}
          </button>
        ) : null}
      </div>
    </div>
  );
}
