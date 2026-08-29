"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Search } from "lucide-react";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import UserAvatar from "@/components/shared/UserAvatar";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useSearchUsersQuery } from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";

const MIN_QUERY = 2;

function SearchContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialQ = searchParams.get("q") ?? "";
  const [query, setQuery] = useState(initialQ);
  const debounced = useDebouncedValue(query.trim(), 300);
  const canSearch = debounced.length >= MIN_QUERY;

  const { data, isFetching, isError, error, isUninitialized } = useSearchUsersQuery(
    debounced,
    { skip: !canSearch }
  );

  useEffect(() => {
    const next = query.trim();
    const current = searchParams.get("q") ?? "";
    if (next === current) return;
    const url = next ? `/search?q=${encodeURIComponent(next)}` : "/search";
    router.replace(url, { scroll: false });
  }, [query, router, searchParams]);

  const users = data?.users ?? [];

  return (
    <div className="mx-auto w-full max-w-xl px-4 pt-6 pb-12 sm:px-6 sm:pt-8">
      <h1 className="mb-1 text-2xl font-bold text-fg">Find people</h1>
      <p className="mb-6 text-sm text-fg-muted">
        Search by username or full name.
      </p>

      <label className="relative block">
        <span className="sr-only">Search users</span>
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

      <div className="mt-6">
        {!canSearch ? (
          <p className="text-center text-sm text-fg-faint">
            {query.trim()
              ? "Keep typing — need at least 2 characters."
              : "Results appear as you type."}
          </p>
        ) : isFetching && !data ? (
          <p className="text-center text-sm text-fg-faint">Searching…</p>
        ) : isError ? (
          <p className="text-center text-sm text-fg-muted">
            {getApiErrorMessage(error)}
          </p>
        ) : users.length === 0 && !isUninitialized ? (
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
        )}
        {canSearch && isFetching && data ? (
          <p className="mt-3 text-center text-xs text-fg-faint">Updating…</p>
        ) : null}
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
