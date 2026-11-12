"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { Loader2, MessageCircle, ThumbsDown, ThumbsUp } from "lucide-react";
import { toast } from "sonner";
import UserAvatar from "@/components/shared/UserAvatar";
import {
  useAddCommentMutation,
  useLazyListCommentRepliesQuery,
  useLazyListCommentsQuery,
  useListCommentsQuery,
  useSetCommentReactionMutation,
} from "@/lib/api/contentApi";
import { useRequireSignIn } from "@/hooks/use-require-signin";
import { useGetUserProfileQuery } from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { relativeTime } from "@/lib/content/helpers";
import type { Comment, Reaction } from "@/lib/content/types";
import { useAppSelector } from "@/lib/hooks";
import { cn } from "@/lib/utils";

function CommentComposer({
  placeholder,
  onSubmit,
  posting,
  autoFocus,
  compact,
}: {
  placeholder: string;
  onSubmit: (text: string) => Promise<void>;
  posting: boolean;
  autoFocus?: boolean;
  compact?: boolean;
}) {
  const { requireSignIn } = useRequireSignIn();
  const [body, setBody] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!requireSignIn("Sign in to comment")) return;
    const text = body.trim();
    if (!text) return;
    await onSubmit(text);
    setBody("");
  }

  return (
    <form
      onSubmit={(e) => void handleSubmit(e)}
      className={cn(
        "flex shrink-0 items-center gap-2",
        compact ? "px-0 py-2" : "border-t border-border-subtle bg-surface px-4 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]"
      )}
    >
      <input
        value={body}
        onChange={(e) => setBody(e.target.value)}
        maxLength={2000}
        placeholder={placeholder}
        autoFocus={autoFocus}
        className="min-w-0 flex-1 rounded-full border border-border-subtle bg-surface-2 px-4 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/40"
      />
      <button
        type="submit"
        disabled={posting || !body.trim()}
        className="shrink-0 px-2 text-sm font-semibold text-brand disabled:opacity-40"
      >
        Post
      </button>
    </form>
  );
}

