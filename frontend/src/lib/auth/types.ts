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

export type UserInfo = {
  user_id: string;
  username: string;
  full_name: string;
  phone: string;
  address: string;
  bio: string;
};

export type ApiErrorBody = {
  error: string;
};

export function toAuthSession(res: AuthResponse) {
  return {
    userId: res.user_id,
    username: res.username,
    accessToken: res.access_token,
    refreshToken: res.refresh_token,
  };
}
