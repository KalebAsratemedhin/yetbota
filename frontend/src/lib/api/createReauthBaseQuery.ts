import {
  fetchBaseQuery,
  type BaseQueryFn,
  type FetchArgs,
  type FetchBaseQueryError,
} from "@reduxjs/toolkit/query";
import { clearCredentials, setCredentials, type AuthState } from "@/lib/auth/authSlice";
import { toAuthSession, type AuthResponse } from "@/lib/auth/types";

export const IDENTITY_API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8081";

export const CONTENT_API_BASE =
  process.env.NEXT_PUBLIC_CONTENT_API_BASE_URL ?? "http://localhost:8082";

type AuthRoot = { auth: AuthState };

/** Refresh always hits identity-service; requests use `serviceBase`. */
export function createReauthBaseQuery(
  serviceBase: string
): BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> {
  const serviceQuery = fetchBaseQuery({
    baseUrl: serviceBase,
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as AuthRoot).auth.accessToken;
      if (token) headers.set("Authorization", `Bearer ${token}`);
      return headers;
    },
  });

  const identityQuery = fetchBaseQuery({
    baseUrl: IDENTITY_API_BASE,
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as AuthRoot).auth.accessToken;
      if (token) headers.set("Authorization", `Bearer ${token}`);
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

      const result = await identityQuery(
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

  return async (args, api, extraOptions) => {
    let result = await serviceQuery(args, api, extraOptions);
    const url = requestUrl(args);

    if (result.error?.status === 401 && !shouldSkipReauth(url)) {
      const refreshed = await tryRefresh(api, extraOptions);
      if (refreshed) {
        result = await serviceQuery(args, api, extraOptions);
      }
    }

    return result;
  };
}
