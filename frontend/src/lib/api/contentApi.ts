import { createApi } from "@reduxjs/toolkit/query/react";
import { createReauthBaseQuery, CONTENT_API_BASE } from "@/lib/api/createReauthBaseQuery";
import type {
  Answer,
  AnswersPage,
  Comment,
  CommentsPage,
  CreatePostRequest,
  CreateQuestionRequest,
  MediaKind,
  Post,
  PostsPage,
  QaVote,
  Question,
  QuestionsPage,
  Reaction,
  RelatedPostsPage,
  ShareResponse,
  SignMediaResponse,
  TrendingCreatorsPage,
  VoteRequest,
} from "@/lib/content/types";

export type PageArgs = {
  limit?: number;
  cursor?: string;
};

export type UserPostsArgs = PageArgs & { id: string };
export type HashtagPostsArgs = PageArgs & { tag: string };
export type SearchPostsArgs = PageArgs & { q: string };
export type RelatedPostsArgs = PageArgs & { id: string; q?: string };
export type CommentsArgs = PageArgs & { postId: string };
export type CommentRepliesArgs = PageArgs & { commentId: string };
export type FeedArgs = PageArgs & { lat?: number; lng?: number };
export type ListQuestionsArgs = PageArgs & { post_id?: string };
export type ListAnswersArgs = PageArgs & { questionId: string };

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
    "RelatedPosts",
    "TrendingPosts",
    "TrendingCreators",
    "Questions",
    "Question",
    "Answers",
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
              "RelatedPosts",
              "TrendingPosts",
              "TrendingCreators",
            ]
          : ["Feed", "RelatedPosts", "TrendingPosts", "TrendingCreators"],
    }),
    getPost: build.query<Post, string>({
      query: (id) => `/posts/${id}`,
      providesTags: (_r, _e, id) => [{ type: "Post", id }],
    }),
    listRelatedPosts: build.query<RelatedPostsPage, RelatedPostsArgs>({
      query: ({ id, q, limit = 12, cursor }) => ({
        url: `/posts/${id}/related`,
        params: {
          limit,
          ...(q?.trim() ? { q: q.trim() } : {}),
          ...(cursor ? { cursor } : {}),
        },
      }),
      providesTags: (_r, _e, { id }) => [{ type: "RelatedPosts", id }],
    }),
    getFeed: build.query<PostsPage, FeedArgs | void>({
      query: (args) => {
        const lat = args?.lat;
        const lng = args?.lng;
        const hasLoc =
          lat != null &&
          lng != null &&
          Number.isFinite(lat) &&
          Number.isFinite(lng);
        return {
          url: "/feed",
          params: {
            limit: args?.limit ?? 20,
            ...(args?.cursor ? { cursor: args.cursor } : {}),
            ...(hasLoc ? { lat, lng } : {}),
          },
        };
      },
      providesTags: ["Feed"],
    }),
    listTrendingPosts: build.query<PostsPage, PageArgs | void>({
      query: (args) => ({
        url: "/posts/trending",
        params: {
          limit: args?.limit ?? 8,
          ...(args?.cursor ? { cursor: args.cursor } : {}),
        },
      }),
      providesTags: ["TrendingPosts"],
    }),
    listTrendingCreators: build.query<TrendingCreatorsPage, { limit?: number } | void>({
      query: (args) => ({
        url: "/creators/trending",
        params: {
          limit: args?.limit ?? 8,
        },
      }),
      providesTags: ["TrendingCreators"],
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
    listQuestions: build.query<QuestionsPage, ListQuestionsArgs | void>({
      query: (args) => ({
        url: "/questions",
        params: {
          limit: args?.limit ?? 20,
          ...(args?.cursor ? { cursor: args.cursor } : {}),
          ...(args?.post_id ? { post_id: args.post_id } : {}),
        },
      }),
      providesTags: (result, _e, args) => [
        "Questions",
        ...(args && "post_id" in args && args.post_id
          ? [{ type: "Questions" as const, id: args.post_id }]
          : []),
        ...(result?.questions.map((q) => ({ type: "Question" as const, id: q.id })) ??
          []),
      ],
    }),
    getQuestion: build.query<Question, string>({
      query: (id) => `/questions/${id}`,
      providesTags: (_r, _e, id) => [{ type: "Question", id }],
    }),
    createQuestion: build.mutation<Question, CreateQuestionRequest>({
      query: (body) => ({
        url: "/questions",
        method: "POST",
        body,
      }),
      invalidatesTags: (result) =>
        result
          ? [
              "Questions",
              { type: "Question", id: result.id },
              ...(result.post_id
                ? [{ type: "Questions" as const, id: result.post_id }]
                : []),
            ]
          : ["Questions"],
    }),
    closeQuestion: build.mutation<Question, string>({
      query: (id) => ({
        url: `/questions/${id}/close`,
        method: "POST",
      }),
      invalidatesTags: (result) =>
        result ? [{ type: "Question", id: result.id }, "Questions"] : ["Questions"],
    }),
    voteQuestion: build.mutation<Question, { id: string; vote: QaVote | null }>({
      query: ({ id, vote }) => ({
        url: `/questions/${id}/vote`,
        method: "POST",
        body: { vote } satisfies VoteRequest,
      }),
      invalidatesTags: (result) =>
        result ? [{ type: "Question", id: result.id }, "Questions"] : ["Questions"],
    }),
    listAnswers: build.query<AnswersPage, ListAnswersArgs>({
      query: ({ questionId, limit = 20, cursor }) => ({
        url: `/questions/${questionId}/answers`,
        params: { limit, ...(cursor ? { cursor } : {}) },
      }),
      providesTags: (_r, _e, { questionId }) => [{ type: "Answers", id: questionId }],
    }),
    createAnswer: build.mutation<Answer, { questionId: string; body: string }>({
      query: ({ questionId, body }) => ({
        url: `/questions/${questionId}/answers`,
        method: "POST",
        body: { body },
      }),
      invalidatesTags: (result, _e, { questionId }) => [
        { type: "Answers", id: questionId },
        { type: "Question", id: questionId },
        "Questions",
        ...(result ? [] : []),
      ],
    }),
    voteAnswer: build.mutation<
      Answer,
      { id: string; questionId: string; vote: QaVote | null }
    >({
      query: ({ id, vote }) => ({
        url: `/answers/${id}/vote`,
        method: "POST",
        body: { vote } satisfies VoteRequest,
      }),
      invalidatesTags: (_r, _e, { questionId }) => [
        { type: "Answers", id: questionId },
        { type: "Question", id: questionId },
      ],
    }),
  }),
});

export const {
  useSignMediaMutation,
  useCreatePostMutation,
  useGetPostQuery,
  useListRelatedPostsQuery,
  useLazyListRelatedPostsQuery,
  useGetFeedQuery,
  useLazyGetFeedQuery,
  useListTrendingPostsQuery,
  useLazyListTrendingPostsQuery,
  useListTrendingCreatorsQuery,
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
  useListQuestionsQuery,
  useLazyListQuestionsQuery,
  useGetQuestionQuery,
  useCreateQuestionMutation,
  useCloseQuestionMutation,
  useVoteQuestionMutation,
  useListAnswersQuery,
  useLazyListAnswersQuery,
  useCreateAnswerMutation,
  useVoteAnswerMutation,
} = contentApi;
