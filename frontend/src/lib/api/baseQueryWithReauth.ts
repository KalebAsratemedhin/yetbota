import {
  fetchBaseQuery,
  type BaseQueryFn,
  type FetchArgs,
  type FetchBaseQueryError,
} from "@reduxjs/toolkit/query";
import { clearCredentials, setCredentials, type AuthState } from "@/lib/auth/authSlice";
import { toAuthSession, type AuthResponse } from "@/lib/auth/types";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8081";

type AuthRoot = { auth: AuthState };

const rawBaseQuery = fetchBaseQuery({
  baseUrl: API_BASE,
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as AuthRoot).auth.accessToken;
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    return headers;
  },
});

function requestUrl(args: string | FetchArgs): string {
  return typeof args === "string" ? args : args.url;
}

function shouldSkipReauth(url: string): boolean {
  return (
    url.includes("/auth/signin") ||
    url.includes("/auth/signup") ||
    url.includes("/auth/refresh")
  );
}

/** Serialize concurrent 401s so only one refresh runs. */
let refreshInFlight: Promise<boolean> | null = null;

async function tryRefresh(
  api: Parameters<BaseQueryFn>[1],
  extraOptions: Parameters<BaseQueryFn>[2]
): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;

  refreshInFlight = (async () => {
    const refreshToken = (api.getState() as AuthRoot).auth.refreshToken;
    if (!refreshToken) {
      api.dispatch(clearCredentials());
      return false;
    }

    const result = await rawBaseQuery(
      {
        url: "/auth/refresh",
        method: "POST",
        body: { refresh_token: refreshToken },
      },
      api,
      extraOptions
    );

    if (result.data) {
      api.dispatch(setCredentials(toAuthSession(result.data as AuthResponse)));
      return true;
    }

    api.dispatch(clearCredentials());
    return false;
  })().finally(() => {
    refreshInFlight = null;
  });

  return refreshInFlight;
}

export const baseQueryWithReauth: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  let result = await rawBaseQuery(args, api, extraOptions);
  const url = requestUrl(args);

  if (result.error?.status === 401 && !shouldSkipReauth(url)) {
    const refreshed = await tryRefresh(api, extraOptions);
    if (refreshed) {
      result = await rawBaseQuery(args, api, extraOptions);
    }
  }

  return result;
};
