"use client";

import Link from "next/link";
import { MapPin } from "lucide-react";
import PostActions from "@/components/posts/PostActions";
import PostMediaGallery from "@/components/posts/PostMediaGallery";
import { placeTitle, relativeTime } from "@/lib/content/helpers";
import type { Post } from "@/lib/content/types";

export default function PostCard({
  post,
  bookmarked,
  fromUser,
  searchQuery,
}: {
  post: Post;
  bookmarked?: boolean;
  /** Opens post in author-continuation mode (profile Locations). */
  fromUser?: boolean;
  /** Opens post with related continuation using this search query. */
  searchQuery?: string;
}) {
  const qs = new URLSearchParams();
  if (fromUser) qs.set("from", "user");
  if (searchQuery?.trim()) qs.set("q", searchQuery.trim());
  const query = qs.toString();
  const href = query ? `/posts/${post.id}?${query}` : `/posts/${post.id}`;
  const commentsQs = new URLSearchParams(qs);
  commentsQs.set("comments", "1");
  const commentsHref = `/posts/${post.id}?${commentsQs.toString()}`;
  return (
    <article className="group overflow-hidden rounded-2xl border border-border-subtle bg-surface shadow-[0_12px_40px_-28px_rgba(0,0,0,0.55)] transition-transform hover:-translate-y-0.5">
      <Link href={href} className="block">
        <PostMediaGallery media={post.media} mode={post.media_mode} />
      </Link>
      <div className="space-y-3 p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <Link
              href={href}
              className="line-clamp-2 text-sm font-semibold text-fg transition-colors hover:text-brand"
            >
              {placeTitle(post)}
            </Link>
            <div className="mt-1.5 flex items-center gap-1.5 text-[11px] text-fg-faint">
              <a
                href={`https://www.openstreetmap.org/?mlat=${post.lat}&mlon=${post.lng}#map=17/${post.lat}/${post.lng}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1 hover:text-fg-muted"
                onClick={(e) => e.stopPropagation()}
              >
                <MapPin className="h-3 w-3 shrink-0 text-brand" />
                Map
              </a>
              <span>·</span>
              <span>{relativeTime(post.created_at)}</span>
            </div>
          </div>
          <Link
            href={`/users/${post.author_id}`}
            className="shrink-0 rounded-full border border-border-subtle bg-overlay px-2 py-0.5 text-[10px] font-semibold text-fg-muted hover:text-fg"
          >
            Author
          </Link>
        </div>

        {post.hashtags.length > 0 ? (
          <div className="flex flex-wrap gap-1.5">
            {post.hashtags.slice(0, 4).map((tag) => (
              <Link
                key={tag}
                href={`/hashtags/${encodeURIComponent(tag)}`}
                className="rounded-full bg-brand/10 px-2 py-0.5 text-[10px] font-bold tracking-wide text-brand uppercase hover:bg-brand/20"
              >
                #{tag}
              </Link>
            ))}
          </div>
        ) : null}

        <PostActions
          post={post}
          initialBookmarked={bookmarked}
          compact
          commentsHref={commentsHref}
        />
      </div>
    </article>
  );
}
