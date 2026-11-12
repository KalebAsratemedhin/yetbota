"use client";

import Link from "next/link";
import { Heart, MessageCircle, Share2 } from "lucide-react";
import Reveal from "@/components/landing/Reveal";
import SectionShell from "@/components/landing/SectionShell";
import { useListTopPostsQuery } from "@/lib/api/contentApi";
import { coverUrl, placeTitle } from "@/lib/content/helpers";
import type { Post } from "@/lib/content/types";
import { useContent } from "@/lib/useContent";
import { cn } from "@/lib/utils";

function postEngagement(post: Post): number {
  return (
    post.like_count +
    post.comment_count +
    post.share_count +
    post.bookmark_count
  );
}

function formatCount(n: number): string {
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, "")}k`;
  return String(n);
}

function SpotCard({ post }: { post: Post }) {
  const thumb = coverUrl(post);
  const title = placeTitle(post);
  const score = postEngagement(post);

  return (
    <Link
      href={`/posts/${post.id}`}
      className="group relative block aspect-[3/4] overflow-hidden rounded-3xl bg-surface-2 outline-none ring-brand/40 transition-transform duration-500 hover:-translate-y-1 focus-visible:ring-2"
    >
      {thumb ? (
        // eslint-disable-next-line @next/next/no-img-element -- Cloudinary URLs
        <img
          src={thumb}
          alt=""
          className="absolute inset-0 h-full w-full object-cover transition-transform duration-700 group-hover:scale-105"
        />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-br from-brand/25 via-surface to-surface-2" />
      )}
      <div className="absolute inset-0 bg-gradient-to-t from-black/85 via-black/20 to-transparent" />
      <div className="absolute top-4 right-4 inline-flex items-center gap-1.5 rounded-full bg-black/55 px-2.5 py-1 text-xs font-semibold text-white backdrop-blur-sm">
        <Heart className="h-3.5 w-3.5 fill-current" aria-hidden />
        {formatCount(score)}
      </div>
      <div className="absolute inset-x-0 bottom-0 p-5">
        <p className="line-clamp-2 text-base font-semibold text-white">{title}</p>
        <p className="mt-2 flex flex-wrap items-center gap-3 text-xs text-white/70">
          <span className="inline-flex items-center gap-1">
            <MessageCircle className="h-3.5 w-3.5" aria-hidden />
            {post.comment_count}
          </span>
          <span className="inline-flex items-center gap-1">
            <Share2 className="h-3.5 w-3.5" aria-hidden />
            {post.share_count}
          </span>
        </p>
      </div>
    </Link>
  );
}

function SkeletonStrip({ className }: { className?: string }) {
  return (
    <div
      className={cn("animate-pulse rounded-3xl bg-overlay", className)}
      aria-hidden
    />
  );
}

export default function DiscoverySection() {
  const t = useContent();
  const postsQ = useListTopPostsQuery({ limit: 5 });
  const posts = postsQ.data?.posts ?? [];

  return (
    <SectionShell className="relative overflow-hidden bg-bg">
      <Reveal className="relative mb-8 max-w-2xl">
        <p className="mb-2 text-[11px] font-bold tracking-[0.18em] text-brand uppercase">
          Top posts
        </p>
        <h2 className="text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
          {t.discovery.title}
        </h2>
        <p className="mt-3 text-sm leading-relaxed text-fg-muted sm:text-base">
          {t.discovery.subtitle}
        </p>
        <Link
          href="/"
          className="mt-4 inline-flex text-sm font-semibold text-brand hover:underline"
        >
          Open the feed →
        </Link>
      </Reveal>

      {postsQ.isLoading ? (
        <div className="relative flex gap-4 overflow-hidden">
          {Array.from({ length: 3 }, (_, i) => (
            <SkeletonStrip key={i} className="h-72 w-52 shrink-0 sm:h-80 sm:w-56" />
          ))}
        </div>
      ) : postsQ.isError ? (
        <p className="relative max-w-md text-sm text-fg-muted">
          Top posts will show here once the engagement ranking API is available.
        </p>
      ) : posts.length === 0 ? (
        <p className="relative text-sm text-fg-faint">
          No highly engaged posts yet — drop a place to get started.
        </p>
      ) : (
        <div className="yb-scrollbar relative flex gap-4 overflow-x-auto pb-1 sm:gap-5">
          {posts.map((post, i) => (
            <Reveal
              key={post.id}
              delay={i * 70}
              direction="scale"
              className="w-52 shrink-0 sm:w-56"
            >
              <SpotCard post={post} />
            </Reveal>
          ))}
        </div>
      )}
    </SectionShell>
  );
}
