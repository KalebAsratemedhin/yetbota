"use client";

import Link from "next/link";
import { Bookmark, HelpCircle, MapPin, Plus } from "lucide-react";
import PostsGrid from "@/components/posts/PostsGrid";
import {
  useLazyListMyBookmarksQuery,
  useLazyListUserPostsQuery,
  useListMyBookmarksQuery,
  useListUserPostsQuery,
} from "@/lib/api/contentApi";
import { cn } from "@/lib/utils";
import { useState } from "react";

type ContributionTab = "locations" | "qa" | "saved";

export default function ContributionsPanel({
  userId,
  mode,
}: {
  userId: string;
  mode: "own" | "public";
}) {
  const tabs: { id: ContributionTab; label: string }[] = [
    { id: "locations", label: "Locations" },
    { id: "qa", label: "Q&A" },
    ...(mode === "own" ? [{ id: "saved" as const, label: "Saved" }] : []),
  ];
  const [tab, setTab] = useState<ContributionTab>("locations");

  const postsQ = useListUserPostsQuery(
    { id: userId, limit: 20 },
    { skip: tab !== "locations" }
  );
  const [fetchMorePosts] = useLazyListUserPostsQuery();

  const bookmarksQ = useListMyBookmarksQuery(
    { limit: 20 },
    { skip: tab !== "saved" || mode !== "own" }
  );
  const [fetchMoreBookmarks] = useLazyListMyBookmarksQuery();

  return (
    <div className="rounded-2xl border border-border-subtle bg-bg p-4 sm:p-5">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h3 className="text-lg font-semibold text-fg">
          {mode === "own" ? "Your contributions" : "Contributions"}
        </h3>
        {mode === "own" ? (
          <Link
            href="/compose"
            className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-brand text-black shadow-[0_0_24px_-6px_rgba(34,197,94,0.7)] transition-transform hover:scale-105"
            aria-label="Create post"
            title="Create post"
          >
            <Plus className="h-5 w-5" strokeWidth={2.5} />
          </Link>
        ) : null}
      </div>

      <div
        role="tablist"
        aria-label="Contribution filters"
        className="mb-5 inline-flex w-fit rounded-xl bg-overlay p-1"
      >
        {tabs.map((t) => (
          <button
            key={t.id}
            type="button"
            role="tab"
            aria-selected={tab === t.id}
            onClick={() => setTab(t.id)}
            className={cn(
              "rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors sm:text-sm",
              tab === t.id
                ? "bg-surface text-fg shadow-sm"
                : "text-fg-muted hover:text-fg"
            )}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "locations" ? (
        postsQ.isError ? (
          <button
            type="button"
            onClick={() => void postsQ.refetch()}
            className="text-sm font-semibold text-brand hover:underline"
          >
            Couldn’t load posts — retry
          </button>
        ) : (
          <PostsGrid
            initial={postsQ.data}
            fromUser
            emptyTitle={mode === "own" ? "No places yet" : "No posts yet"}
            emptyBody={
              mode === "own"
                ? "Pin your first drop from the green + button."
                : "This explorer hasn’t published places yet."
            }
            onLoadMore={async (cursor) =>
              fetchMorePosts({ id: userId, limit: 20, cursor }).unwrap()
            }
          />
        )
      ) : null}

      {tab === "saved" && mode === "own" ? (
        bookmarksQ.isError ? (
          <button
            type="button"
            onClick={() => void bookmarksQ.refetch()}
            className="text-sm font-semibold text-brand hover:underline"
          >
            Couldn’t load saved — retry
          </button>
        ) : (
          <PostsGrid
            initial={bookmarksQ.data}
            bookmarked
            emptyTitle="Nothing saved"
            emptyBody="Bookmark places from a post to collect them here."
            onLoadMore={async (cursor) =>
              fetchMoreBookmarks({ limit: 20, cursor }).unwrap()
            }
          />
        )
      ) : null}

      {tab === "qa" ? (
        <div className="flex flex-col items-center justify-center gap-3 rounded-xl border border-border-subtle bg-surface px-6 py-12 text-center">
          <HelpCircle className="h-8 w-8 text-fg-faint" />
          <p className="text-sm font-semibold text-fg">Community Q&A</p>
          <p className="max-w-xs text-xs text-fg-faint">
            Ask and answer place questions — votes, answers, and closed threads.
          </p>
          <div className="mt-1 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/qa"
              className="text-sm font-semibold text-brand hover:underline"
            >
              Browse Q&A
            </Link>
            {mode === "own" ? (
              <Link
                href="/qa/ask"
                className="inline-flex items-center gap-1 rounded-full bg-brand px-3 py-1.5 text-xs font-bold text-black hover:bg-brand-dark"
              >
                <Plus className="h-3.5 w-3.5" strokeWidth={2.5} />
                Ask
              </Link>
            ) : null}
          </div>
        </div>
      ) : null}

      {tab === "locations" && mode === "own" && postsQ.data?.posts.length === 0 ? (
        <div className="mt-3 flex items-center justify-center gap-2 text-[11px] text-fg-faint">
          <MapPin className="h-3 w-3" />
          Tips: photo drops travel farthest with a sharp caption + hashtags.
        </div>
      ) : null}

      {tab === "saved" && mode === "own" && bookmarksQ.data?.posts.length === 0 ? (
        <div className="mt-3 flex items-center justify-center gap-2 text-[11px] text-fg-faint">
          <Bookmark className="h-3 w-3" />
          Saved lists are private — only you see this tab.
        </div>
      ) : null}
    </div>
  );
}
