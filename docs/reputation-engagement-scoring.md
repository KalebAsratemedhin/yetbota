# Research: engagement-weighted reputation with freshness

**Status:** Option 1 evolved — `post_interaction_events` log + per-event gravity decay (2026-09-04).  
**Date:** 2026-09-03 (research); 2026-09-04 (implementation)  
**Scope:** How likes, comments, shares, and saves can be combined mathematically, with time freshness, into a **user reputation** signal for YetBota (crowdsourced location discovery).  
**Out of scope:** ML recommenders, creator monetization.

---

## 1. Problem framing

YetBota reputation v1 awards fixed ledger deltas:

| Event | Delta |
| --- | --- |
| `post.created` | +2 |
| `post.liked` (received) | +1 |
| `post.unliked` | −1 |

That ignores **comment**, **share**, and **bookmark/save** volume, treats every like equally forever, and does not care **how fresh** the engagement is.

The goal of this note is to document how other systems value **multi-signal engagement** and **freshness**, then propose math that fits YetBota without copying platforms whose product goals differ (ads, watch time, Q&A privileges).

**Desired property:** a user’s score should rise when their *posts* earn meaningful interactions, with higher-effort interactions counting more, and recent evidence counting more than stale evidence.

---

## 2. Signal hierarchy (what industry treats as “strong”)

Across open-sourced and publicly described ranking stacks, **effort and intent** dominate raw tap counts.

### 2.1 Typical relative strength (qualitative)

| Signal | Relative strength | Why |
| --- | --- | --- |
| Like / reaction | Weakest | Lowest friction; easy to farm |
| Comment | Medium–high | Requires composition; starts conversation |
| Save / bookmark | High | Private “I need this later” — lasting utility |
| Share / send / repost | Highest among positive public acts | User spends social capital to recommend |

Industry practice (marketing analyses + open sources; exact coefficients are never public and change often):

- **Instagram (Mosseri, 2025+):** sends/shares and watch time dominate distribution; saves are strong “lasting value”; likes still matter but are demoted relative to sends (often cited as shares ~3–5× likes for unconnected reach; some analyses claim higher).
- **X / Twitter open algorithm:** final score is a **weighted sum of predicted engagement probabilities**, not raw counts. Historical Heavy Ranker weights (public 2023 snapshot) put replies far above favorites (e.g. reply ~27× favorite, author-engaged reply ~150× favorite); reports and “not interested” are large negatives.
- **Analyst / growth tooling:** common hand-tuned marketing weights for *reporting* engagement quality (not ranking): like 1×, comment 2×, share 3×, save 3× (sometimes link-click 4×).

**YetBota mapping:** we already store `like_count`, `dislike_count`, `comment_count`, `share_count`, `bookmark_count` on posts. That is enough for a **deterministic weighted engagement** without ML.

---

## 3. Core math: post engagement score

### 3.1 Weighted engagement (base)

Let for a post \(p\):

\[
E(p) = w_L L + w_C C + w_S S + w_B B - w_D D
\]

| Symbol | Meaning | Suggested v1 weight |
| --- | --- | --- |
| \(L\) | like count | \(w_L = 1\) |
| \(C\) | comment count | \(w_C = 2\)–\(4\) |
| \(S\) | share count | \(w_S = 3\)–\(5\) |
| \(B\) | bookmark/save count | \(w_B = 3\)–\(4\) |
| \(D\) | dislike count (optional) | \(w_D = 1\)–\(2\) |

**Notes:**

- Weights should be **config**, not hard-coded forever.
- Prefer **unique actors** (distinct users who liked/commented/…) over raw counts when the data model allows; raw counts are fine for v1 if uniqueness is expensive.
- Cap or log-compress extreme counts so one viral post cannot mint infinite reputation (see §5).

### 3.2 Sublinear compression (anti-viral monopoly)

Raw \(E\) grows linearly with spam rings and lucky virality. Industry almost always compresses:

**Option A — power (HN-style)**

\[
E' = (E + 1)^{\alpha}, \quad 0 < \alpha < 1
\]

HN uses \(\alpha \approx 0.8\) on points. Softens big numbers without discarding them.

**Option B — log (Reddit hot–style)**

\[
E' = \mathrm{sign}(E)\cdot\log_{10}(\max(|E|, 1))
\]

First interactions matter most; later ones add less.

**Recommendation for YetBota:** **Option A** with \(\alpha = 0.8\), aligned with the existing feed ranker.

---

## 4. Freshness / time decay

“Freshness of the data” can mean two different clocks. Pick explicitly.

### 4.1 Clock A — **post age** (content freshness)

Score decays as the post gets older. Classic **Hacker News gravity**:

\[
\mathrm{rank}(p) = \frac{(P-1)^{0.8}}{(T+2)^{G}}
\]

- \(P\) = points (votes)  
- \(T\) = age in hours  
- \(G\) = gravity (HN historically ~1.8; YetBota feed already uses **1.5**)

