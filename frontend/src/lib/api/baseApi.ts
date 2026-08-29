import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "@/lib/api/baseQueryWithReauth";

/**
 * Shared RTK Query API. Feature endpoints are injected from modules such as authApi.
 */
export const baseApi = createApi({
  reducerPath: "api",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Me"],
  endpoints: () => ({}),
});
