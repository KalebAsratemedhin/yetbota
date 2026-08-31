# YetBota — Frontend docs

Decisions and structure for the Next.js app under `frontend/`.

## Stack

| Choice | Value | Notes |
| --- | --- | --- |
| Framework | Next.js 16 (App Router) | TypeScript, Tailwind CSS v4, `src/` directory |
| Package manager | npm | Matches create-next-app default for this repo |
| UI | shadcn/ui (`base-nova`, CSS variables, Lucide) | Components live under `src/components/ui/` |
| Font | Plus Jakarta Sans | Via `next/font/google`; applied on `<body>` |
| Theme | Light + dark (`.dark` on `<html>`) | Default **dark** when `localStorage['yetbota.theme']` unset |
| Design tokens | Brand/surface/fg in `globals.css` | From `practices/design.inspo-1.md` / prior Yet Bota web |
| Client state / API | Redux Toolkit + RTK Query | `makeStore` per client tree; no module-level store singleton |
| Landing copy | `content/en.json` via `useContent()` | English |
| Auth session | Redux `auth` + `localStorage` (`yetbota.auth`) | access + refresh; refresh-on-401 |
| Identity API | Direct to service URL | CORS on identity-service; default `http://localhost:8081` |
| Content API | Direct to content-service | CORS on content-service; default `http://localhost:8082` |
| App location | `frontend/` at repo root | Mirrors `backend/` |

## Auth (identity-service)

Contract: `backend/identity-service/api/openapi.yaml`.

| Concern | Behavior |
| --- | --- |
| Signup | `POST /auth/signup` `{ username, password, full_name }` → store tokens → `/` |
| Signin | `POST /auth/signin` `{ username, password }` → store tokens → `?next=` (relative) or `/` |
| Session | `accessToken` + `refreshToken` (+ user id/username) in Redux and `yetbota.auth` |
| Refresh | On **401** from protected calls: `POST /auth/refresh` with refresh token, rotate pair, retry once |
| Logout | `POST /auth/logout` (Bearer + refresh body) then clear local session |
| Me | `GET /auth/me` available via `useMeQuery` |
| Social | Follow graph + search via `socialApi`; avatar via `uploadAvatar` / `deleteAvatar` |

## Content (content-service)

Contract: `backend/content-service/api/openapi.yaml`.

| Concern | Behavior |
| --- | --- |
| Base URL | `NEXT_PUBLIC_CONTENT_API_BASE_URL` (default `http://localhost:8082`); separate RTK `contentApi` |
| Auth | Same Bearer access token; 401 → identity refresh → retry |
| Create | Sign media → Cloudinary upload → `POST /posts` |
| Read | `GET /posts/{id}`, `GET /users/{id}/posts`, bookmarks, search, hashtags |
| Feed | `GET /feed?lat=&lng=` — following + nearby; signed-in `/` uses geolocation |
| Engage | Reactions, comments (incl. replies + comment reactions), bookmark, share |

## Reputation (identity-service)

Exposed on `UserInfo` from `/auth/me` and public profiles: `reputation_score`, `reputation_level`, `badges[]`. Levels: Explorer → Contributor (50) → Local Guide (200) → Pathfinder (500).

## Why these choices

- **App Router** — current Next.js default; official Redux + shadcn docs target it.
- **Token set reused** — brand green `#22C55E` / `#00e05a`, `bg-bg`, `surface*`, `fg*` so marketing + auth match the prior product look.
- **localStorage session** — simple for SPA + RTK; httpOnly cookies would need a BFF later.
- **Refresh on 401** — matches short-lived access JWT; no proactive timer.
- **Signed-in home = feed** — marketing landing stays for guests; authenticated `/` requests location and loads ranked feed.

## Routes

| Route | Notes |
| --- | --- |
| `/` | Guest: landing. Signed-in: home feed (geo + following/nearby) |
| `/signin` | Wired to identity signin |
| `/signup` | Wired to identity signup (full name + username + password) |
| `/profile` | Own profile; reputation/badges + Locations + Saved |
| `/users/[id]` | Public profile; reputation/badges + Locations |
| `/search` | People + Places tabs; hashtag shortcut |
| `/compose` | Create post (media modes + geo) |
| `/posts/[id]` | Post detail, reactions, comments, bookmark, share |
| `/hashtags/[tag]` | Posts for one hashtag |
| `/forgot-password` | Stub (“reset unavailable”) |

## Layout

```
frontend/
  content/en.json
  public/images/
  src/
    app/
      page.tsx
      (auth)/
    components/
      landing/
      auth/
      providers/
      ui/
    lib/
      api/baseApi.ts
      api/baseQueryWithReauth.ts
      api/authApi.ts
      api/socialApi.ts
      auth/authSlice.ts
      auth/types.ts
      store.ts
      validation/phone.ts
```

## Config knobs

- `NEXT_PUBLIC_API_BASE_URL` — identity base URL (default `http://localhost:8081`).
- `NEXT_PUBLIC_CONTENT_API_BASE_URL` — content base URL (default `http://localhost:8082`).
- `yetbota.theme` / `yetbota.auth` in localStorage.
- shadcn: `frontend/components.json`.
- Design reference: `practices/design.inspo-1.md`.

## Run

```bash
cd frontend && npm run dev
# identity-service on :8081 with CORS allowing the frontend origin
```

## Still TBD

Embla champions, live trending places, theme toggle UI, deploy.
