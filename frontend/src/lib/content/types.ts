/** OpenAPI types for content-service */

export type MediaKind = "image" | "video" | "audio";
export type MediaMode = "images" | "images_audio" | "video";
export type Reaction = "like" | "dislike";

export type SignMediaResponse = {
  cloud_name: string;
  api_key: string;
  timestamp: number;
  signature: string;
  folder: string;
  public_id: string;
  resource_type: "image" | "video";
  /** Echo of server limit; prefer max_upload_bytes for client checks. Not a Cloudinary form field. */
  max_file_size: number;
  /** Client-side size limit before upload (bytes). Not signed / not sent to Cloudinary. */
  max_upload_bytes: number;
};

export type CreateMedia = {
  kind: MediaKind;
  cloudinary_public_id: string;
  url: string;
  sort_order?: number;
  duration_ms?: number;
  width?: number;
  height?: number;
  mime?: string;
  /** Client-reported file size; when set, API validates ≤ max_upload_bytes. */
  bytes?: number;
};

export type CreatePostRequest = {
  title: string;
  caption?: string;
  lat: number;
  lng: number;
  media_mode: MediaMode;
  media: CreateMedia[];
  hashtags?: string[];
};

/** Partial PATCH — omit fields to leave unchanged. Send hashtags: [] to clear. */
export type UpdatePostRequest = {
  title?: string;
  caption?: string;
  lat?: number;
  lng?: number;
  media_mode?: MediaMode;
  media?: CreateMedia[];
  hashtags?: string[];
};

export type Media = {
  id: string;
  kind: string;
  cloudinary_public_id: string;
  url: string;
  sort_order: number;
  duration_ms: number;
  width: number;
  height: number;
  mime: string;
};

export type Post = {
  id: string;
  author_id: string;
  title: string;
  caption: string;
  lat: number;
  lng: number;
  media_mode: string;
  like_count: number;
  dislike_count: number;
  comment_count: number;
  share_count: number;
  bookmark_count: number;
  created_at: string;
  media: Media[];
  hashtags: string[];
};

export type PostsPage = {
  posts: Post[];
  next_cursor?: string;
};

export type RelatedPostsSource = "search" | "similar" | "mixed";

export type RelatedPostsPage = {
  posts: Post[];
  next_cursor?: string;
  source: RelatedPostsSource;
};

export type TrendingCreator = {
  user_id: string;
  score: number;
};

export type TrendingCreatorsPage = {
  creators: TrendingCreator[];
};

export type Comment = {
  id: string;
  post_id: string;
  author_id: string;
  parent_id?: string | null;
  body: string;
  like_count: number;
  dislike_count: number;
  reply_count: number;
  created_at: string;
};

export type CommentsPage = {
  comments: Comment[];
  next_cursor?: string;
};

export type ShareResponse = {
  post: Post;
  share_url: string;
};

export type QaVote = "up" | "down";

export type CreateQuestionRequest = {
  body: string;
  post_id?: string | null;
  /** Place latitude when attaching a location (with lng). */
  lat?: number | null;
  /** Place longitude when attaching a location (with lat). */
  lng?: number | null;
};

/** Partial PATCH — at least one field. `post_id: ""` clears link; `clear_place: true` clears lat/lng. */
export type UpdateQuestionRequest = {
  body?: string;
  post_id?: string | null;
  lat?: number;
  lng?: number;
  clear_place?: boolean;
};

export type UpdateAnswerRequest = {
  body: string;
};

export type VoteRequest = {
  vote?: QaVote | null;
};

export type Question = {
  id: string;
  author_id: string;
  body: string;
  post_id?: string | null;
  lat?: number | null;
  lng?: number | null;
  upvote_count: number;
  downvote_count: number;
  answer_count: number;
  closed_at?: string | null;
  closed_by?: string | null;
  created_at: string;
};

export type QuestionsPage = {
  questions: Question[];
  next_cursor?: string;
};

export type Answer = {
  id: string;
  question_id: string;
  author_id: string;
  body: string;
  upvote_count: number;
  downvote_count: number;
  created_at: string;
};

export type AnswersPage = {
  answers: Answer[];
  next_cursor?: string;
};

export type CloudinaryUploadResult = {
  public_id: string;
  secure_url: string;
  width?: number;
  height?: number;
  duration?: number;
  format?: string;
  resource_type?: string;
};

export type PlaceEntryCategory = "tip" | "change" | "important";
export type PlaceEntrySource = "ai" | "user";

export type Place = {
  id: string;
  name: string;
  lat: number;
  lng: number;
  summary: string;
  summary_model?: string;
  summary_updated_at?: string;
  post_count: number;
  created_at: string;
  updated_at: string;
};

export type PlacesPage = {
  places: Place[];
  next_cursor?: string;
};

export type PlaceEntry = {
  id: string;
  place_id: string;
  category: PlaceEntryCategory;
  body: string;
  source: PlaceEntrySource;
  source_post_id?: string;
  author_id?: string;
  created_at: string;
};

export type PlaceEntriesPage = {
  entries: PlaceEntry[];
  next_cursor?: string;
};
