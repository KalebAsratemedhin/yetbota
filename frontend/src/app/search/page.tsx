"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Hash, MapPin, Search, Users } from "lucide-react";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import PlaceCard from "@/components/places/PlaceCard";
import PostsGrid from "@/components/posts/PostsGrid";
import UserAvatar from "@/components/shared/UserAvatar";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import {
  useLazySearchPostsQuery,
  useListPlacesQuery,
  useSearchPostsQuery,
} from "@/lib/api/contentApi";
import { useSearchUsersQuery } from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { cn } from "@/lib/utils";

const MIN_QUERY = 2;

type Tab = "people" | "directory" | "posts";

function parseTab(raw: string | null): Tab {
  if (raw === "posts") return "posts";
  if (raw === "places" || raw === "directory") return "directory";
  return "people";
}

function SearchContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialQ = searchParams.get("q") ?? "";
  const initialTab = parseTab(searchParams.get("tab"));
  const [query, setQuery] = useState(initialQ);
  const [tab, setTab] = useState<Tab>(initialTab);
  const debounced = useDebouncedValue(query.trim(), 300);
  const canSearch = debounced.length >= MIN_QUERY;

  const usersQ = useSearchUsersQuery(debounced, {
    skip: !canSearch || tab !== "people",
  });
  const postsQ = useSearchPostsQuery(
    { q: debounced, limit: 20 },
    { skip: !canSearch || tab !== "posts" }
  );
  const placesQ = useListPlacesQuery(
    { q: debounced, limit: 20 },
    { skip: !canSearch || tab !== "directory" }
  );
  const [fetchMorePosts] = useLazySearchPostsQuery();

  useEffect(() => {
    const nextQ = query.trim();
    const currentQ = searchParams.get("q") ?? "";
    const currentTab = parseTab(searchParams.get("tab"));
    if (nextQ === currentQ && tab === currentTab) return;
    const params = new URLSearchParams();
    if (nextQ) params.set("q", nextQ);
    if (tab !== "people") params.set("tab", tab);
    const qs = params.toString();
    router.replace(qs ? `/search?${qs}` : "/search", { scroll: false });
  }, [query, tab, router, searchParams]);

  const users = usersQ.data?.users ?? [];
  const places = placesQ.data?.places ?? [];
  const maybeTag = debounced.replace(/^#/, "");

  return (
    <div className="mx-auto w-full max-w-2xl px-4 pt-6 pb-12 sm:px-6 sm:pt-8">
      <h1 className="mb-1 text-2xl font-bold text-fg">Search</h1>
      <p className="mb-6 text-sm text-fg-muted">
        Find people, directory places, or post drops by text and hashtags.
      </p>

      <label className="relative block">
        <span className="sr-only">Search</span>
        <Search className="pointer-events-none absolute top-1/2 left-3.5 h-4 w-4 -translate-y-1/2 text-fg-faint" />
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          autoFocus
          autoComplete="off"
          placeholder="Type at least 2 characters…"
          className="w-full rounded-xl border border-border-subtle bg-surface-2 py-3 pr-4 pl-10 text-sm text-fg outline-none transition-all placeholder:text-gray-600 focus:border-brand/50 focus:ring-1 focus:ring-brand/20"
        />
      </label>

      <div className="mt-4 mb-6 inline-flex flex-wrap rounded-xl bg-overlay p-1">
        {(
          [
            { id: "people" as const, label: "People", icon: Users },
            { id: "directory" as const, label: "Directory", icon: MapPin },
            { id: "posts" as const, label: "Posts", icon: Hash },
          ] as const
        ).map((t) => {
          const Icon = t.icon;
          return (
            <button
              key={t.id}
              type="button"
              onClick={() => setTab(t.id)}
              className={cn(
                "inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors sm:text-sm",
                tab === t.id
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

      {tab === "posts" && canSearch ? (
        <Link
          href={`/hashtags/${encodeURIComponent(maybeTag)}`}
          className="mb-4 flex items-center gap-3 rounded-2xl border border-brand/25 bg-brand/10 px-4 py-3 transition-colors hover:bg-brand/15"
        >
          <Hash className="h-5 w-5 text-brand" />
          <div>
            <p className="text-sm font-semibold text-fg">Browse #{maybeTag}</p>
            <p className="text-xs text-fg-faint">Open the hashtag gallery</p>
          </div>
        </Link>
      ) : null}

      <div className="mt-2">
        {!canSearch ? (
          <p className="text-center text-sm text-fg-faint">
            {query.trim()
              ? "Keep typing — need at least 2 characters."
              : "Results appear as you type."}
          </p>
        ) : tab === "people" ? (
          usersQ.isFetching && !usersQ.data ? (
            <p className="text-center text-sm text-fg-faint">Searching people…</p>
          ) : usersQ.isError ? (
            <p className="text-center text-sm text-fg-muted">
              {getApiErrorMessage(usersQ.error)}
            </p>
          ) : users.length === 0 ? (
            <p className="text-center text-sm text-fg-faint">
              No users match “{debounced}”.
            </p>
          ) : (
            <ul className="space-y-1">
              {users.map((user) => (
                <li key={user.user_id}>
                  <Link
                    href={`/users/${user.user_id}`}
                    className="flex items-center gap-3 rounded-2xl border border-transparent px-3 py-3 transition-colors hover:border-border-subtle hover:bg-surface"
                  >
                    <UserAvatar user={user} size="md" />
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-semibold text-fg">
                        {user.full_name.trim() || user.username}
                      </p>
                      <p className="truncate text-sm text-fg-muted">@{user.username}</p>
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          )
        ) : tab === "directory" ? (
          placesQ.isFetching && !placesQ.data ? (
            <p className="text-center text-sm text-fg-faint">Searching places…</p>
          ) : placesQ.isError ? (
            <p className="text-center text-sm text-fg-muted">
              {getApiErrorMessage(placesQ.error)}
            </p>
          ) : places.length === 0 ? (
            <p className="text-center text-sm text-fg-faint">
              No directory places match “{debounced}”.
            </p>
          ) : (
            <ul className="space-y-3">
              {places.map((p) => (
                <PlaceCard key={p.id} place={p} />
              ))}
            </ul>
          )
        ) : postsQ.isError ? (
          <p className="text-center text-sm text-fg-muted">
            {getApiErrorMessage(postsQ.error)}
          </p>
        ) : (
          <PostsGrid
            initial={postsQ.data}
            emptyTitle="No matching posts"
            emptyBody="Try another phrase or open the hashtag shortcut above."
            searchQuery={debounced}
            onLoadMore={async (cursor) =>
              fetchMorePosts({ q: debounced, limit: 20, cursor }).unwrap()
            }
          />
        )}
      </div>
    </div>
  );
}

export default function SearchPage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <Suspense
            fallback={
              <div className="flex min-h-[40vh] items-center justify-center text-sm text-fg-faint">
                Loading search…
              </div>
            }
          >
            <SearchContent />
          </Suspense>
        </div>
      </main>
    </RequireAuth>
  );
}
