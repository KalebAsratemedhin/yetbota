import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import type { AuthState } from "@/lib/auth/authSlice";
import type { ChatRequest, ChatResponse } from "@/lib/ai/types";

/**
 * Browser calls same-origin Next proxy (`/api/ai`) to avoid AI-service CORS.
 * The route forwards to `NEXT_PUBLIC_AI_API_BASE_URL` / `AI_API_BASE_URL`.
 */
export const AI_API_BASE = "/api/ai";

type AuthRoot = { auth: AuthState };

export const aiApi = createApi({
  reducerPath: "aiApi",
  baseQuery: fetchBaseQuery({
    baseUrl: AI_API_BASE,
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as AuthRoot).auth.accessToken;
      if (token) headers.set("Authorization", `Bearer ${token}`);
      return headers;
    },
  }),
  endpoints: (build) => ({
    chat: build.mutation<ChatResponse, ChatRequest>({
      query: (body) => ({
        url: "/chat",
        method: "POST",
        body,
      }),
    }),
  }),
});

export const { useChatMutation } = aiApi;
