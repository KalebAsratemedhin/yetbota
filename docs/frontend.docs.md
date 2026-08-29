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

## Why these choices

- **App Router** — current Next.js default; official Redux + shadcn docs target it.
- **Token set reused** — brand green `#22C55E` / `#00e05a`, `bg-bg`, `surface*`, `fg*` so marketing + auth match the prior product look.
- **localStorage session** — simple for SPA + RTK; httpOnly cookies would need a BFF later.
- **Refresh on 401** — matches short-lived access JWT; no proactive timer.
- **Mocked Discovery / Champions** — no content/users APIs yet; static cards + CSS marquee (Embla deferred).

## Routes

| Route | Notes |
| --- | --- |
| `/` | Landing (navbar → hero → … → footer) |
| `/signin` | Wired to identity signin |
| `/signup` | Wired to identity signup (full name + username + password) |
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
      auth/authSlice.ts
      auth/types.ts
      store.ts
```

## Config knobs

- `NEXT_PUBLIC_API_BASE_URL` — identity base URL (default `http://localhost:8081`).
- `yetbota.theme` / `yetbota.auth` in localStorage.
- shadcn: `frontend/components.json`.
- Design reference: `practices/design.inspo-1.md`.

## Run

```bash
cd frontend && npm run dev
# identity-service on :8081 with CORS allowing the frontend origin
```

## Still TBD

Route guards / middleware, profile edit, Embla champions, live trending places, theme toggle UI, deploy.
