"use client";

import Link from "next/link";
import { MapPin } from "lucide-react";
import PostActions from "@/components/posts/PostActions";
import PostMediaGallery from "@/components/posts/PostMediaGallery";
import UserAvatar from "@/components/shared/UserAvatar";
import { useGetUserProfileQuery } from "@/lib/api/socialApi";
import { relativeTime } from "@/lib/content/helpers";
import type { Post } from "@/lib/content/types";
import { cn } from "@/lib/utils";

function AuthorHeader({ authorId, createdAt }: { authorId: string; createdAt: string }) {
  const { data: author } = useGetUserProfileQuery(authorId);

  return (
    <div className="flex items-center gap-3 px-4 py-3">
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
        <Link
          href={`/users/${authorId}`}
          className="block truncate text-sm font-semibold text-fg hover:underline"
        >
          {author?.username ?? "Loading…"}
        </Link>
        <p className="truncate text-xs text-fg-faint">{relativeTime(createdAt)}</p>
      </div>
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
  layout = "default",
}: {
  post: Post;
  active: boolean;
  className?: string;
  fromUser?: boolean;
  onOpenComments?: () => void;
  /** `snap` = full-height slide for vertical social feed. */
  layout?: "default" | "snap";
}) {
  const caption = post.caption.trim();
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
          "shrink-0 space-y-2 px-4 pt-2",
          snap ? "border-t border-border-subtle/60 pb-4" : "space-y-3 pb-6"
        )}
      >
        <PostActions
          post={post}
          commentsHref={commentsHref}
          onCommentsClick={onOpenComments}
        />

        {caption ? (
          <p
            className={cn(
              "text-sm leading-relaxed whitespace-pre-wrap text-fg",
              snap && "line-clamp-2"
            )}
          >
            <CaptionAuthorLink authorId={post.author_id} />
            {caption}
          </p>
        ) : null}

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

        <a
          href={`https://www.openstreetmap.org/?mlat=${post.lat}&mlon=${post.lng}#map=17/${post.lat}/${post.lng}`}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1.5 text-xs text-fg-faint hover:text-fg-muted"
        >
          <MapPin className="h-3.5 w-3.5 text-brand" />
          View map
        </a>
      </div>
    </article>
  );
}
