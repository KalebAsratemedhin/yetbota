"use client";

import { use, useCallback, useState } from "react";
import Link from "next/link";
import { ArrowLeft, ChevronDown, ChevronUp, Lock } from "lucide-react";
import { toast } from "sonner";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import UserAvatar from "@/components/shared/UserAvatar";
import {
  useCloseQuestionMutation,
  useCreateAnswerMutation,
  useGetQuestionQuery,
  useGetPostQuery,
  useLazyListAnswersQuery,
  useListAnswersQuery,
  useUpdateAnswerMutation,
  useVoteAnswerMutation,
  useVoteQuestionMutation,
} from "@/lib/api/contentApi";
import { useGetUserProfileQuery } from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { placeTitle, relativeTime } from "@/lib/content/helpers";
import type { Answer, QaVote, Question } from "@/lib/content/types";
import { useAppSelector } from "@/lib/hooks";
import { cn } from "@/lib/utils";

function AuthorLine({ userId }: { userId: string }) {
  const { data } = useGetUserProfileQuery(userId);
  return (
    <Link
      href={`/users/${userId}`}
      className="inline-flex items-center gap-1.5 hover:opacity-90"
    >
      <UserAvatar
        user={data ?? { full_name: "", username: "…", avatar_url: "" }}
        size="xs"
        className="!h-6 !w-6 !text-[9px]"
      />
      <span className="text-xs font-medium text-fg-muted">
        @{data?.username ?? "…"}
      </span>
    </Link>
  );
}

function VoteColumn({
  score,
  onUp,
  onDown,
  myVote,
  disabled,
  disabledReason,
}: {
  score: number;
  onUp: () => void;
  onDown: () => void;
  myVote: QaVote | null;
  disabled?: boolean;
  disabledReason?: string;
}) {
  return (
    <div className="flex w-10 shrink-0 flex-col items-center gap-0.5">
      <button
        type="button"
        disabled={disabled}
        title={disabled ? disabledReason : undefined}
        onClick={onUp}
        className={cn(
          "rounded-lg p-1 transition-colors disabled:cursor-not-allowed disabled:opacity-40",
          myVote === "up"
            ? "bg-brand/15 text-brand"
            : "text-fg-faint hover:bg-overlay hover:text-fg"
        )}
        aria-label="Upvote"
      >
        <ChevronUp className="h-5 w-5" />
      </button>
      <span className="text-sm font-bold tabular-nums text-fg">{score}</span>
      <button
        type="button"
        disabled={disabled}
        title={disabled ? disabledReason : undefined}
        onClick={onDown}
        className={cn(
          "rounded-lg p-1 transition-colors disabled:cursor-not-allowed disabled:opacity-40",
          myVote === "down"
            ? "bg-rose-500/15 text-rose-400"
            : "text-fg-faint hover:bg-overlay hover:text-fg"
        )}
        aria-label="Downvote"
      >
        <ChevronDown className="h-5 w-5" />
      </button>
    </div>
  );
}

function useLocalVote() {
  const [myVote, setMyVote] = useState<QaVote | null>(null);
  const nextVote = useCallback(
    (clicked: QaVote): QaVote | null => (myVote === clicked ? null : clicked),
    [myVote]
  );
  return { myVote, setMyVote, nextVote };
}

