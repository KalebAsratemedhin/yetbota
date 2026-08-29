import { baseApi } from "@/lib/api/baseApi";
import { clearCredentials, setCredentials } from "@/lib/auth/authSlice";
import {
  toAuthSession,
  type AuthResponse,
  type LogoutRequest,
  type SignInRequest,
  type SignUpRequest,
  type UserInfo,
} from "@/lib/auth/types";

type AuthRoot = {
  auth: {
    refreshToken: string | null;
  };
};

export const authApi = baseApi.injectEndpoints({
  endpoints: (build) => ({
    signUp: build.mutation<AuthResponse, SignUpRequest>({
      query: (body) => ({
        url: "/auth/signup",
        method: "POST",
        body,
      }),
      async onQueryStarted(_arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(setCredentials(toAuthSession(data)));
        } catch {
          // leave session unchanged
        }
      },
    }),
    signIn: build.mutation<AuthResponse, SignInRequest>({
      query: (body) => ({
        url: "/auth/signin",
        method: "POST",
        body,
      }),
      async onQueryStarted(_arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(setCredentials(toAuthSession(data)));
        } catch {
          // leave session unchanged
        }
      },
    }),
    logout: build.mutation<{ status: string }, void>({
      queryFn: async (_arg, api, _extra, baseQuery) => {
        const refreshToken = (api.getState() as AuthRoot).auth.refreshToken;
        if (!refreshToken) {
          api.dispatch(clearCredentials());
          return { data: { status: "ok" } };
        }
        const result = await baseQuery({
          url: "/auth/logout",
          method: "POST",
          body: { refresh_token: refreshToken } satisfies LogoutRequest,
        });
        api.dispatch(clearCredentials());
        api.dispatch(baseApi.util.resetApiState());
        if (result.error) {
          return { data: { status: "ok" } };
        }
        return { data: result.data as { status: string } };
      },
    }),
    me: build.query<UserInfo, void>({
      query: () => "/auth/me",
      providesTags: ["Me"],
    }),
  }),
});

export const { useSignUpMutation, useSignInMutation, useLogoutMutation, useMeQuery } = authApi;
