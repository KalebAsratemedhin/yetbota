"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, FileText, MapPin, X } from "lucide-react";
import { toast } from "sonner";
import RequireAuth from "@/components/auth/RequireAuth";
import LocationPicker from "@/components/compose/LocationPicker";
import { Navbar } from "@/components/landing";
import PostSearchPicker from "@/components/qa/PostSearchPicker";
import {
  useCreateQuestionMutation,
  useGetPostQuery,
  useGetQuestionQuery,
  useUpdateQuestionMutation,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { coverUrl, placeTitle } from "@/lib/content/helpers";
import type { Post, UpdateQuestionRequest } from "@/lib/content/types";
import { useAppSelector } from "@/lib/hooks";
import { cn } from "@/lib/utils";

type AttachMode = "map" | "post";

function AskForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialPostId = searchParams.get("post_id")?.trim() || "";
  const editId = searchParams.get("edit")?.trim() || "";
  const isEdit = Boolean(editId);
  const viewerId = useAppSelector((s) => s.auth.userId);

  const questionQ = useGetQuestionQuery(editId, { skip: !editId });
  const [hydratedEditId, setHydratedEditId] = useState<string | null>(null);

  const [createQuestion, { isLoading: creating }] = useCreateQuestionMutation();
  const [updateQuestion, { isLoading: updating }] = useUpdateQuestionMutation();
  const isLoading = creating || updating;

  const [body, setBody] = useState("");
  const [mode, setMode] = useState<AttachMode>(initialPostId ? "post" : "map");
  const [selectedPostId, setSelectedPostId] = useState<string | null>(
    initialPostId || null
  );
  const [lat, setLat] = useState<number | null>(null);
  const [lng, setLng] = useState<number | null>(null);
  const [placeLabel, setPlaceLabel] = useState<string | null>(null);

  const linkedPostQ = useGetPostQuery(selectedPostId ?? "", {
    skip: !selectedPostId,
  });

  useEffect(() => {
    if (isEdit) return;
    if (!initialPostId) return;
    setMode("post");
    setSelectedPostId(initialPostId);
  }, [initialPostId, isEdit]);

  useEffect(() => {
    if (!isEdit || !questionQ.data || hydratedEditId === questionQ.data.id) return;
    const q = questionQ.data;
    if (viewerId && q.author_id !== viewerId) {
      toast.error("You can only edit your own questions");
      router.replace(`/qa/${q.id}`);
      return;
    }
    if (q.closed_at) {
      toast.error("Closed questions cannot be edited");
      router.replace(`/qa/${q.id}`);
      return;
    }
    setBody(q.body);
    if (q.post_id) {
      setMode("post");
      setSelectedPostId(q.post_id);
    } else {
      setMode("map");
      setSelectedPostId(null);
    }
    setLat(q.lat ?? null);
    setLng(q.lng ?? null);
    setPlaceLabel(null);
    setHydratedEditId(q.id);
  }, [isEdit, questionQ.data, hydratedEditId, viewerId, router]);

  useEffect(() => {
    if (mode !== "post" || !linkedPostQ.data) return;
    setLat(linkedPostQ.data.lat);
    setLng(linkedPostQ.data.lng);
    setPlaceLabel(placeTitle(linkedPostQ.data));
  }, [mode, linkedPostQ.data]);

  function setLocation(nextLat: number, nextLng: number, label?: string) {
    setLat(nextLat);
    setLng(nextLng);
    setPlaceLabel(label?.trim() || null);
  }

  function switchMode(next: AttachMode) {
    setMode(next);
    if (next === "map") {
      setSelectedPostId(null);
    }
  }

  function onSelectPost(post: Post) {
    setSelectedPostId(post.id);
    setLat(post.lat);
    setLng(post.lng);
    setPlaceLabel(placeTitle(post));
    setMode("post");
  }

  function onClearPost() {
    setSelectedPostId(null);
    setLat(null);
    setLng(null);
    setPlaceLabel(null);
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const b = body.trim();
    if (!b) {
      toast.error("Write your question");
      return;
    }
    if (mode === "post" && !selectedPostId) {
      toast.error("Select a place post, or switch to the map");
      return;
    }
    if (
      mode === "map" &&
      (lat == null ||
        lng == null ||
        !Number.isFinite(lat) ||
        !Number.isFinite(lng))
    ) {
      toast.error("Pin a place on the map");
      return;
    }

    try {
      if (isEdit) {
        const patch: UpdateQuestionRequest = {
          body: b.slice(0, 10_000),
        };
        if (mode === "post" && selectedPostId) {
          patch.post_id = selectedPostId;
          if (lat != null && lng != null) {
            patch.lat = lat;
            patch.lng = lng;
          }
        } else {
          patch.post_id = "";
          if (lat != null && lng != null) {
            patch.lat = lat;
            patch.lng = lng;
          }
        }
        const q = await updateQuestion({ id: editId, body: patch }).unwrap();
        toast.success("Question updated");
        router.replace(`/qa/${q.id}`);
        return;
      }

      const q = await createQuestion({
        body: b.slice(0, 10_000),
        ...(mode === "post" && selectedPostId
          ? { post_id: selectedPostId }
          : {}),
        ...(lat != null && lng != null ? { lat, lng } : {}),
      }).unwrap();
      toast.success("Question posted");
      router.replace(`/qa/${q.id}`);
    } catch (err) {
      toast.error(isEdit ? "Could not update question" : "Could not post question", {
        description: getApiErrorMessage(err),
      });
    }
  }

  const linked = linkedPostQ.data;
  const cover = linked ? coverUrl(linked) : null;
  const canPost =
    body.trim().length > 0 &&
    (mode === "post"
      ? Boolean(selectedPostId)
      : lat != null &&
        lng != null &&
        Number.isFinite(lat) &&
        Number.isFinite(lng));

  if (isEdit && questionQ.isLoading) {
    return (
      <p className="py-20 text-center text-sm text-fg-faint">Loading question…</p>
    );
  }

  if (isEdit && (questionQ.isError || !questionQ.data)) {
    return (
      <div className="mx-auto max-w-lg px-4 py-20 text-center">
        <p className="text-sm text-fg-muted">
          {getApiErrorMessage(questionQ.error) || "Question not found."}
        </p>
        <Link
          href="/qa"
          className="mt-3 inline-block text-sm font-semibold text-brand hover:underline"
        >
          Back to Q&A
        </Link>
      </div>
    );
  }

  return (
    <form
      onSubmit={onSubmit}
      className="mx-auto grid w-full max-w-5xl gap-5 px-4 py-6 sm:gap-6 sm:px-6 sm:py-8 lg:grid-cols-[minmax(17rem,22rem)_minmax(0,1fr)] lg:items-start xl:max-w-6xl"
    >
      <section className="relative z-0 isolate h-[30vh] min-h-[14rem] max-h-64 overflow-hidden rounded-2xl border border-border-subtle lg:order-2 lg:sticky lg:top-24 lg:h-[min(32rem,calc(100dvh-8rem))] lg:max-h-none lg:min-h-0">
        {mode === "map" ? (
          <LocationPicker lat={lat} lng={lng} onChange={setLocation} />
        ) : (
          <PostSearchPicker
            selectedPostId={selectedPostId}
            onSelect={onSelectPost}
            onClear={onClearPost}
          />
        )}
      </section>

      <aside className="relative z-10 flex flex-col overflow-hidden rounded-2xl border border-border-subtle bg-surface lg:order-1">
        <div className="border-b border-border-subtle px-5 py-5 sm:px-6">
          <Link
            href={isEdit ? `/qa/${editId}` : "/qa"}
            className="mb-3 inline-flex items-center gap-1.5 text-sm text-fg-muted hover:text-fg"
          >
            <ArrowLeft className="h-4 w-4" />
            {isEdit ? "Back to question" : "Back to Q&A"}
          </Link>
          <h1 className="text-xl font-bold tracking-tight text-fg">
            {isEdit ? "Edit question" : "Ask a question"}
          </h1>
          <p className="mt-1 text-sm text-fg-muted">
            {isEdit
              ? "Update your question text and where it’s attached."
              : "Write your question, then pin a map location or link a place post."}
          </p>
        </div>

        <div className="yb-scrollbar flex-1 space-y-6 overflow-y-auto px-5 py-5 sm:px-6">
          <section>
            <label className="mb-2 block text-sm font-medium text-fg">
              Attach to
            </label>
            <div className="grid grid-cols-2 gap-1.5">
              {(
                [
                  {
                    id: "map" as const,
                    title: "Map",
                    blurb: "Pick on the map",
                    icon: MapPin,
                  },
                  {
                    id: "post" as const,
                    title: "Place post",
                    blurb: "Search YetBota",
                    icon: FileText,
                  },
                ] as const
              ).map((m) => {
                const Icon = m.icon;
                const active = mode === m.id;
                return (
                  <button
                    key={m.id}
                    type="button"
                    onClick={() => switchMode(m.id)}
                    className={cn(
                      "rounded-xl border px-2.5 py-2.5 text-left transition-colors",
                      active
                        ? "border-brand/50 bg-brand/10"
                        : "border-border-subtle bg-surface hover:bg-surface-2"
                    )}
                  >
                    <Icon
                      className={cn(
                        "mb-1.5 h-4 w-4",
                        active ? "text-brand" : "text-fg-faint"
                      )}
                    />
                    <p className="text-[11px] font-semibold text-fg">{m.title}</p>
                    <p className="text-[10px] text-fg-faint">{m.blurb}</p>
                  </button>
                );
              })}
            </div>
          </section>

          {mode === "post" && selectedPostId ? (
            <section className="rounded-xl border border-brand/25 bg-brand/10 p-3">
              <div className="flex items-start gap-3">
                <div className="relative h-14 w-14 shrink-0 overflow-hidden rounded-lg bg-overlay">
                  {cover ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={cover}
                      alt=""
                      className="h-full w-full object-cover"
                    />
                  ) : null}
                </div>
                <div className="min-w-0 flex-1">
                  <p className="text-[10px] font-bold tracking-wide text-brand uppercase">
                    Linked place
                  </p>
                  {linked ? (
                    <Link
                      href={`/posts/${linked.id}`}
                      className="mt-0.5 block truncate text-sm font-semibold text-fg hover:underline"
                    >
                      {placeTitle(linked)}
                    </Link>
                  ) : (
                    <p className="mt-0.5 text-sm text-fg-muted">Loading…</p>
                  )}
                </div>
                <button
                  type="button"
                  onClick={onClearPost}
                  className="rounded-lg p-1.5 text-fg-muted hover:bg-brand/15 hover:text-fg"
                  aria-label="Clear linked post"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>
            </section>
          ) : null}

          <section>
            <label
              htmlFor="qa-body"
              className="mb-1.5 block text-sm font-medium text-fg"
            >
              Question
            </label>
            <textarea
              id="qa-body"
              rows={8}
              value={body}
              onChange={(e) => setBody(e.target.value.slice(0, 10_000))}
              placeholder="What do you want to know?"
              className="yb-scrollbar w-full resize-y rounded-xl border border-border-subtle bg-surface px-3 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
            />
            {mode === "map" ? (
              placeLabel ? (
                <p className="mt-2 text-xs text-fg-muted">
                  <span className="text-fg-faint">Pinned: </span>
                  {placeLabel}
                </p>
              ) : (
                <p className="mt-2 text-xs text-fg-faint">
                  Pin a place on the map to continue.
                </p>
              )
            ) : !selectedPostId ? (
              <p className="mt-2 text-xs text-fg-faint">
                Search and select a place post on the side.
              </p>
            ) : null}
          </section>
        </div>

        <div className="sticky bottom-0 border-t border-border-subtle bg-surface px-5 py-4 sm:px-6">
          <button
            type="submit"
            disabled={isLoading || !canPost}
            className="inline-flex h-11 w-full items-center justify-center rounded-xl bg-brand text-sm font-semibold text-black transition-colors hover:bg-brand-dark disabled:opacity-40"
          >
            {isLoading
              ? isEdit
                ? "Saving…"
                : "Posting…"
              : isEdit
                ? "Save changes"
                : "Post question"}
          </button>
        </div>
      </aside>
    </form>
  );
}

export default function AskQuestionPage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20 pb-8 sm:pb-10">
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