function QuestionBlock({ question }: { question: Question }) {
  const viewerId = useAppSelector((s) => s.auth.userId);
  const isAuthor = Boolean(viewerId && viewerId === question.author_id);
  const closed = Boolean(question.closed_at);
  const [voteQuestion] = useVoteQuestionMutation();
  const [closeQuestion, { isLoading: closing }] = useCloseQuestionMutation();
  const { myVote, setMyVote, nextVote } = useLocalVote();
  const postQ = useGetPostQuery(question.post_id ?? "", {
    skip: !question.post_id,
  });

  async function vote(clicked: QaVote) {
    if (isAuthor) return;
    const prev = myVote;
    const vote = nextVote(clicked);
    setMyVote(vote);
    try {
      await voteQuestion({ id: question.id, vote }).unwrap();
    } catch (err) {
      setMyVote(prev);
      toast.error("Could not vote", { description: getApiErrorMessage(err) });
    }
  }

  async function onClose() {
    try {
      await closeQuestion(question.id).unwrap();
      toast.success("Question closed");
    } catch (err) {
      toast.error("Could not close", { description: getApiErrorMessage(err) });
    }
  }

  return (
    <article className="rounded-2xl border border-border-subtle bg-surface p-5 sm:p-6">
      <div className="flex gap-4">
        <VoteColumn
          score={question.upvote_count - question.downvote_count}
          myVote={myVote}
          disabled={isAuthor}
          disabledReason="You can’t vote on your own question"
          onUp={() => void vote("up")}
          onDown={() => void vote("down")}
        />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-start gap-2">
            <p className="min-w-0 flex-1 whitespace-pre-wrap text-sm leading-relaxed text-fg sm:text-[15px]">
              {question.body}
            </p>
            {closed ? (
              <span className="inline-flex shrink-0 items-center gap-1 rounded bg-fg-faint/15 px-1.5 py-0.5 text-[10px] font-bold tracking-wide text-fg-muted uppercase">
                <Lock className="h-3 w-3" />
                Closed
              </span>
            ) : null}
          </div>
          <div className="mt-4 flex flex-wrap items-center gap-3 text-xs text-fg-faint">
            <AuthorLine userId={question.author_id} />
            <span>{relativeTime(question.created_at)}</span>
            {question.post_id && postQ.data ? (
              <Link
                href={`/posts/${question.post_id}`}
                className="font-semibold text-brand hover:underline"
              >
                {placeTitle(postQ.data)}
              </Link>
            ) : null}
          </div>
          {isAuthor && !closed ? (
            <div className="mt-4 flex flex-wrap items-center gap-4">
              <Link
                href={`/qa/ask?edit=${question.id}`}
                className="text-xs font-semibold text-fg-muted hover:text-fg"
              >
                Edit
              </Link>
              <button
                type="button"
                disabled={closing}
                onClick={() => void onClose()}
                className="text-xs font-semibold text-fg-muted hover:text-fg disabled:opacity-50"
              >
                Close question
              </button>
            </div>
          ) : null}
        </div>
      </div>
    </article>
  );
}

function AnswerRow({
  answer,
  questionId,
}: {
  answer: Answer;
  questionId: string;
}) {
  const viewerId = useAppSelector((s) => s.auth.userId);
  const isAuthor = Boolean(viewerId && viewerId === answer.author_id);
  const [voteAnswer] = useVoteAnswerMutation();
  const [updateAnswer, { isLoading: saving }] = useUpdateAnswerMutation();
  const { myVote, setMyVote, nextVote } = useLocalVote();
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(answer.body);

  async function vote(clicked: QaVote) {
    if (isAuthor) return;
    const prev = myVote;
    const vote = nextVote(clicked);
    setMyVote(vote);
    try {
      await voteAnswer({ id: answer.id, questionId, vote }).unwrap();
    } catch (err) {
      setMyVote(prev);
      toast.error("Could not vote", { description: getApiErrorMessage(err) });
    }
  }

  function startEdit() {
    setDraft(answer.body);
    setEditing(true);
  }

  function cancelEdit() {
    setDraft(answer.body);
    setEditing(false);
  }

  async function saveEdit() {
    const text = draft.trim();
    if (!text) {
      toast.error("Answer can’t be empty");
      return;
    }
    try {
      await updateAnswer({
        id: answer.id,
        questionId,
        body: { body: text.slice(0, 10_000) },
      }).unwrap();
      setEditing(false);
      toast.success("Answer updated");
    } catch (err) {
      toast.error("Could not update answer", {
        description: getApiErrorMessage(err),
      });
    }
  }

  return (
    <li className="flex gap-4 rounded-2xl border border-border-subtle bg-surface px-4 py-4 sm:px-5">
      <VoteColumn
        score={answer.upvote_count - answer.downvote_count}
        myVote={myVote}
        disabled={isAuthor}
        disabledReason="You can’t vote on your own answer"
        onUp={() => void vote("up")}
        onDown={() => void vote("down")}
      />
      <div className="min-w-0 flex-1">
        {editing ? (
          <div className="space-y-3">
            <textarea
              rows={4}
              value={draft}
              onChange={(e) => setDraft(e.target.value.slice(0, 10_000))}
              className="yb-scrollbar w-full resize-y rounded-xl border border-border-subtle bg-bg px-3 py-2.5 text-sm text-fg outline-none focus:border-brand/50"
            />
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                disabled={saving || !draft.trim()}
                onClick={() => void saveEdit()}
                className="inline-flex h-9 items-center rounded-xl bg-brand px-4 text-sm font-semibold text-black hover:bg-brand-dark disabled:opacity-40"
              >
                {saving ? "Saving…" : "Save"}
              </button>
              <button
                type="button"
                disabled={saving}
                onClick={cancelEdit}
                className="inline-flex h-9 items-center rounded-xl border border-border-subtle px-4 text-sm font-semibold text-fg-muted hover:text-fg disabled:opacity-50"
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <p className="whitespace-pre-wrap text-sm leading-relaxed text-fg">
            {answer.body}
          </p>
        )}
        <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-fg-faint">
          <AuthorLine userId={answer.author_id} />
          <span>{relativeTime(answer.created_at)}</span>
          {isAuthor && !editing ? (
            <button
              type="button"
              onClick={startEdit}
              className="font-semibold text-fg-muted hover:text-fg"
            >
              Edit
            </button>
          ) : null}
        </div>
      </div>
    </li>
  );
}

