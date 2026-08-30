import { createReauthBaseQuery, IDENTITY_API_BASE } from "@/lib/api/createReauthBaseQuery";

export const baseQueryWithReauth = createReauthBaseQuery(IDENTITY_API_BASE);