function CommentRow({
  comment,
  postId,
  isReply,
}: {
  comment: Comment;
  postId: string;
  isReply?: boolean;
}) {
  const { data: author } = useGetUserProfileQuery(comment.author_id);
  const viewerId = useAppSelector((s) => s.auth.userId);
  const { requireSignIn } = useRequireSignIn();
  const isOwn = Boolean(viewerId && viewerId === comment.author_id);
  const [setReaction, { isLoading: reacting }] = useSetCommentReactionMutation();
  const [addComment, { isLoading: postingReply }] = useAddCommentMutation();
  const [fetchReplies] = useLazyListCommentRepliesQuery();

  const [likes, setLikes] = useState(comment.like_count);
  const [dislikes, setDislikes] = useState(comment.dislike_count);
  const [replyCount, setReplyCount] = useState(comment.reply_count);
  const [reaction, setReactionLocal] = useState<Reaction | null>(null);
  const [replyOpen, setReplyOpen] = useState(false);
  const [repliesOpen, setRepliesOpen] = useState(false);
  const [replies, setReplies] = useState<Comment[]>([]);
  const [repliesCursor, setRepliesCursor] = useState<string | null | undefined>(
    undefined
  );
  const [loadingReplies, setLoadingReplies] = useState(false);

  async function onReact(next: Reaction) {
    if (!requireSignIn("Sign in to react")) return;
    if (isOwn) return;
    const prev = reaction;
    const nextReaction = prev === next ? null : next;
    setReactionLocal(nextReaction);
    setLikes((n) => n + (nextReaction === "like" ? 1 : 0) - (prev === "like" ? 1 : 0));
    setDislikes(
      (n) => n + (nextReaction === "dislike" ? 1 : 0) - (prev === "dislike" ? 1 : 0)
    );
    try {
      const updated = await setReaction({
        commentId: comment.id,
        reaction: nextReaction,
        postId,
        parentId: comment.parent_id,
      }).unwrap();
      setLikes(updated.like_count);
      setDislikes(updated.dislike_count);
    } catch (err) {
      setReactionLocal(prev);
      setLikes(comment.like_count);
      setDislikes(comment.dislike_count);
      toast.error("Could not react", { description: getApiErrorMessage(err) });
    }
  }

  async function loadReplies(reset = false) {
    if (loadingReplies) return;
    setLoadingReplies(true);
    try {
      const page = await fetchReplies({
        commentId: comment.id,
        limit: 20,
        ...(!reset && repliesCursor ? { cursor: repliesCursor } : {}),
      }).unwrap();
      setReplies((prev) => (reset ? page.comments : [...prev, ...page.comments]));
      setRepliesCursor(page.next_cursor ?? null);
      setRepliesOpen(true);
    } catch (err) {
      toast.error("Could not load replies", {
        description: getApiErrorMessage(err),
      });
    } finally {
      setLoadingReplies(false);
    }
  }

  async function onReply(text: string) {
    try {
      const created = await addComment({
        postId,
        body: text,
        parentId: comment.id,
      }).unwrap();
      setReplies((prev) => [created, ...prev]);
      setReplyCount((n) => n + 1);
      setRepliesOpen(true);
      setReplyOpen(false);
    } catch (err) {
      toast.error("Could not reply", { description: getApiErrorMessage(err) });
    }
  }

  return (
    <li className={cn("px-4 py-3", isReply && "pl-12")}>
      <div className="flex gap-3">
        <Link href={`/users/${comment.author_id}`} className="shrink-0 self-start">
          <UserAvatar
            user={
              author ?? {
                full_name: "",
                username: "…",
                avatar_url: "",
              }
            }
            size="xs"
            className="!h-8 !w-8 !rounded-full !text-[10px]"
          />
        </Link>
        <div className="min-w-0 flex-1 overflow-hidden">
          <p className="text-sm leading-snug text-fg">
            <Link
              href={`/users/${comment.author_id}`}
              className="mr-1.5 font-semibold hover:underline"
            >
              {author?.username ?? "user"}
            </Link>
            <span className="break-words whitespace-pre-wrap [overflow-wrap:anywhere]">
              {comment.body}
            </span>
          </p>
          <p className="mt-1 text-[11px] text-fg-faint">
            {relativeTime(comment.created_at)}
          </p>

          <div className="mt-2 flex flex-wrap items-center gap-1">
            <button
              type="button"
              disabled={isOwn || reacting}
              title={isOwn ? "You can’t react to your own comment" : undefined}
              onClick={() => void onReact("like")}
              className={cn(
                "inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-xs disabled:opacity-60",
                isOwn && "cursor-not-allowed",
                reaction === "like" ? "text-brand" : "text-fg-muted hover:text-fg"
              )}
            >
              <ThumbsUp className={cn("h-3.5 w-3.5", reaction === "like" && "fill-current")} />
              <span className="tabular-nums">{likes}</span>
            </button>
            <button
              type="button"
              disabled={isOwn || reacting}
              title={isOwn ? "You can’t react to your own comment" : undefined}
              onClick={() => void onReact("dislike")}
              className={cn(
                "inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-xs disabled:opacity-60",
                isOwn && "cursor-not-allowed",
                reaction === "dislike" ? "text-red-400" : "text-fg-muted hover:text-fg"
              )}
            >
              <ThumbsDown
                className={cn("h-3.5 w-3.5", reaction === "dislike" && "fill-current")}
              />
              <span className="tabular-nums">{dislikes}</span>
            </button>
            {!isReply ? (
              <button
                type="button"
                onClick={() => setReplyOpen((v) => !v)}
                className="inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-xs text-fg-muted hover:text-fg"
              >
                <MessageCircle className="h-3.5 w-3.5" />
                Reply
              </button>
            ) : null}
            {!isReply && replyCount > 0 ? (
              <button
                type="button"
                onClick={() => {
                  if (repliesOpen) {
                    setRepliesOpen(false);
                    return;
                  }
                  if (replies.length > 0) setRepliesOpen(true);
                  else void loadReplies(true);
                }}
                className="rounded-md px-1.5 py-1 text-xs font-medium text-brand hover:underline"
              >
                {repliesOpen
                  ? "Hide replies"
                  : `View ${replyCount} ${replyCount === 1 ? "reply" : "replies"}`}
              </button>
            ) : null}
          </div>

          {replyOpen && !isReply ? (
            <div className="mt-2">
              <CommentComposer
                placeholder="Write a reply…"
                onSubmit={onReply}
                posting={postingReply}
                autoFocus
                compact
              />
            </div>
          ) : null}
        </div>
      </div>

      {repliesOpen && replies.length > 0 ? (
        <ul className="mt-1 divide-y divide-border-subtle border-l border-border-subtle ml-4">
          {replies.map((r) => (
            <CommentRow key={r.id} comment={r} postId={postId} isReply />
          ))}
        </ul>
      ) : null}

      {repliesOpen && repliesCursor ? (
        <button
          type="button"
          disabled={loadingReplies}
          onClick={() => void loadReplies(false)}
          className="ml-12 py-2 text-xs font-medium text-fg-muted hover:text-fg disabled:opacity-60"
        >
          {loadingReplies ? "Loading…" : "More replies"}
        </button>
      ) : null}
    </li>
  );
}

