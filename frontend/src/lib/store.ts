import { configureStore } from "@reduxjs/toolkit";
import { baseApi } from "@/lib/api/baseApi";
import { contentApi } from "@/lib/api/contentApi";
import { aiApi } from "@/lib/api/aiApi";
import { authReducer, loadAuthFromStorage } from "@/lib/auth/authSlice";
import "@/lib/api/authApi";
import "@/lib/api/socialApi";

export const makeStore = () =>
  configureStore({
    reducer: {
      auth: authReducer,
      [baseApi.reducerPath]: baseApi.reducer,
      [contentApi.reducerPath]: contentApi.reducer,
      [aiApi.reducerPath]: aiApi.reducer,
    },
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware().concat(
        baseApi.middleware,
        contentApi.middleware,
        aiApi.middleware
      ),
    preloadedState: {
      auth: loadAuthFromStorage(),
    },
  });

export type AppStore = ReturnType<typeof makeStore>;
export type RootState = ReturnType<AppStore["getState"]>;
export type AppDispatch = AppStore["dispatch"];
