"use client";

import { Suspense, use, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import AuthorPostFeed from "@/components/posts/AuthorPostFeed";
import RelatedPostFeed from "@/components/posts/RelatedPostFeed";
import { useGetPostQuery } from "@/lib/api/contentApi";
import { prefetchPostMedia } from "@/lib/media/prefetch";

function PostDetail({ postId }: { postId: string }) {
  const searchParams = useSearchParams();
  const fromUser = searchParams.get("from") === "user";
  const commentsOpen = searchParams.get("comments") === "1";
  const searchQuery = searchParams.get("q")?.trim() || undefined;
  const { data: post, isLoading, isError, refetch } = useGetPostQuery(postId);

  useEffect(() => {
    if (post) prefetchPostMedia(post);
  }, [post]);

  if (isLoading) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center text-sm text-fg-faint">
        Loading…
      </div>
    );
  }

  if (isError || !post) {
    return (
      <div className="flex min-h-[40vh] flex-col items-center justify-center gap-3 px-4">
        <p className="text-sm text-fg-muted">Post not found or unavailable.</p>
        <button
          type="button"
          onClick={() => void refetch()}
          className="text-sm font-medium text-brand hover:underline"
        >
          Retry
        </button>
      </div>
    );
  }

  if (fromUser) {
    return (
      <AuthorPostFeed seed={post} initialMobileComments={commentsOpen} />
    );
  }

  return (
    <RelatedPostFeed
      key={`${post.id}:${searchQuery ?? ""}`}
      seed={post}
      query={searchQuery}
      initialCommentsOpen={commentsOpen}
    />
  );
}

export default function PostPage({
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
          <Suspense
            fallback={
              <div className="flex min-h-[40vh] items-center justify-center text-sm text-fg-faint">
                Loading…
              </div>
            }
          >
            <PostDetail postId={id} />
          </Suspense>
        </div>
      </main>
    </RequireAuth>
  );
}
