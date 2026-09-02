"use client";

import { Suspense, use, useCallback, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import AuthorPostFeed from "@/components/posts/AuthorPostFeed";
import CommentsPanel from "@/components/posts/CommentsPanel";
import { sidePanelGridClass } from "@/components/posts/commentsLayout";
import PostDetailCard from "@/components/posts/PostDetailCard";
import PostMapPanel from "@/components/posts/PostMapPanel";
import { usePostSidePanel } from "@/hooks/use-post-side-panel";
import { useGetPostQuery } from "@/lib/api/contentApi";
import { prefetchPostMedia } from "@/lib/media/prefetch";
import type { Post } from "@/lib/content/types";
import { cn } from "@/lib/utils";

function SinglePostWithComments({
  post,
  initialCommentsOpen,
}: {
  post: Post;
  initialCommentsOpen: boolean;
}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const {
    commentsOpen,
    mapOpen,
    sideOpen,
    onCommentsClick,
    onMapClick,
    closeMobile,
  } = usePostSidePanel({
    initialMode: initialCommentsOpen ? "comments" : "none",
    initialPostId: post.id,
  });

  const syncCommentsUrl = useCallback(
    (open: boolean) => {
      const next = new URLSearchParams(searchParams.toString());
      if (open) next.set("comments", "1");
      else next.delete("comments");
      const qs = next.toString();
      router.replace(qs ? `?${qs}` : "?", { scroll: false });
    },
    [router, searchParams]
  );

  useEffect(() => {
    const has = searchParams.get("comments") === "1";
    if (has === commentsOpen) return;
    syncCommentsUrl(commentsOpen);
    // Only react to commentsOpen; searchParams read for comparison.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [commentsOpen]);

  const handleComments = useCallback(() => {
    onCommentsClick(post.id);
  }, [onCommentsClick, post.id]);

  const handleMap = useCallback(() => {
    onMapClick(post.id);
  }, [onMapClick, post.id]);

  return (
    <div
      className={cn(
        sidePanelGridClass(sideOpen),
        sideOpen && "pb-[min(52vh,28rem)] lg:pb-0"
      )}
    >
      <div className="mx-auto w-full max-w-xl lg:mx-0">
        <PostDetailCard
          post={post}
          active
          onOpenComments={handleComments}
          onOpenMap={handleMap}
        />
      </div>
      {commentsOpen ? (
        <CommentsPanel
          postId={post.id}
          mobileOpen={commentsOpen}
          onMobileClose={closeMobile}
          desktopOpen={commentsOpen}
        />
      ) : null}
      {mapOpen ? (
        <PostMapPanel
          lat={post.lat}
          lng={post.lng}
          mobileOpen={mapOpen}
          onMobileClose={closeMobile}
          desktopOpen={mapOpen}
        />
      ) : null}
    </div>
  );
}

function PostDetail({ postId }: { postId: string }) {
  const searchParams = useSearchParams();
  const fromUser = searchParams.get("from") === "user";
  const commentsOpen = searchParams.get("comments") === "1";
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

  return (
    <div className="mx-auto w-full max-w-6xl pb-16">
      {fromUser ? (
        <AuthorPostFeed seed={post} initialMobileComments={commentsOpen} />
      ) : (
        <SinglePostWithComments post={post} initialCommentsOpen={commentsOpen} />
      )}
    </div>
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
