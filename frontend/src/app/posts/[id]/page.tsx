"use client";

import { Suspense, use, useCallback, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import AuthorPostFeed from "@/components/posts/AuthorPostFeed";
import CommentsPanel from "@/components/posts/CommentsPanel";
import PostDetailCard from "@/components/posts/PostDetailCard";
import { useGetPostQuery } from "@/lib/api/contentApi";
import type { Post } from "@/lib/content/types";
import { cn } from "@/lib/utils";

function SinglePostWithComments({
  post,
  initialMobileOpen,
}: {
  post: Post;
  initialMobileOpen: boolean;
}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [mobileOpen, setMobileOpen] = useState(initialMobileOpen);

  const closeMobile = useCallback(() => {
    setMobileOpen(false);
    const next = new URLSearchParams(searchParams.toString());
    next.delete("comments");
    const qs = next.toString();
    router.replace(qs ? `?${qs}` : "?", { scroll: false });
  }, [router, searchParams]);

  const openMobile = useCallback(() => {
    setMobileOpen(true);
    const next = new URLSearchParams(searchParams.toString());
    next.set("comments", "1");
    router.replace(`?${next.toString()}`, { scroll: false });
  }, [router, searchParams]);

  return (
    <div
      className={cn(
        "lg:grid lg:grid-cols-[minmax(0,36rem)_22rem] lg:items-start lg:gap-6",
        mobileOpen && "pb-[min(52vh,28rem)] lg:pb-0"
      )}
    >
      <div className="mx-auto w-full max-w-xl lg:mx-0">
        <PostDetailCard post={post} active onOpenComments={openMobile} />
      </div>
      <CommentsPanel
        postId={post.id}
        mobileOpen={mobileOpen}
        onMobileClose={closeMobile}
      />
    </div>
  );
}

function PostDetail({ postId }: { postId: string }) {
  const searchParams = useSearchParams();
  const fromUser = searchParams.get("from") === "user";
  const commentsOpen = searchParams.get("comments") === "1";
  const { data: post, isLoading, isError, refetch } = useGetPostQuery(postId);

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
        <SinglePostWithComments post={post} initialMobileOpen={commentsOpen} />
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