/** Scrollable comments thread + composer. Fills a flex parent. */
export default function CommentsView({
  postId,
  className,
}: {
  postId: string;
  className?: string;
}) {
  const { data, isLoading, isError, refetch } = useListCommentsQuery({
    postId,
    limit: 20,
  });
  const [fetchMore] = useLazyListCommentsQuery();
  const [extra, setExtra] = useState<Comment[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);
  const [addComment, { isLoading: posting }] = useAddCommentMutation();

  const first = data?.comments ?? [];
  const comments =
    extra.length === 0
      ? first
      : (() => {
          const seen = new Set(first.map((c) => c.id));
          return [...first, ...extra.filter((c) => !seen.has(c.id))];
        })();
  const next =
    cursor === null ? undefined : cursor === undefined ? data?.next_cursor : cursor;

  async function onSubmit(text: string) {
    try {
      const created = await addComment({ postId, body: text }).unwrap();
      setExtra((prev) => [created, ...prev]);
    } catch (err) {
      toast.error("Could not comment", { description: getApiErrorMessage(err) });
    }
  }

  async function loadMore() {
    if (!next || loadingMore) return;
    setLoadingMore(true);
    try {
      const page = await fetchMore({ postId, limit: 20, cursor: next }).unwrap();
      setExtra((prev) => [...prev, ...page.comments]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }

  return (
    <div className={cn("flex min-h-0 min-w-0 w-full flex-1 flex-col overflow-hidden", className)}>
      <div className="yb-scrollbar min-h-0 min-w-0 flex-1 overflow-x-hidden overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center gap-2 py-16 text-sm text-fg-faint">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading…
          </div>
        ) : isError ? (
          <div className="px-4 py-16 text-center">
            <button
              type="button"
              onClick={() => void refetch()}
              className="text-sm font-medium text-brand hover:underline"
            >
              Couldn’t load comments — retry
            </button>
          </div>
        ) : comments.length === 0 ? (
          <p className="px-4 py-16 text-center text-sm text-fg-faint">No comments yet.</p>
        ) : (
          <ul className="divide-y divide-border-subtle">
            {comments.map((c) => (
              <CommentRow key={c.id} comment={c} postId={postId} />
            ))}
          </ul>
        )}

        {next ? (
          <button
            type="button"
            onClick={() => void loadMore()}
            disabled={loadingMore}
            className="w-full py-4 text-sm font-medium text-fg-muted hover:text-fg disabled:opacity-60"
          >
            {loadingMore ? "Loading…" : "View more"}
          </button>
        ) : null}
      </div>

      <CommentComposer
        placeholder="Add a comment…"
        onSubmit={onSubmit}
        posting={posting}
      />
    </div>
  );
}
