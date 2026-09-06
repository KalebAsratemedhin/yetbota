"use client";

import { useCallback, useEffect, useState, type ReactNode } from "react";
import Link from "next/link";
import { CircleHelp, MessageSquare, X } from "lucide-react";
import {
  useLazyListQuestionsQuery,
  useListQuestionsQuery,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { relativeTime } from "@/lib/content/helpers";
import type { Question } from "@/lib/content/types";
import { cn } from "@/lib/utils";

function PanelChrome({
  onClose,
  children,
  className,
  grabber,
  showClose,
}: {
  onClose?: () => void;
  children: ReactNode;
  className?: string;
  grabber?: boolean;
  showClose?: boolean;
}) {
  return (
    <div className={cn("flex h-full min-h-0 flex-col bg-surface", className)}>
      {grabber ? (
        <div className="flex shrink-0 justify-center pt-2" aria-hidden>
          <span className="h-1 w-10 rounded-full bg-border-subtle" />
        </div>
      ) : null}
      <header className="flex shrink-0 items-center justify-between gap-3 px-4 py-2.5">
        <h2 className="inline-flex items-center gap-1.5 text-sm font-semibold text-fg">
          <CircleHelp className="h-4 w-4 text-brand" />
          Place Q&A
        </h2>
        {showClose && onClose ? (
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1.5 text-fg-muted hover:bg-surface-2 hover:text-fg"
            aria-label="Close Q&A"
          >
            <X className="h-4 w-4" />
          </button>
        ) : (
          <span className="w-8" aria-hidden />
        )}
      </header>
      {children}
    </div>
  );
}

function QuestionsBody({ postId }: { postId: string }) {
  const listQ = useListQuestionsQuery({ post_id: postId, limit: 20 });
  const [fetchMore] = useLazyListQuestionsQuery();
  const [extra, setExtra] = useState<Question[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  useEffect(() => {
    setExtra([]);
    setCursor(undefined);
  }, [postId]);

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
      const page = await fetchMore({
        post_id: postId,
        limit: 20,
        cursor: next,
      }).unwrap();
      setExtra((prev) => [...prev, ...page.questions]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }, [next, loadingMore, fetchMore, postId]);

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="yb-scrollbar min-h-0 flex-1 overflow-y-auto px-3 py-2">
        {listQ.isLoading ? (
          <p className="px-2 py-10 text-center text-sm text-fg-faint">
            Loading questions…
          </p>
        ) : listQ.isError ? (
          <div className="px-2 py-10 text-center">
            <p className="text-sm text-fg-muted">
              {getApiErrorMessage(listQ.error) || "Could not load questions."}
            </p>
            <button
              type="button"
              onClick={() => void listQ.refetch()}
              className="mt-2 text-sm font-semibold text-brand hover:underline"
            >
              Retry
            </button>
          </div>
        ) : questions.length === 0 ? (
          <p className="px-2 py-10 text-center text-sm text-fg-faint">
            No questions about this place yet.
          </p>
        ) : (
          <ul className="space-y-2">
            {questions.map((q) => (
              <li key={q.id}>
                <Link
                  href={`/qa/${q.id}`}
                  className="block rounded-xl border border-border-subtle bg-bg px-3 py-3 transition-colors hover:border-brand/30 hover:bg-overlay/60"
                >
                  <div className="flex items-start justify-between gap-2">
                    <p className="line-clamp-4 text-xs leading-relaxed text-fg">
                      {q.body}
                    </p>
                    {q.closed_at ? (
                      <span className="shrink-0 text-[10px] font-bold tracking-wide text-fg-faint uppercase">
                        Closed
                      </span>
                    ) : null}
                  </div>
                  <div className="mt-2 flex items-center gap-2 text-[11px] text-fg-faint">
                    <span className="inline-flex items-center gap-1">
                      <MessageSquare className="h-3 w-3" />
                      {q.answer_count}
                    </span>
                    <span>{relativeTime(q.created_at)}</span>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}
        {next ? (
          <div className="py-3">
            <button
              type="button"
              disabled={loadingMore}
              onClick={() => void loadMore()}
              className="w-full rounded-xl border border-border-subtle bg-surface-2 py-2 text-sm font-semibold text-fg hover:bg-overlay disabled:opacity-60"
            >
              {loadingMore ? "Loading…" : "Load more"}
            </button>
          </div>
        ) : null}
      </div>

      <div className="shrink-0 border-t border-border-subtle px-3 py-3">
        <Link
          href={`/qa/ask?post_id=${encodeURIComponent(postId)}`}
          className="inline-flex h-10 w-full items-center justify-center rounded-xl bg-brand text-sm font-semibold text-black hover:bg-brand-dark"
        >
          Ask about this place
        </Link>
      </div>
    </div>
  );
}

export default function PostQuestionsPanel({
  postId,
  mobileOpen,
  onMobileClose,
  desktopOpen,
}: {
  postId: string;
  lat?: number;
  lng?: number;
  mobileOpen: boolean;
  onMobileClose: () => void;
  desktopOpen: boolean;
}) {
  useEffect(() => {
    if (!mobileOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onMobileClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [mobileOpen, onMobileClose]);

  if (!postId) return null;

  return (
    <>
      {desktopOpen ? (
        <aside
          className="sticky top-24 hidden h-[calc(100dvh-6.5rem)] w-[28rem] min-w-[28rem] max-w-[28rem] shrink-0 flex-col overflow-hidden rounded-2xl border border-border-subtle lg:flex"
          aria-label="Place Q&A"
        >
          <PanelChrome className="h-full min-h-0 w-full min-w-0 overflow-hidden">
            <QuestionsBody key={postId} postId={postId} />
          </PanelChrome>
        </aside>
      ) : null}

      {mobileOpen ? (
        <div
          className="fixed inset-x-0 bottom-0 z-40 flex h-[min(52vh,28rem)] flex-col overflow-hidden rounded-t-2xl border border-b-0 border-border-subtle shadow-[0_-12px_40px_-20px_rgba(0,0,0,0.45)] lg:hidden"
          role="dialog"
          aria-modal="false"
          aria-label="Place Q&A"
        >
          <PanelChrome onClose={onMobileClose} grabber showClose className="h-full">
            <QuestionsBody key={postId} postId={postId} />
          </PanelChrome>
        </div>
      ) : null}
    </>
  );
}