Effect: even strong posts fall; the front page stays alive.

YetBota **feed** already uses a close form:

\[
\mathrm{feed\_score}(p) = \frac{(E_{\mathrm{feed}}+1)^{0.8}}{(T+2)^{G}} \cdot \mathrm{follow\_boost} \cdot \mathrm{geo\_factor}
\]

with \(E_{\mathrm{feed}} = L + w_C C + w_S S\) (bookmarks not in feed formula today).

### 4.2 Clock B — **engagement age** (evidence freshness)

Each interaction has a timestamp. Recent likes/comments weigh more than year-old ones. Typical form:

\[
E_{\mathrm{fresh}}(p) = \sum_{i \in \mathrm{events}(p)} w_{\mathrm{type}(i)}\, e^{-t_i / \tau}
\]

- \(t_i\) = age of interaction \(i\) in hours (or days)  
- \(\tau\) = half-life scale (e.g. \(\tau = 24\cdot 14\) for ~two-week half-life via \(e^{-\ln 2\cdot t/t_{1/2}}\))

**Exponential decay** is what review platforms describe for “recent reviews matter more” (e.g. Google Maps–style recency weighting in star aggregates — exact \(\tau\) unpublished).

Cost: needs an **event stream** (or periodic recompute from `reactions` / `comments` / `shares` / `bookmarks` tables), not just denormalized counters.

### 4.3 Clock C — **velocity** (early momentum)

X-style feeds care about engagement in the **first ~30 minutes**. Velocity is better for **feed ranking** than for long-lived reputation. For reputation, prefer Clock A or B so a good place post can keep earning slowly without being forgotten after day one.

### 4.4 Reddit “hot” (alternative time embedding)

\[
\mathrm{hot} = \mathrm{sign}(s)\cdot\log_{10}(\max(|s|,1)) + \frac{t_{\mathrm{unix}} - t_0}{45000}
\]

Time is an **additive** offset, not gravity in the denominator. Good for forums; less natural when you want “old but useful place posts” to still contribute some reputation.

### 4.5 Wilson score / Bayesian average (confidence, not freshness)

- **Wilson lower bound** (Evan Miller; Reddit comment “best” sort): ranks by conservative estimate of upvote *ratio*, so 10/10 beats 1000/600 early noise. Useful for **quality ratio**, not for “how much engagement.”  
- **Bayesian average:** pull sparse averages toward a prior until enough samples. Useful if reputation were an *average rating*; YetBota is closer to a **contribution mass** problem.

These are complements, not substitutes for weighted \(E\) + decay.

---

## 5. From post score → **user reputation**

Engagement formulas above score **posts**. Reputation needs an **author aggregate**.

### 5.1 Pattern used in practice

| Pattern | Idea | Pros | Cons |
| --- | --- | --- | --- |
| **A. Event ledger (current YetBota)** | Each like fires `+1` to author | Auditable, real-time | Hard to encode freshness; ignores comments/shares/saves unless many event types |
| **B. Snapshot of weighted posts** | Periodically \( R(u) = \sum_{p \in u} f(E(p), age(p)) \) | Matches “math on interactions + freshness” cleanly | Needs recompute job; less “ledger purity” |
| **C. Hybrid** | Ledger stores raw engagement events; score view applies weights + decay at read/recompute | Best of both | More moving parts |

### 5.2 Recommended hybrid for YetBota

**Post instantaneous quality** (same spirit as feed):

\[
Q(p) = \frac{(E(p)+1)^{\alpha}}{(T_p + T_0)^{G}}
\]

with \(E\) from §3.1 including bookmarks; \(T_p\) = post age hours; \(T_0 = 2\); \(\alpha = 0.8\); \(G \in [1.2, 1.8]\).

**Author reputation** (freshness-aware sum with diminishing returns across posts):

\[
R(u) = \sum_{p \in \mathrm{posts}(u)} Q(p)^{\beta}
\quad\text{or}\quad
R(u) = \sum_{p} \log(1 + Q(p))
\]

with \(\beta < 1\) (e.g. 0.5) so one mega-post cannot dominate a consistent local contributor.

**Optional lookback:** only include posts with \(T_p \le W\) days (e.g. 90) in the *live* score, while badges/history remain permanent. That keeps reputation “about recent usefulness,” which matches location discovery better than lifetime like totals.

### 5.3 Ledger-friendly approximation (if we keep event-driven)

Instead of one `post.liked → +1`, emit or recompute deltas from engagement *changes*:

\[
\Delta R \approx w_{\mathrm{type}} \cdot \underbrace{\frac{1}{(T+T_0)^{G}}}_{\text{freshness at event time}}
\]

- Like received while post is 1h old → larger \(\Delta R\)  
- Like received at day 30 → smaller \(\Delta R\)  
- Comment/share/save use higher \(w_{\mathrm{type}}\)  
- Unlike / delete comment / unshare / unbookmark → reverse with same formula (or snapshot recompute)

