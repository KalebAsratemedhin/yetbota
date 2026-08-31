import { createApi } from "@reduxjs/toolkit/query/react";
import { createReauthBaseQuery, CONTENT_API_BASE } from "@/lib/api/createReauthBaseQuery";
import type {
  Comment,
  CommentsPage,
  CreatePostRequest,
  MediaKind,
  Post,
  PostsPage,
  Reaction,
  ShareResponse,
  SignMediaResponse,
} from "@/lib/content/types";

export type PageArgs = {
  limit?: number;
  cursor?: string;
};

export type UserPostsArgs = PageArgs & { id: string };
export type HashtagPostsArgs = PageArgs & { tag: string };
export type SearchPostsArgs = PageArgs & { q: string };
export type CommentsArgs = PageArgs & { postId: string };
export type CommentRepliesArgs = PageArgs & { commentId: string };
export type FeedArgs = PageArgs & { lat: number; lng: number };

export const contentApi = createApi({
  reducerPath: "contentApi",
  baseQuery: createReauthBaseQuery(CONTENT_API_BASE),
  tagTypes: [
    "Post",
    "UserPosts",
    "Bookmarks",
    "Comments",
    "CommentReplies",
    "HashtagPosts",
    "SearchPosts",
    "Feed",
  ],
  endpoints: (build) => ({
    signMedia: build.mutation<SignMediaResponse, { kind: MediaKind }>({
      query: (body) => ({
        url: "/media/sign",
        method: "POST",
        body,
      }),
    }),
    createPost: build.mutation<Post, CreatePostRequest>({
      query: (body) => ({
        url: "/posts",
        method: "POST",
        body,
      }),
      invalidatesTags: (result) =>
        result
          ? [
              { type: "UserPosts", id: result.author_id },
              { type: "SearchPosts" },
              { type: "HashtagPosts" },
              "Feed",
            ]
          : ["Feed"],
    }),
    getPost: build.query<Post, string>({
      query: (id) => `/posts/${id}`,
      providesTags: (_r, _e, id) => [{ type: "Post", id }],
    }),
    getFeed: build.query<PostsPage, FeedArgs>({
      query: ({ lat, lng, limit = 20, cursor }) => ({
        url: "/feed",
        params: { lat, lng, limit, ...(cursor ? { cursor } : {}) },
      }),
      providesTags: ["Feed"],
    }),
    searchPosts: build.query<PostsPage, SearchPostsArgs>({
      query: ({ q, limit = 20, cursor }) => ({
        url: "/posts/search",
        params: { q, limit, ...(cursor ? { cursor } : {}) },
      }),
      providesTags: ["SearchPosts"],
    }),
    listUserPosts: build.query<PostsPage, UserPostsArgs>({
      query: ({ id, limit = 20, cursor }) => ({
        url: `/users/${id}/posts`,
        params: { limit, ...(cursor ? { cursor } : {}) },
      }),
      providesTags: (_r, _e, { id }) => [{ type: "UserPosts", id }],
    }),
    listMyBookmarks: build.query<PostsPage, PageArgs | void>({
      query: (args) => ({
        url: "/me/bookmarks",
        params: {
          limit: args?.limit ?? 20,
          ...(args?.cursor ? { cursor: args.cursor } : {}),
        },
      }),
      providesTags: ["Bookmarks"],
    }),
    listHashtagPosts: build.query<PostsPage, HashtagPostsArgs>({
      query: ({ tag, limit = 20, cursor }) => ({
        url: `/hashtags/${encodeURIComponent(tag)}/posts`,
        params: { limit, ...(cursor ? { cursor } : {}) },
      }),
      providesTags: (_r, _e, { tag }) => [{ type: "HashtagPosts", id: tag }],
    }),
    setReaction: build.mutation<Post, { postId: string; reaction: Reaction | null }>({
      query: ({ postId, reaction }) => ({
        url: `/posts/${postId}/reaction`,
        method: "POST",
        body: { reaction },
      }),
      invalidatesTags: (result) =>
        result
          ? [
              { type: "Post", id: result.id },
              "Bookmarks",
              "UserPosts",
              "SearchPosts",
              "Feed",
            ]
          : ["Feed"],
    }),
    addComment: build.mutation<
      Comment,
      { postId: string; body: string; parentId?: string }
    >({
      query: ({ postId, body, parentId }) => ({
        url: `/posts/${postId}/comments`,
        method: "POST",
        body: parentId ? { body, parent_id: parentId } : { body },
      }),
      invalidatesTags: (result, _e, { postId, parentId }) => [
        { type: "Comments", id: postId },
        { type: "Post", id: postId },
        ...(parentId ? [{ type: "CommentReplies" as const, id: parentId }] : []),
        ...(result?.parent_id
          ? [{ type: "CommentReplies" as const, id: result.parent_id }]
          : []),
      ],
    }),
    listComments: build.query<CommentsPage, CommentsArgs>({
      query: ({ postId, limit = 20, cursor }) => ({
        url: `/posts/${postId}/comments`,
        params: { limit, ...(cursor ? { cursor } : {}) },
      }),
      providesTags: (_r, _e, { postId }) => [{ type: "Comments", id: postId }],
    }),
    listCommentReplies: build.query<CommentsPage, CommentRepliesArgs>({
      query: ({ commentId, limit = 20, cursor }) => ({
        url: `/comments/${commentId}/replies`,
        params: { limit, ...(cursor ? { cursor } : {}) },
      }),
      providesTags: (_r, _e, { commentId }) => [
        { type: "CommentReplies", id: commentId },
      ],
    }),
    setCommentReaction: build.mutation<
      Comment,
      { commentId: string; reaction: Reaction | null; postId: string; parentId?: string | null }
    >({
      query: ({ commentId, reaction }) => ({
        url: `/comments/${commentId}/reaction`,
        method: "POST",
        body: { reaction },
      }),
      invalidatesTags: (_r, _e, { postId, parentId }) => [
        { type: "Comments", id: postId },
        ...(parentId ? [{ type: "CommentReplies" as const, id: parentId }] : []),
      ],
    }),
    addBookmark: build.mutation<Post, string>({
      query: (postId) => ({
        url: `/posts/${postId}/bookmark`,
        method: "POST",
      }),
      invalidatesTags: (result) =>
        result ? [{ type: "Post", id: result.id }, "Bookmarks", "Feed"] : ["Bookmarks", "Feed"],
    }),
    removeBookmark: build.mutation<Post, string>({
      query: (postId) => ({
        url: `/posts/${postId}/bookmark`,
        method: "DELETE",
      }),
      invalidatesTags: (result) =>
        result ? [{ type: "Post", id: result.id }, "Bookmarks", "Feed"] : ["Bookmarks", "Feed"],
    }),
    sharePost: build.mutation<ShareResponse, string>({
      query: (postId) => ({
        url: `/posts/${postId}/share`,
        method: "POST",
      }),
      invalidatesTags: (result) =>
        result ? [{ type: "Post", id: result.post.id }, "Feed"] : ["Feed"],
    }),
  }),
});

export const {
  useSignMediaMutation,
  useCreatePostMutation,
  useGetPostQuery,
  useGetFeedQuery,
  useLazyGetFeedQuery,
  useSearchPostsQuery,
  useLazySearchPostsQuery,
  useListUserPostsQuery,
  useLazyListUserPostsQuery,
  useListMyBookmarksQuery,
  useLazyListMyBookmarksQuery,
  useListHashtagPostsQuery,
  useLazyListHashtagPostsQuery,
  useSetReactionMutation,
  useAddCommentMutation,
  useListCommentsQuery,
  useLazyListCommentsQuery,
  useListCommentRepliesQuery,
  useLazyListCommentRepliesQuery,
  useSetCommentReactionMutation,
  useAddBookmarkMutation,
  useRemoveBookmarkMutation,
  useSharePostMutation,
} = contentApi;
