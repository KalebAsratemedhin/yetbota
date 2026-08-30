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
};

export type CreatePostRequest = {
  caption?: string;
  lat: number;
  lng: number;
  media_mode: MediaMode;
  media: CreateMedia[];
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

export type CloudinaryUploadResult = {
  public_id: string;
  secure_url: string;
  width?: number;
  height?: number;
  duration?: number;
  format?: string;
  resource_type?: string;
};
