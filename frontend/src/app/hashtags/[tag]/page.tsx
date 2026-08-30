"use client";

import { use } from "react";
import Link from "next/link";
import { ArrowLeft, Hash } from "lucide-react";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import PostsGrid from "@/components/posts/PostsGrid";
import {
  useLazyListHashtagPostsQuery,
  useListHashtagPostsQuery,
} from "@/lib/api/contentApi";

function HashtagContent({ tag }: { tag: string }) {
  const clean = decodeURIComponent(tag).replace(/^#/, "");
  const { data, isLoading, isError, refetch } = useListHashtagPostsQuery({
    tag: clean,
    limit: 20,
  });
  const [fetchMore] = useLazyListHashtagPostsQuery();

  return (
    <div className="mx-auto w-full max-w-[1152px] px-4 pt-6 pb-16 sm:px-6 sm:pt-8">
      <Link
        href="/search?tab=posts"
        className="mb-4 inline-flex items-center gap-1.5 text-sm text-fg-muted hover:text-fg"
      >
        <ArrowLeft className="h-4 w-4" /> Search
      </Link>

      <div className="mb-8 flex items-end gap-3">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-brand/15 text-brand">
          <Hash className="h-7 w-7" />
        </div>
        <div>
          <p className="text-[10px] font-bold tracking-[0.18em] text-fg-faint uppercase">
            Hashtag
          </p>
          <h1 className="text-3xl font-extrabold text-fg">#{clean}</h1>
        </div>
      </div>

      {isLoading ? (
        <p className="text-sm text-fg-faint">Loading tagged places…</p>
      ) : isError ? (
        <button
          type="button"
          onClick={() => void refetch()}
          className="text-sm font-semibold text-brand hover:underline"
        >
          Couldn’t load — retry
        </button>
      ) : (
        <PostsGrid
          initial={data}
          emptyTitle="No posts with this tag"
          emptyBody="Be the first to publish something under this hashtag."
          onLoadMore={async (cursor) =>
            fetchMore({ tag: clean, limit: 20, cursor }).unwrap()
          }
        />
      )}
    </div>
  );
}

export default function HashtagPage({
  params,
}: {
  params: Promise<{ tag: string }>;
}) {
  const { tag } = use(params);

  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <HashtagContent tag={tag} />
        </div>
      </main>
    </RequireAuth>
  );
}
