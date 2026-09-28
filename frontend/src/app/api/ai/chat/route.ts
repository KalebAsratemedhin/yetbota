import { NextRequest, NextResponse } from "next/server";

const AI_SERVICE_BASE =
  process.env.AI_API_BASE_URL?.trim() ||
  process.env.NEXT_PUBLIC_AI_API_BASE_URL?.trim() ||
  "http://localhost:8083";

/**
 * Same-origin proxy for RAG chat. AI service has no CORS middleware yet;
 * the browser talks to Next, Next forwards to :8083.
 */
export async function POST(req: NextRequest) {
  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return NextResponse.json({ detail: "Invalid JSON body" }, { status: 400 });
  }

  const upstream = `${AI_SERVICE_BASE.replace(/\/$/, "")}/chat`;
  const headers: HeadersInit = {
    Accept: "application/json",
    "Content-Type": "application/json",
  };
  const auth = req.headers.get("authorization");
  if (auth) headers.Authorization = auth;

  try {
    const res = await fetch(upstream, {
      method: "POST",
      headers,
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const text = await res.text();
    const contentType = res.headers.get("content-type") ?? "application/json";
    return new NextResponse(text, {
      status: res.status,
      headers: { "Content-Type": contentType },
    });
  } catch {
    return NextResponse.json(
      { detail: "Cannot reach AI service. Is it running on :8083?" },
      { status: 502 }
    );
  }
}