function QuestionDetail({ questionId }: { questionId: string }) {
  const q = useGetQuestionQuery(questionId);
  const answersQ = useListAnswersQuery({ questionId, limit: 20 });
  const [fetchMore] = useLazyListAnswersQuery();
  const [createAnswer, { isLoading: posting }] = useCreateAnswerMutation();
  const [extra, setExtra] = useState<Answer[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);
  const [body, setBody] = useState("");

  const first = answersQ.data?.answers ?? [];
  const answers =
    extra.length === 0
      ? first
      : (() => {
          const seen = new Set(first.map((a) => a.id));
          return [...first, ...extra.filter((a) => !seen.has(a.id))];
        })();

  const next =
    cursor === null
      ? undefined
      : cursor === undefined
        ? answersQ.data?.next_cursor
        : cursor;

  const loadMore = useCallback(async () => {
    if (!next || loadingMore) return;
    setLoadingMore(true);
    try {
      const page = await fetchMore({
        questionId,
        limit: 20,
        cursor: next,
      }).unwrap();
      setExtra((prev) => [...prev, ...page.answers]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }, [next, loadingMore, fetchMore, questionId]);

  async function onAnswer(e: React.FormEvent) {
    e.preventDefault();
    const text = body.trim();
    if (!text) return;
    try {
      await createAnswer({ questionId, body: text.slice(0, 10_000) }).unwrap();
      setBody("");
      toast.success("Answer posted");
    } catch (err) {
      toast.error("Could not post answer", {
        description: getApiErrorMessage(err),
      });
    }
  }

  if (q.isLoading) {
    return (
      <p className="py-20 text-center text-sm text-fg-faint">Loading…</p>
    );
  }

  if (q.isError || !q.data) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-20 text-center">
        <p className="text-sm text-fg-muted">
          {getApiErrorMessage(q.error) || "Question not found."}
        </p>
        <Link href="/qa" className="mt-3 inline-block text-sm font-semibold text-brand hover:underline">
          Back to Q&A
        </Link>
      </div>
    );
  }

  const closed = Boolean(q.data.closed_at);

  return (
    <div className="mx-auto w-full max-w-3xl px-4 pt-8 pb-16 sm:px-6 sm:pt-10">
      <Link
        href="/qa"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-fg-muted hover:text-fg"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Q&A
      </Link>

      <QuestionBlock question={q.data} />

      <section className="mt-10">
        <h2 className="mb-4 text-sm font-bold tracking-wide text-fg-faint uppercase">
          {q.data.answer_count}{" "}
          {q.data.answer_count === 1 ? "Answer" : "Answers"}
        </h2>

        {answersQ.isLoading ? (
          <p className="py-8 text-sm text-fg-faint">Loading answers…</p>
        ) : answers.length === 0 ? (
          <p className="rounded-xl border border-dashed border-border-subtle px-4 py-8 text-center text-sm text-fg-faint">
            No answers yet.
          </p>
        ) : (
          <ul className="space-y-3">
            {answers.map((a) => (
              <AnswerRow key={a.id} answer={a} questionId={questionId} />
            ))}
          </ul>
        )}

        {next ? (
          <div className="mt-4 flex justify-center">
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
      </section>

      <section className="mt-10">
        <h2 className="mb-3 text-sm font-bold tracking-wide text-fg-faint uppercase">
          Your answer
        </h2>
        {closed ? (
          <p className="rounded-xl border border-border-subtle bg-surface px-4 py-6 text-sm text-fg-muted">
            This question is closed — new answers aren’t accepted.
          </p>
        ) : (
          <form onSubmit={onAnswer} className="space-y-3">
            <textarea
              rows={5}
              value={body}
              onChange={(e) => setBody(e.target.value.slice(0, 10_000))}
              placeholder="Share what you know…"
              className="yb-scrollbar w-full resize-y rounded-xl border border-border-subtle bg-surface px-3 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
            />
            <button
              type="submit"
              disabled={posting || !body.trim()}
              className="inline-flex h-10 items-center rounded-xl bg-brand px-5 text-sm font-semibold text-black hover:bg-brand-dark disabled:opacity-40"
            >
              {posting ? "Posting…" : "Post answer"}
            </button>
          </form>
        )}
      </section>
    </div>
  );
}

export default function QuestionPage({
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
          <QuestionDetail questionId={id} />
        </div>
      </main>
    </RequireAuth>
  );
}
