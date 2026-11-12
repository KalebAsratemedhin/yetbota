"use client";

import Link from "next/link";
import { HelpCircle, MessageSquare } from "lucide-react";
import Reveal from "@/components/landing/Reveal";
import SectionShell from "@/components/landing/SectionShell";
import { useListQuestionsQuery } from "@/lib/api/contentApi";
import { useGetUserProfileQuery } from "@/lib/api/socialApi";
import { relativeTime } from "@/lib/content/helpers";
import type { Question } from "@/lib/content/types";
import { useContent } from "@/lib/useContent";

function AuthorName({ userId }: { userId: string }) {
  const { data } = useGetUserProfileQuery(userId);
  return <span>@{data?.username ?? "…"}</span>;
}

function QuestionPreview({ q }: { q: Question }) {
  const score = q.upvote_count - q.downvote_count;
  return (
    <Link
      href={`/qa/${q.id}`}
      className="block rounded-3xl border border-border-subtle bg-bg px-5 py-6 transition-colors hover:border-brand/35 hover:bg-surface-2 sm:px-6 sm:py-7"
    >
      <div className="flex gap-5">
        <div className="flex w-14 shrink-0 flex-col items-center justify-center rounded-2xl bg-overlay py-3">
          <span className="text-lg font-bold tabular-nums text-fg">{score}</span>
          <span className="text-[10px] text-fg-faint">votes</span>
        </div>
        <div className="min-w-0 flex-1">
          <p className="line-clamp-3 text-base leading-relaxed text-fg">{q.body}</p>
          <div className="mt-4 flex flex-wrap items-center gap-3 text-sm text-fg-faint">
            <AuthorName userId={q.author_id} />
            <span>{relativeTime(q.created_at)}</span>
            <span className="inline-flex items-center gap-1.5">
              <MessageSquare className="h-3.5 w-3.5" aria-hidden />
              {q.answer_count}
            </span>
          </div>
        </div>
      </div>
    </Link>
  );
}

export default function CommunityQaSection() {
  const t = useContent();
  const listQ = useListQuestionsQuery({ limit: 3 });
  const questions = listQ.data?.questions ?? [];

  return (
    <SectionShell className="bg-brand/[0.04] dark:bg-surface">
      <Reveal className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div className="max-w-2xl">
          <p className="mb-2 text-[11px] font-bold tracking-[0.18em] text-brand uppercase">
            Local Q&A
          </p>
          <h2 className="text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
            {t.community.helpTitle}
          </h2>
          <p className="mt-2 text-sm leading-relaxed text-fg-muted sm:text-base">
            Real questions from the community — parking tips, what’s open, what’s changed.
          </p>
        </div>
        <Link href="/qa" className="text-sm font-semibold text-brand hover:underline">
          {t.community.browseQuestions} →
        </Link>
      </Reveal>

      {listQ.isLoading ? (
        <div className="grid gap-5 lg:grid-cols-1">
          {Array.from({ length: 3 }, (_, i) => (
            <div
              key={i}
              className="h-36 animate-pulse rounded-3xl bg-overlay"
              aria-hidden
            />
          ))}
        </div>
      ) : listQ.isError ? (
        <p className="text-base text-fg-muted">Couldn’t load questions right now.</p>
      ) : questions.length === 0 ? (
        <div className="flex flex-col items-center gap-4 rounded-3xl border border-dashed border-border-subtle px-6 py-16 text-center">
          <HelpCircle className="h-10 w-10 text-fg-faint" />
          <p className="text-base font-semibold text-fg">No questions yet</p>
          <Link href="/qa/ask" className="text-sm font-semibold text-brand hover:underline">
            Ask the first one
          </Link>
        </div>
      ) : (
        <ul className="mx-auto grid max-w-3xl gap-5">
          {questions.map((q, i) => (
            <Reveal key={q.id} delay={i * 70}>
              <li>
                <QuestionPreview q={q} />
              </li>
            </Reveal>
          ))}
        </ul>
      )}
    </SectionShell>
  );
}
