"use client";

import { Suspense, use, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import RequireAuth from "@/components/auth/RequireAuth";

/** Legacy `/comments` URL → post page with comments panel open. */
function RedirectBody({ postId }: { postId: string }) {
  const router = useRouter();
  const searchParams = useSearchParams();

  useEffect(() => {
    const next = new URLSearchParams(searchParams.toString());
    next.set("comments", "1");
    router.replace(`/posts/${postId}?${next.toString()}`);
  }, [postId, router, searchParams]);

  return (
    <div className="flex min-h-[40vh] items-center justify-center text-sm text-fg-faint">
      Opening comments…
    </div>
  );
}

export default function PostCommentsRedirectPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);

  return (
    <RequireAuth>
      <main className="min-h-screen bg-bg text-fg">
        <Suspense
          fallback={
            <div className="flex min-h-[40vh] items-center justify-center text-sm text-fg-faint">
              Loading…
            </div>
          }
        >
          <RedirectBody postId={id} />
        </Suspense>
      </main>
    </RequireAuth>
  );
}
