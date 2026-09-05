"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, MapPin } from "lucide-react";
import { toast } from "sonner";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import {
  useCreateQuestionMutation,
  useGetPostQuery,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { placeTitle } from "@/lib/content/helpers";
import { useLocationPreference } from "@/hooks/use-location-preference";

function AskForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const postId = searchParams.get("post_id")?.trim() || undefined;
  const { enabled: locationEnabled } = useLocationPreference();
  const postQ = useGetPostQuery(postId ?? "", { skip: !postId });
  const [createQuestion, { isLoading }] = useCreateQuestionMutation();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [attachLoc, setAttachLoc] = useState(false);
  const [coords, setCoords] = useState<{ lat: number; lng: number } | null>(null);

  useEffect(() => {
    if (!attachLoc || !locationEnabled) {
      setCoords(null);
      return;
    }
    if (!navigator.geolocation) return;
    navigator.geolocation.getCurrentPosition(
      (pos) =>
        setCoords({
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
        }),
      () => {
        toast.error("Could not read location");
        setAttachLoc(false);
      },
      { enableHighAccuracy: true, timeout: 12_000, maximumAge: 60_000 }
    );
  }, [attachLoc, locationEnabled]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const t = title.trim();
    const b = body.trim();
    if (!t || !b) {
      toast.error("Title and body are required");
      return;
    }
    try {
      const q = await createQuestion({
        title: t.slice(0, 300),
        body: b.slice(0, 10_000),
        ...(postId ? { post_id: postId } : {}),
        ...(coords
          ? { lat: coords.lat, lng: coords.lng }
          : {}),
      }).unwrap();
      toast.success("Question posted");
      router.replace(`/qa/${q.id}`);
    } catch (err) {
      toast.error("Could not post question", {
        description: getApiErrorMessage(err),
      });
    }
  }

  return (
    <div className="mx-auto w-full max-w-2xl px-4 pt-8 pb-16 sm:px-6 sm:pt-10">
      <Link
        href="/qa"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-fg-muted hover:text-fg"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Q&A
      </Link>

      <h1 className="text-2xl font-extrabold tracking-tight text-fg">
        Ask a question
      </h1>
      <p className="mt-2 text-sm text-fg-muted">
        Clear titles get better answers. Link a place post when it helps.
      </p>

      {postId ? (
        <div className="mt-6 rounded-xl border border-brand/25 bg-brand/10 px-4 py-3 text-sm">
          {postQ.data ? (
            <p className="text-fg">
              Linked to{" "}
              <Link
                href={`/posts/${postId}`}
                className="font-semibold text-brand hover:underline"
              >
                {placeTitle(postQ.data)}
              </Link>
            </p>
          ) : (
            <p className="text-fg-muted">Linked to a place post…</p>
          )}
        </div>
      ) : null}

      <form onSubmit={onSubmit} className="mt-8 space-y-5">
        <div>
          <label htmlFor="qa-title" className="mb-1.5 block text-sm font-medium text-fg">
            Title
          </label>
          <input
            id="qa-title"
            value={title}
            onChange={(e) => setTitle(e.target.value.slice(0, 300))}
            maxLength={300}
            placeholder="e.g. Best coffee near Piazza on a Sunday?"
            className="w-full rounded-xl border border-border-subtle bg-surface px-3 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
          />
        </div>
        <div>
          <label htmlFor="qa-body" className="mb-1.5 block text-sm font-medium text-fg">
            Body
          </label>
          <textarea
            id="qa-body"
            rows={8}
            value={body}
            onChange={(e) => setBody(e.target.value.slice(0, 10_000))}
            placeholder="Add details, context, and what you’ve already tried…"
            className="yb-scrollbar w-full resize-y rounded-xl border border-border-subtle bg-surface px-3 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
          />
        </div>

        {locationEnabled ? (
          <label className="flex cursor-pointer items-center gap-2.5 text-sm text-fg-muted">
            <input
              type="checkbox"
              checked={attachLoc}
              onChange={(e) => setAttachLoc(e.target.checked)}
              className="rounded border-border-subtle"
            />
            <MapPin className="h-3.5 w-3.5 text-brand" />
            Attach my current location
            {coords ? (
              <span className="text-xs text-fg-faint">
                ({coords.lat.toFixed(4)}, {coords.lng.toFixed(4)})
              </span>
            ) : null}
          </label>
        ) : null}

        <button
          type="submit"
          disabled={isLoading || !title.trim() || !body.trim()}
          className="inline-flex h-11 w-full items-center justify-center rounded-xl bg-brand text-sm font-semibold text-black hover:bg-brand-dark disabled:opacity-40 sm:w-auto sm:px-8"
        >
          {isLoading ? "Posting…" : "Post question"}
        </button>
      </form>
    </div>
  );
}

export default function AskQuestionPage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <Suspense
            fallback={
              <p className="py-20 text-center text-sm text-fg-faint">Loading…</p>
            }
          >
            <AskForm />
          </Suspense>
        </div>
      </main>
    </RequireAuth>
  );
}
