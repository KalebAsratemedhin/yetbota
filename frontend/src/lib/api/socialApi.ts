import { baseApi } from "@/lib/api/baseApi";
import type { StatusOK, UserInfo, Users, UsersPage } from "@/lib/auth/types";

export type UserListArgs = {
  id: string;
  limit?: number;
  cursor?: string;
};

export const socialApi = baseApi.injectEndpoints({
  // HMR / Fast Refresh re-evaluates this module; allow re-inject.
  overrideExisting: true,
  endpoints: (build) => ({
    searchUsers: build.query<Users, string>({
      query: (q) => ({
        url: "/users/search",
        params: { q },
      }),
    }),
    getUserProfile: build.query<UserInfo, string>({
      query: (id) => `/users/${id}`,
      providesTags: (_r, _e, id) => [{ type: "User", id }],
    }),
    /** Top users by denormalized reputation_score. Public. */
    listReputationLeaders: build.query<Users, { limit?: number } | void>({
      query: (args) => ({
        url: "/users/leaderboard",
        params: { limit: args?.limit ?? 12 },
      }),
      providesTags: ["User"],
    }),
    listFollowers: build.query<UsersPage, UserListArgs>({
      query: ({ id, limit = 20, cursor }) => ({
        url: `/users/${encodeURIComponent(id)}/followers`,
        params: {
          limit,
          ...(cursor ? { cursor } : {}),
        },
      }),
      serializeQueryArgs: ({ endpointName, queryArgs }) =>
        `${endpointName}:${queryArgs.id}:${queryArgs.limit ?? 20}:${queryArgs.cursor ?? ""}`,
      providesTags: (_r, _e, { id }) => [{ type: "Followers", id }],
    }),
    listFollowing: build.query<UsersPage, UserListArgs>({
      query: ({ id, limit = 20, cursor }) => ({
        url: `/users/${encodeURIComponent(id)}/following`,
        params: {
          limit,
          ...(cursor ? { cursor } : {}),
        },
      }),
      serializeQueryArgs: ({ endpointName, queryArgs }) =>
        `${endpointName}:${queryArgs.id}:${queryArgs.limit ?? 20}:${queryArgs.cursor ?? ""}`,
      providesTags: (_r, _e, { id }) => [{ type: "Following", id }],
    }),
    followUser: build.mutation<StatusOK, string>({
      query: (id) => ({
        url: `/users/${id}/follow`,
        method: "POST",
      }),
      invalidatesTags: (_r, _e, id) => [
        { type: "Followers", id },
        { type: "Following" },
      ],
    }),
    unfollowUser: build.mutation<StatusOK, string>({
      query: (id) => ({
        url: `/users/${id}/follow`,
        method: "DELETE",
      }),
      invalidatesTags: (_r, _e, id) => [
        { type: "Followers", id },
        { type: "Following" },
      ],
    }),
  }),
});

export const {
  useSearchUsersQuery,
  useGetUserProfileQuery,
  useListReputationLeadersQuery,
  useListFollowersQuery,
  useListFollowingQuery,
  useLazyListFollowersQuery,
  useLazyListFollowingQuery,
  useFollowUserMutation,
  useUnfollowUserMutation,
} = socialApi;
