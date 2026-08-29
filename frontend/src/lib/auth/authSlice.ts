import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

const AUTH_STORAGE_KEY = "yetbota.auth";

export type AuthSession = {
  userId: string;
  username: string;
  accessToken: string;
  refreshToken: string;
};

export type AuthState = {
  userId: string | null;
  username: string | null;
  accessToken: string | null;
  refreshToken: string | null;
};

const emptyAuth: AuthState = {
  userId: null,
  username: null,
  accessToken: null,
  refreshToken: null,
};

export function loadAuthFromStorage(): AuthState {
  if (typeof window === "undefined") return emptyAuth;
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY);
    if (!raw) return emptyAuth;
    const parsed = JSON.parse(raw) as Partial<AuthSession>;
    if (!parsed.accessToken || !parsed.refreshToken || !parsed.userId || !parsed.username) {
      return emptyAuth;
    }
    return {
      userId: parsed.userId,
      username: parsed.username,
      accessToken: parsed.accessToken,
      refreshToken: parsed.refreshToken,
    };
  } catch {
    return emptyAuth;
  }
}

function persistAuth(state: AuthState) {
  if (typeof window === "undefined") return;
  if (!state.accessToken || !state.refreshToken || !state.userId || !state.username) {
    localStorage.removeItem(AUTH_STORAGE_KEY);
    return;
  }
  const session: AuthSession = {
    userId: state.userId,
    username: state.username,
    accessToken: state.accessToken,
    refreshToken: state.refreshToken,
  };
  localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(session));
}

const authSlice = createSlice({
  name: "auth",
  initialState: emptyAuth,
  reducers: {
    setCredentials(state, action: PayloadAction<AuthSession>) {
      state.userId = action.payload.userId;
      state.username = action.payload.username;
      state.accessToken = action.payload.accessToken;
      state.refreshToken = action.payload.refreshToken;
      persistAuth(state);
    },
    clearCredentials(state) {
      state.userId = null;
      state.username = null;
      state.accessToken = null;
      state.refreshToken = null;
      persistAuth(state);
    },
  },
});

export const { setCredentials, clearCredentials } = authSlice.actions;
export const authReducer = authSlice.reducer;
