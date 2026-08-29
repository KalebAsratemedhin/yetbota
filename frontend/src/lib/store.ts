import { configureStore } from "@reduxjs/toolkit";
import { baseApi } from "@/lib/api/baseApi";
import { authReducer, loadAuthFromStorage } from "@/lib/auth/authSlice";
import "@/lib/api/authApi";

export const makeStore = () =>
  configureStore({
    reducer: {
      auth: authReducer,
      [baseApi.reducerPath]: baseApi.reducer,
    },
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware().concat(baseApi.middleware),
    preloadedState: {
      auth: loadAuthFromStorage(),
    },
  });

export type AppStore = ReturnType<typeof makeStore>;
export type RootState = ReturnType<AppStore["getState"]>;
export type AppDispatch = AppStore["dispatch"];