Daily cap can remain, but should apply to **positive \(\Delta R\)** after weighting.

### 5.4 What **not** to copy blindly

| Source | Why it doesn’t transfer 1:1 |
| --- | --- |
| X Phoenix / Heavy Ranker | Predicts *viewer-specific* P(action); reputation should be author-global and explainable |
| Instagram watch-time dominance | YetBota is place discovery, not short-form video sessions |
| Stack Overflow privilege ladder | Different product; still useful for *caps* and *no self-award* |
| Paying large fixed points for `post.created` | Rewards spam volume; prefer small participation credit + engagement-weighted quality |

---

## 6. Worked example

Assume weights \(w_L=1, w_C=3, w_S=4, w_B=3\), \(\alpha=0.8\), \(G=1.5\), \(T_0=2\).

**Post A** (6 hours old): 20 likes, 4 comments, 2 shares, 5 saves  

\[
E = 20 + 3\cdot4 + 4\cdot2 + 3\cdot5 = 55,\quad
Q = \frac{56^{0.8}}{(6+2)^{1.5}} \approx \frac{23.6}{22.6} \approx 1.04
\]

**Post B** (48 hours old): same engagement  

\[
Q = \frac{56^{0.8}}{(48+2)^{1.5}} \approx \frac{23.6}{353.6} \approx 0.067
\]

Same interactions are worth ~**15× less** when two days older under this gravity. That is the “freshness of the data” effect in one number.

If author \(u\) has only these two posts and \(R = Q_A + Q_B\):

\[
R(u) \approx 1.04 + 0.067 \approx 1.11
\]

(Scale \(R\) by a constant for human-readable levels, e.g. multiply by 100.)

---

## 7. Relation to YetBota today

| Layer | Today | Gap vs this research |
| --- | --- | --- |
| **Feed ranking** | Weighted \(L,C,S\) + HN gravity + follow/geo | Bookmarks not in feed \(E\); already has freshness |
| **Reputation** | Fixed +2 create / +1 like | No comments/shares/saves; no freshness; create-heavy |
| **Data** | Counters on `posts` | Enough for snapshot \(Q(p)\); event-age decay needs timestamps per interaction |

**Practical takeaway:** YetBota already solved half the math for **posts** (feed). Reputation should **reuse the same engagement + gravity ideas**, aggregate per author, and stop treating “got a like sometime” as +1 forever.

---

## 8. Design options (for a future SRS decision)

### Option 1 — Minimal (recommended first step)

1. Define \(E(p)\) with likes, comments, shares, bookmarks (± dislikes).  
2. Define \(Q(p)\) with feed-aligned gravity.  
3. Nightly or on-engagement job: \(R(u) = \sum \log(1+Q(p))\) over posts in lookback window; write denormalized `reputation_score`.  
4. Keep badges; retune level bands to the new scale.  
5. Shrink or remove `post.created` fixed points (or keep a tiny +1 once).

### Option 2 — Event ledger with freshness-at-award

On each engagement change, award \(w_{\mathrm{type}} / (T+T_0)^G\) to the author (capped daily). Reversible. Closer to current architecture.

### Option 3 — Full event-age exponential decay

Recompute \(E_{\mathrm{fresh}}\) from interaction timestamps (§4.2). Most accurate “freshness of the data”; highest engineering cost.

---

## 9. Anti-abuse checklist (any option)

- No self-engagement credit (already true for likes).  
- Prefer unique actors when possible.  
- Daily / weekly positive earn caps.  
- Sublinear \(E\) and/or per-post contribution to \(R\).  
- Clawback on delete, unlike, unbookmark, removed share.  
- Optional: ignore engagements from accounts newer than \(N\) hours (cold-start farming).

---

## 10. Sources (selected)

- Hacker News gravity / frontpage rank (news.arc analyses; Ken Shirriff; amix HN write-ups).  
- Reddit hot + Wilson confidence sort (`reddit-archive` `_sorts.pyx`; Evan Miller, “How Not To Sort By Average Rating”).  
- X algorithm: weighted sum of engagement probabilities (twitter/the-algorithm-ml Recap README, 2023 weights; xai-org/x-algorithm Phoenix + RankingScorer descriptions).  
- Instagram ranking signal hierarchy (Adam Mosseri public statements 2025+; industry summaries on sends/saves vs likes).  
- Weighted engagement reporting practice (like/comment/share/save multipliers in growth analytics literature).  
- YetBota internal: `docs/srs.md` feed formula + reputation defaults; content-service feed `rankScore`.

---

## 11. Suggested next step

If product agrees with **Option 1**, turn §8 into SRS FRs (event types / recompute cadence / weights / lookback / level retune) before changing identity ledger code. Keep feed and reputation formulas **parameter-aligned** (\(w_C, w_S, G, \alpha\)) so “what the feed thinks is good” and “what reputation rewards” do not diverge.
