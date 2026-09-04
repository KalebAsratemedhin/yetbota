/** OpenAPI AuthResponse from identity-service */
export type AuthResponse = {
  user_id: string;
  username: string;
  access_token: string;
  refresh_token: string;
};

export type SignUpRequest = {
  username: string;
  password: string;
  full_name: string;
};

export type SignInRequest = {
  username: string;
  password: string;
};

export type RefreshRequest = {
  refresh_token: string;
};

export type LogoutRequest = {
  refresh_token: string;
};

export type Badge = {
  id: string;
  name: string;
  description: string;
  /** Present when this badge has been awarded to the user */
  awarded_at?: string;
};

export type UserInfo = {
  user_id: string;
  username: string;
  full_name: string;
  phone: string;
  address: string;
  bio: string;
  /** Empty when no avatar */
  avatar_url: string;
  reputation_score: number;
  /** Explorer (<100) | Contributor (≥100) | Local Guide (≥400) | Pathfinder (≥1000) */
  reputation_level: string;
  badges: Badge[];
};

export type Users = {
  users: UserInfo[];
};

/** OpenAPI UsersPage — paginated followers / following */
export type UsersPage = {
  users: UserInfo[];
  total: number;
  next_cursor?: string;
};

export type UpdateProfileRequest = {
  full_name?: string;
  phone?: string;
  address?: string;
  bio?: string;
};

export type ChangePasswordRequest = {
  current_password: string;
  new_password: string;
};

export type StatusOK = {
  status: string;
};

/** OpenAPI Error schema — identity-service structured errors */
export type ApiErrorCode =
  | "VALIDATION_FAILED"
  | "USERNAME_TAKEN"
  | "INVALID_CREDENTIALS"
  | "UNAUTHORIZED"
  | "NOT_FOUND"
  | "INTERNAL_ERROR"
  | (string & {});

export type ApiErrorDetail = {
  field: string;
  code: string;
  message: string;
};

export type ApiErrorPayload = {
  code: ApiErrorCode;
  message: string;
  details?: ApiErrorDetail[];
};

export type ApiErrorBody = {
  error: ApiErrorPayload;
};

export function toAuthSession(res: AuthResponse) {
  return {
    userId: res.user_id,
    username: res.username,
    accessToken: res.access_token,
    refreshToken: res.refresh_token,
  };
}
