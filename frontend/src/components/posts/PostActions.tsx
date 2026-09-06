"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Bookmark,
  BookmarkCheck,
  CircleHelp,
  MessageCircle,
  Share2,
  ThumbsDown,
  ThumbsUp,
} from "lucide-react";
import { toast } from "sonner";
import {
  useAddBookmarkMutation,
  useRemoveBookmarkMutation,
  useSetReactionMutation,
  useSharePostMutation,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { useAppSelector } from "@/lib/hooks";
import type { Post, Reaction } from "@/lib/content/types";
import { cn } from "@/lib/utils";

export default function PostActions({
  post,
  initialReaction = null,
  initialBookmarked = false,
  compact,
  commentsHref,
  onCommentsClick,
  onQaClick,
}: {
  post: Post;
  initialReaction?: Reaction | null;
  initialBookmarked?: boolean;
  compact?: boolean;
  /** Link fallback when `onCommentsClick` is not provided. */
  commentsHref?: string;
  /** Opens the in-page comments panel (preferred on post detail). */
  onCommentsClick?: () => void;
  /** Opens the place Q&A panel. */
  onQaClick?: () => void;
}) {
  const viewerId = useAppSelector((s) => s.auth.userId);
  const isOwn = Boolean(viewerId && viewerId === post.author_id);
  const commentsTo = commentsHref ?? `/posts/${post.id}?comments=1`;
  const [reaction, setReactionLocal] = useState<Reaction | null>(initialReaction);
  const [bookmarked, setBookmarked] = useState(initialBookmarked);
  const [counts, setCounts] = useState({
    like: post.like_count,
    dislike: post.dislike_count,
    bookmark: post.bookmark_count,
    share: post.share_count,
    comment: post.comment_count,
  });

  const [setReaction, { isLoading: reacting }] = useSetReactionMutation();
  const [addBookmark, { isLoading: bookmarking }] = useAddBookmarkMutation();
  const [removeBookmark, { isLoading: unbookmarking }] = useRemoveBookmarkMutation();
  const [sharePost, { isLoading: sharing }] = useSharePostMutation();

  async function onReact(next: Reaction) {
    if (isOwn) return;
    const prev = reaction;
    const clearing = prev === next;
    const applied: Reaction | null = clearing ? null : next;
    setReactionLocal(applied);
    setCounts((c) => {
      let like = c.like;
      let dislike = c.dislike;
      if (prev === "like") like -= 1;
      if (prev === "dislike") dislike -= 1;
      if (applied === "like") like += 1;
      if (applied === "dislike") dislike += 1;
      return { ...c, like: Math.max(0, like), dislike: Math.max(0, dislike) };
    });
    try {
      const updated = await setReaction({ postId: post.id, reaction: applied }).unwrap();
      setCounts((c) => ({
        ...c,
        like: updated.like_count,
        dislike: updated.dislike_count,
      }));
    } catch (err) {
      setReactionLocal(prev);
      setCounts((c) => ({
        ...c,
        like: post.like_count,
        dislike: post.dislike_count,
      }));
      toast.error("Could not update reaction", { description: getApiErrorMessage(err) });
    }
  }

  async function onBookmark() {
    const prev = bookmarked;
    setBookmarked(!prev);
    setCounts((c) => ({
      ...c,
      bookmark: Math.max(0, c.bookmark + (prev ? -1 : 1)),
    }));
    try {
      const updated = prev
        ? await removeBookmark(post.id).unwrap()
        : await addBookmark(post.id).unwrap();
      setCounts((c) => ({ ...c, bookmark: updated.bookmark_count }));
      toast.success(prev ? "Removed from saved" : "Saved");
    } catch (err) {
      setBookmarked(prev);
      toast.error("Could not update bookmark", { description: getApiErrorMessage(err) });
    }
  }

  async function onShare() {
    try {
      const res = await sharePost(post.id).unwrap();
      setCounts((c) => ({ ...c, share: res.post.share_count }));
      await navigator.clipboard.writeText(res.share_url);
      toast.success("Link copied");
    } catch (err) {
      toast.error("Could not share", { description: getApiErrorMessage(err) });
    }
  }

  if (compact) {
    const btn = "h-8 gap-1 px-2 text-[11px]";
    return (
      <div className="flex flex-wrap items-center gap-1.5">
        <button
          type="button"
          disabled={isOwn || reacting}
          title={isOwn ? "You can’t react to your own post" : undefined}
          onClick={() => void onReact("like")}
          className={cn(
            "inline-flex items-center rounded-full border font-semibold transition-colors disabled:opacity-60",
            btn,
            isOwn && "cursor-not-allowed",
            reaction === "like"
              ? "border-brand/40 bg-brand/15 text-brand"
              : "border-border-subtle bg-surface text-fg-muted hover:text-fg"
          )}
        >
          <ThumbsUp className="h-3.5 w-3.5" />
          {counts.like}
        </button>
        <button
          type="button"
          disabled={isOwn || reacting}
          title={isOwn ? "You can’t react to your own post" : undefined}
          onClick={() => void onReact("dislike")}
          className={cn(
            "inline-flex items-center rounded-full border font-semibold transition-colors disabled:opacity-60",
            btn,
            isOwn && "cursor-not-allowed",
            reaction === "dislike"
              ? "border-red-400/40 bg-red-500/10 text-red-400"
              : "border-border-subtle bg-surface text-fg-muted hover:text-fg"
          )}
        >
          <ThumbsDown className="h-3.5 w-3.5" />
          {counts.dislike}
        </button>
        {onCommentsClick ? (
          <button
            type="button"
            onClick={onCommentsClick}
            className={cn(
              "inline-flex items-center rounded-full border border-border-subtle bg-surface font-semibold text-fg-muted transition-colors hover:text-fg",
              btn
            )}
          >
            <MessageCircle className="h-3.5 w-3.5" />
            {counts.comment}
          </button>
        ) : (
          <Link
            href={commentsTo}
            className={cn(
              "inline-flex items-center rounded-full border border-border-subtle bg-surface font-semibold text-fg-muted transition-colors hover:text-fg",
              btn
            )}
          >
            <MessageCircle className="h-3.5 w-3.5" />
            {counts.comment}
          </Link>
        )}
        <button
          type="button"
          disabled={bookmarking || unbookmarking}
          onClick={() => void onBookmark()}
          className={cn(
            "inline-flex items-center rounded-full border font-semibold transition-colors disabled:opacity-60",
            btn,
            bookmarked
              ? "border-brand/40 bg-brand/15 text-brand"
              : "border-border-subtle bg-surface text-fg-muted hover:text-fg"
          )}
        >
          {bookmarked ? (
            <BookmarkCheck className="h-3.5 w-3.5" />
          ) : (
            <Bookmark className="h-3.5 w-3.5" />
          )}
        </button>
        <button
          type="button"
          disabled={sharing}
          onClick={() => void onShare()}
          className={cn(
            "inline-flex items-center rounded-full border border-border-subtle bg-surface font-semibold text-fg-muted transition-colors hover:text-fg disabled:opacity-60",
            btn
          )}
        >
          <Share2 className="h-3.5 w-3.5" />
        </button>
        {onQaClick ? (
          <button
            type="button"
            onClick={onQaClick}
            className={cn(
              "inline-flex items-center rounded-full border border-border-subtle bg-surface font-semibold text-fg-muted transition-colors hover:text-fg",
              btn
            )}
            aria-label="Place Q&A"
            title="Place Q&A"
          >
            <CircleHelp className="h-3.5 w-3.5" />
          </button>
        ) : null}
      </div>
    );
  }

  // Detail / feed-style icon row
  return (
    <div className="flex items-center gap-1">
      <button
        type="button"
        disabled={isOwn || reacting}
        title={isOwn ? "You can’t react to your own post" : undefined}
        onClick={() => void onReact("like")}
        className={cn(
          "inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm transition-colors disabled:opacity-60",
          isOwn && "cursor-not-allowed",
          reaction === "like" ? "text-brand" : "text-fg hover:bg-surface-2"
        )}
        aria-label="Like"
      >
        <ThumbsUp className={cn("h-5 w-5", reaction === "like" && "fill-current")} />
        <span className="tabular-nums">{counts.like}</span>
      </button>
      <button
        type="button"
        disabled={isOwn || reacting}
        title={isOwn ? "You can’t react to your own post" : undefined}
        onClick={() => void onReact("dislike")}
        className={cn(
          "inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm transition-colors disabled:opacity-60",
          isOwn && "cursor-not-allowed",
          reaction === "dislike" ? "text-red-400" : "text-fg hover:bg-surface-2"
        )}
        aria-label="Dislike"
      >
        <ThumbsDown className={cn("h-5 w-5", reaction === "dislike" && "fill-current")} />
        <span className="tabular-nums">{counts.dislike}</span>
      </button>
      {onCommentsClick ? (
        <button
          type="button"
          onClick={onCommentsClick}
          className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm text-fg hover:bg-surface-2"
          aria-label="Comments"
        >
          <MessageCircle className="h-5 w-5" />
          <span className="tabular-nums">{counts.comment}</span>
        </button>
      ) : (
        <Link
          href={commentsTo}
          className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm text-fg hover:bg-surface-2"
          aria-label="Comments"
        >
          <MessageCircle className="h-5 w-5" />
          <span className="tabular-nums">{counts.comment}</span>
        </Link>
      )}
      <button
        type="button"
        disabled={sharing}
        onClick={() => void onShare()}
        className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm text-fg hover:bg-surface-2 disabled:opacity-60"
        aria-label="Share"
      >
        <Share2 className="h-5 w-5" />
      </button>
      {onQaClick ? (
        <button
          type="button"
          onClick={onQaClick}
          className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm text-fg hover:bg-surface-2"
          aria-label="Place Q&A"
          title="Questions about this place"
        >
          <CircleHelp className="h-5 w-5" />
        </button>
      ) : null}
      <button
        type="button"
        disabled={bookmarking || unbookmarking}
        onClick={() => void onBookmark()}
        className={cn(
          "ml-auto inline-flex items-center rounded-lg px-2.5 py-2 transition-colors disabled:opacity-60",
          bookmarked ? "text-brand" : "text-fg hover:bg-surface-2"
        )}
        aria-label={bookmarked ? "Remove bookmark" : "Save"}
      >
        {bookmarked ? (
          <BookmarkCheck className="h-5 w-5 fill-current" />
        ) : (
          <Bookmark className="h-5 w-5" />
        )}
      </button>
    </div>
  );
}
