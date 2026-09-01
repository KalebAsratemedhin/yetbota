import type { Post } from "@/lib/content/types";

const prefetched = new Set<string>();

/** Warm the browser HTTP cache for a media URL (deduped). */
export function prefetchMediaUrl(url: string) {
  if (!url || prefetched.has(url)) return;
  prefetched.add(url);

  const lower = url.toLowerCase();
  const isVideo =
    /\.(mp4|webm|mov|m4v)(\?|$)/.test(lower) ||
    lower.includes("/video/upload/");

  if (isVideo) {
    const link = document.createElement("link");
    link.rel = "prefetch";
    link.as = "video";
    link.href = url;
    document.head.appendChild(link);
    return;
  }

  const img = new Image();
  img.decoding = "async";
  img.src = url;
}

export function prefetchPostMedia(post: Post) {
  for (const item of post.media) {
    prefetchMediaUrl(item.url);
  }
}
