import type { ChatSource } from "@/lib/ai/types";

/** Deep-link a RAG source into the app when we have enough IDs. */
export function chatSourceHref(source: ChatSource): string | null {
  const kind = source.kind.trim().toLowerCase();
  switch (kind) {
    case "post":
      return `/posts/${source.entity_id}`;
    case "comment":
      if (source.post_id) return `/posts/${source.post_id}`;
      if (source.parent_id) return `/posts/${source.parent_id}`;
      return null;
    case "question":
      return `/qa/${source.entity_id}`;
    case "answer":
      return source.parent_id ? `/qa/${source.parent_id}` : null;
    default:
      if (source.post_id) return `/posts/${source.post_id}`;
      return null;
  }
}

export function chatSourceLabel(source: ChatSource): string {
  const title = source.title?.trim();
  if (title) return title;
  const kind = source.kind.trim() || "source";
  return kind.charAt(0).toUpperCase() + kind.slice(1);
}
