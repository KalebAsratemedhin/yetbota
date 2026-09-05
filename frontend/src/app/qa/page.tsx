"use client";

import { useCallback, useState } from "react";
import Link from "next/link";
import { HelpCircle, MessageSquare, Plus } from "lucide-react";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import UserAvatar from "@/components/shared/UserAvatar";
import {
  useLazyListQuestionsQuery,
  useListQuestionsQuery,
} from "@/lib/api/contentApi";
import { useGetUserProfileQuery } from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { relativeTime } from "@/lib/content/helpers";
import type { Question } from "@/lib/content/types";

function QuestionAuthor({ userId }: { userId: string }) {
  const { data } = useGetUserProfileQuery(userId);
  return (
    <span className="inline-flex items-center gap-1.5">
      <UserAvatar
        user={data ?? { full_name: "", username: "…", avatar_url: "" }}
        size="xs"
        className="!h-5 !w-5 !text-[8px]"
      />
      <span className="text-xs text-fg-muted">
        @{data?.username ?? "…"}
      </span>
    </span>
  );
}

function QuestionRow({ q }: { q: Question }) {
  const score = q.upvote_count - q.downvote_count;
  const closed = Boolean(q.closed_at);

  return (
    <li>
      <Link
        href={`/qa/${q.id}`}
        className="flex gap-4 rounded-2xl border border-border-subtle bg-surface px-4 py-4 transition-colors hover:border-brand/30 hover:bg-surface-2"
      >
        <div className="flex w-12 shrink-0 flex-col items-center justify-center rounded-xl bg-overlay py-2">
          <span className="text-sm font-bold tabular-nums text-fg">{score}</span>
          <span className="text-[10px] text-fg-faint">votes</span>
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold text-fg">{q.title}</h2>
            {closed ? (
              <span className="rounded bg-fg-faint/20 px-1.5 py-0.5 text-[10px] font-bold tracking-wide text-fg-muted uppercase">
                Closed
              </span>
            ) : null}
          </div>
          <p className="mt-1 line-clamp-2 text-sm text-fg-muted">{q.body}</p>
          <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-fg-faint">
            <QuestionAuthor userId={q.author_id} />
            <span>{relativeTime(q.created_at)}</span>
            <span className="inline-flex items-center gap-1">
              <MessageSquare className="h-3 w-3" />
              {q.answer_count} {q.answer_count === 1 ? "answer" : "answers"}
            </span>
          </div>
        </div>
      </Link>
    </li>
  );
}

function QaList() {
  const listQ = useListQuestionsQuery({ limit: 20 });
  const [fetchMore] = useLazyListQuestionsQuery();
  const [extra, setExtra] = useState<Question[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  const first = listQ.data?.questions ?? [];
  const questions =
    extra.length === 0
      ? first
      : (() => {
          const seen = new Set(first.map((q) => q.id));
          return [...first, ...extra.filter((q) => !seen.has(q.id))];
        })();

  const next =
    cursor === null
      ? undefined
      : cursor === undefined
        ? listQ.data?.next_cursor
        : cursor;

  const loadMore = useCallback(async () => {
    if (!next || loadingMore) return;
    setLoadingMore(true);
    try {
      const page = await fetchMore({ limit: 20, cursor: next }).unwrap();
      setExtra((prev) => [...prev, ...page.questions]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }, [next, loadingMore, fetchMore]);

  return (
    <div className="mx-auto w-full max-w-3xl px-4 pt-8 pb-16 sm:px-6 sm:pt-10">
      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-[10px] font-bold tracking-[0.2em] text-brand uppercase">
            Community
          </p>
          <h1 className="mt-1 text-3xl font-extrabold tracking-tight text-fg">
            Q&A
          </h1>
          <p className="mt-2 max-w-md text-sm text-fg-muted">
            Ask about places, routes, and local tips — Stack Overflow style.
          </p>
        </div>
        <Link
          href="/qa/ask"
          className="inline-flex h-10 items-center gap-1.5 rounded-full bg-brand px-4 text-sm font-bold text-black hover:bg-brand-dark"
        >
          <Plus className="h-4 w-4" strokeWidth={2.5} />
          Ask
        </Link>
      </div>

      {listQ.isLoading ? (
        <p className="py-16 text-center text-sm text-fg-faint">Loading questions…</p>
      ) : listQ.isError ? (
        <div className="py-16 text-center">
          <p className="text-sm text-fg-muted">
            {getApiErrorMessage(listQ.error) || "Could not load questions."}
          </p>
          <button
            type="button"
            onClick={() => void listQ.refetch()}
            className="mt-3 text-sm font-semibold text-brand hover:underline"
          >
            Retry
          </button>
        </div>
      ) : questions.length === 0 ? (
        <div className="flex flex-col items-center gap-3 rounded-2xl border border-dashed border-border-subtle px-6 py-16 text-center">
          <HelpCircle className="h-8 w-8 text-fg-faint" />
          <p className="text-sm font-semibold text-fg">No questions yet</p>
          <p className="max-w-sm text-xs text-fg-faint">
            Be the first to ask the community about a place or route.
          </p>
          <Link
            href="/qa/ask"
            className="mt-2 text-sm font-semibold text-brand hover:underline"
          >
            Ask a question
          </Link>
        </div>
      ) : (
        <>
          <ul className="space-y-3">
            {questions.map((q) => (
              <QuestionRow key={q.id} q={q} />
            ))}
          </ul>
          {next ? (
            <div className="mt-6 flex justify-center">
              <button
                type="button"
                disabled={loadingMore}
                onClick={() => void loadMore()}
                className="rounded-full border border-border-subtle bg-surface px-5 py-2 text-sm font-semibold text-fg hover:bg-surface-2 disabled:opacity-60"
              >
                {loadingMore ? "Loading…" : "Load more"}
              </button>
            </div>
          ) : null}
        </>
      )}
    </div>
  );
}

export default function QaPage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <QaList />
        </div>
      </main>
    </RequireAuth>
  );
}
