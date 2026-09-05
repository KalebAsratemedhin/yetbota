"use client";

import { useEffect, useRef, useState, type MouseEvent } from "react";
import Link from "next/link";
import { MapPin, Plus, UserCheck } from "lucide-react";
import { toast } from "sonner";
import PostActions from "@/components/posts/PostActions";
import PostMediaGallery from "@/components/posts/PostMediaGallery";
import UserAvatar from "@/components/shared/UserAvatar";
import {
  useFollowUserMutation,
  useGetUserProfileQuery,
  useListFollowingQuery,
  useUnfollowUserMutation,
} from "@/lib/api/socialApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { placeTitle, relativeTime } from "@/lib/content/helpers";
import type { Post } from "@/lib/content/types";
import { useAppSelector } from "@/lib/hooks";
import { cn } from "@/lib/utils";

function AuthorHeader({ authorId, createdAt }: { authorId: string; createdAt: string }) {
  const viewerId = useAppSelector((s) => s.auth.userId);
  const isSelf = Boolean(viewerId && viewerId === authorId);
  const { data: author } = useGetUserProfileQuery(authorId);
  const { data: myFollowing } = useListFollowingQuery(
    { id: viewerId ?? "", limit: 100 },
    { skip: !viewerId || isSelf }
  );
  const [followOverride, setFollowOverride] = useState<boolean | null>(null);
  const [busy, setBusy] = useState(false);
  const [trackedAuthor, setTrackedAuthor] = useState(authorId);

  if (authorId !== trackedAuthor) {
    setTrackedAuthor(authorId);
    setFollowOverride(null);
  }

  const [followUser] = useFollowUserMutation();
  const [unfollowUser] = useUnfollowUserMutation();

  const fromApi = Boolean(
    myFollowing?.users.some((u) => u.user_id === authorId)
  );
  const isFollowing = followOverride ?? fromApi;
  const level = author?.reputation_level?.trim() || "Explorer";
  const username = author?.username ?? "Loading…";

  async function toggleFollow(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (!viewerId || isSelf || busy || !author) return;
    setBusy(true);
    setFollowOverride(!isFollowing);
    try {
      if (isFollowing) {
        await unfollowUser(authorId).unwrap();
        toast.success("Unfollowed", { description: `@${author.username}` });
      } else {
        await followUser(authorId).unwrap();
        toast.success("Following", { description: `@${author.username}` });
      }
    } catch (err) {
      setFollowOverride(null);
      toast.error(isFollowing ? "Could not unfollow" : "Could not follow", {
        description: getApiErrorMessage(err),
      });
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex items-center gap-3 px-4 py-3 sm:px-5">
      <Link href={`/users/${authorId}`} className="shrink-0">
        <UserAvatar
          user={
            author ?? {
              full_name: "",
              username: "…",
              avatar_url: "",
            }
          }
          size="xs"
          className="!rounded-full"
        />
      </Link>
      <div className="min-w-0 flex-1">
        <div className="flex min-w-0 items-center gap-2">
          <Link
            href={`/users/${authorId}`}
            className="truncate text-sm font-semibold text-fg hover:underline"
          >
            {username}
          </Link>
          {author ? (
            <span className="shrink-0 rounded bg-brand/15 px-1.5 py-0.5 text-[10px] font-bold tracking-wide text-brand uppercase">
              {level}
            </span>
          ) : null}
        </div>
        <p className="truncate text-xs text-fg-faint">{relativeTime(createdAt)}</p>
      </div>
      {!isSelf && viewerId ? (
        <button
          type="button"
          disabled={busy || !author}
          onClick={(e) => void toggleFollow(e)}
          className={cn(
            "inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors disabled:opacity-50",
            isFollowing
              ? "border border-border-subtle bg-surface text-brand hover:bg-surface-2"
              : "bg-brand text-black hover:bg-brand-dark"
          )}
          aria-label={isFollowing ? "Following" : "Follow"}
          title={isFollowing ? "Following" : "Follow"}
        >
          {isFollowing ? (
            <UserCheck className="h-4 w-4" />
          ) : (
            <Plus className="h-4 w-4" strokeWidth={2.5} />
          )}
        </button>
      ) : null}
    </div>
  );
}

function CaptionAuthorLink({ authorId }: { authorId: string }) {
  const { data: author } = useGetUserProfileQuery(authorId);
  if (!author?.username) return null;
  return (
    <Link href={`/users/${authorId}`} className="mr-1.5 font-semibold hover:underline">
      {author.username}
    </Link>
  );
}

function PostCaption({
  authorId,
  caption,
  clamp,
}: {
  authorId: string;
  caption: string;
  clamp: boolean;
}) {
  const [expanded, setExpanded] = useState(false);
  const [needsMore, setNeedsMore] = useState(false);
  const textRef = useRef<HTMLParagraphElement>(null);

  useEffect(() => {
    setExpanded(false);
  }, [caption, authorId]);

  useEffect(() => {
    if (!clamp || expanded) {
      setNeedsMore(false);
      return;
    }
    const el = textRef.current;
    if (!el) return;
    const measure = () => {
      setNeedsMore(el.scrollHeight > el.clientHeight + 1);
    };
    measure();
    const ro = new ResizeObserver(measure);
    ro.observe(el);
    return () => ro.disconnect();
  }, [caption, clamp, expanded]);

  return (
    <div>
      <p
        ref={textRef}
        className={cn(
          "text-sm leading-relaxed whitespace-pre-wrap text-fg",
          clamp && !expanded && "line-clamp-2"
        )}
      >
        <CaptionAuthorLink authorId={authorId} />
        {caption}
      </p>
      {clamp && needsMore && !expanded ? (
        <button
          type="button"
          onClick={() => setExpanded(true)}
          className="mt-0.5 text-sm font-semibold text-fg-muted hover:text-fg"
        >
          more
        </button>
      ) : null}
      {clamp && expanded ? (
        <button
          type="button"
          onClick={() => setExpanded(false)}
          className="mt-0.5 text-sm font-semibold text-fg-muted hover:text-fg"
        >
          less
        </button>
      ) : null}
    </div>
  );
}

/**
 * One post in the detail / author-continuation / snap feed.
 * `active` is owned by the feed — only one post plays at a time.
 */
export default function PostDetailCard({
  post,
  active,
  className,
  fromUser,
  onOpenComments,
  onOpenMap,
  layout = "default",
}: {
  post: Post;
  active: boolean;
  className?: string;
  fromUser?: boolean;
  onOpenComments?: () => void;
  onOpenMap?: () => void;
  /** `snap` = full-height slide for vertical social feed. */
  layout?: "default" | "snap";
}) {
  const caption = post.caption.trim();
  const title = placeTitle(post);
  const commentsHref = fromUser
    ? `/posts/${post.id}?from=user&comments=1`
    : `/posts/${post.id}?comments=1`;
  const snap = layout === "snap";

  return (
    <article
      data-post-id={post.id}
      className={cn(
        snap
          ? "flex h-[calc(100dvh-5rem)] min-h-[calc(100dvh-5rem)] flex-col bg-bg"
          : "border-b border-border-subtle",
        className
      )}
    >
      <div className={cn("shrink-0", snap && "border-b border-border-subtle/60")}>
        <AuthorHeader authorId={post.author_id} createdAt={post.created_at} />
      </div>

      <div className={cn(snap ? "min-h-0 flex-1 basis-0" : null)}>
        <PostMediaGallery
          media={post.media}
          mode={post.media_mode}
          tall={!snap}
          fill={snap}
          active={active}
          className={snap ? "h-full min-h-0" : undefined}
        />
      </div>

      <div
        className={cn(
          "shrink-0 space-y-2 px-4 pt-2 sm:px-5",
          snap ? "border-t border-border-subtle/60 pb-4" : "space-y-3 pb-6"
        )}
      >
        <PostActions
          post={post}
          commentsHref={commentsHref}
          onCommentsClick={onOpenComments}
        />

        <p className="text-sm font-semibold text-fg">{title}</p>

        {caption ? (
          <PostCaption
            authorId={post.author_id}
            caption={caption}
            clamp={snap}
          />
        ) : (
          <p className="text-sm text-fg">
            <CaptionAuthorLink authorId={post.author_id} />
          </p>
        )}

        <Link
          href={`/qa/ask?post_id=${encodeURIComponent(post.id)}`}
          className="inline-flex text-xs font-semibold text-brand hover:underline"
        >
          Ask about this place
        </Link>

        {!snap && post.hashtags.length > 0 ? (
          <p className="flex flex-wrap gap-x-2 gap-y-1 text-sm">
            {post.hashtags.map((tag) => (
              <Link
                key={tag}
                href={`/hashtags/${encodeURIComponent(tag)}`}
                className="text-brand hover:underline"
              >
                #{tag}
              </Link>
            ))}
          </p>
        ) : null}

        {onOpenMap ? (
          <button
            type="button"
            onClick={onOpenMap}
            className="inline-flex items-center gap-1.5 text-xs text-fg-faint hover:text-fg-muted"
          >
            <MapPin className="h-3.5 w-3.5 text-brand" />
            View map
          </button>
        ) : (
          <a
            href={`https://www.openstreetmap.org/?mlat=${post.lat}&mlon=${post.lng}#map=17/${post.lat}/${post.lng}`}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-1.5 text-xs text-fg-faint hover:text-fg-muted"
          >
            <MapPin className="h-3.5 w-3.5 text-brand" />
            View map
          </a>
        )}
      </div>
    </article>
  );
}
