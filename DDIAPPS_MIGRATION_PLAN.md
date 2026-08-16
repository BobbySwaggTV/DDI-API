# DDIApps Phase 4.1 Migration Plan

> Reference: `DDIAPPS_INTEGRATION_AUDIT.md`  
> Scope: configure CavApps to consume the DDI API without rewriting the frontend architecture or the BFF shape.  
> No code changes.

## 1. Current API Communication

### 1.1 Frontend → BFF

The Next.js client only talks to the Express BFF on `localhost:4000` (or the deployed BFF host). All client calls are `fetch` with an `Authorization` header set to `NEXT_PUBLIC_CLIENT_TOKEN` <ref_file file="/home/bobbyswagg/research/CavApps/client/app/reusableModules/getCombatRoster.jsx" />. The BFF `checkToken` middleware compares that header verbatim to `CLIENT_TOKEN` <ref_snippet file="/home/bobbyswagg/research/CavApps/server/server.js" lines="35-46" />.

### 1.2 BFF → upstream API

The BFF uses `axios` to call `https://api.7cav.us` with `Authorization: Bearer {API_TOKEN}` <ref_file file="/home/bobbyswagg/research/CavApps/server/controllers/cacheManager.js" />. Four upstream endpoints are used:

* `GET /api/v1/roster/1` (combat)
* `GET /api/v1/roster/2` (reserve)
* `GET /api/v1/milpacs/profile/username/{username}`
* `GET /api/v1/milpacs/position/groups`

The client is insulated from the upstream; only the BFF needs to repoint.

## 2. Configuration Values That Reference 7CAV

| Config key / hardcoded value | Location | What it does |
| --- | --- | --- |
| `https://api.7cav.us` | `server/controllers/cacheManager.js` (4 calls) | upstream API base URL |
| `https://7cav.us/data/roster_ranks/...` | `server/db/milpacDb.js` | hardcoded rank image URL builder |
| `CLIENT_TOKEN` / `NEXT_PUBLIC_CLIENT_TOKEN` | `.env.example`, `server/server.js`, client fetch helpers | shared secret between client and BFF |
| `API_TOKEN` | `.env.example`, `server/controllers/cacheManager.js` | bearer token for upstream API |
| `apps.7cav.us`, `appsbeta.7cav.us`, `localhost` | `server/server.js` CORS allowlist | allowed client origins |
| `https://7cav.us/forums/` | `client/app/page.jsx` | branding link |
| `https://7cav.us/tickets/categories/2/create` | `client/app/not-found.jsx` | support ticket link |
| `https://7cav.us/rosters/profile/{item.itemKey}` | `client/app/adr/modules/ArrayMap.jsx` | ADR profile link |
| `"Please enter a 7Cav Username..."` | `client/app/uniformbuilder/page.jsx` | placeholder text |
| `ROSTER_TYPE_*` (7CAV names) | `server/routes/diffRoutes.js`, `server/controllers/diffPoller.js`, `client/app/rosterhistory/lib/constants.js` | roster history wire-name constants |
| `RANK_SHORT` table | `server/db/milpacDb.js` | maps `rank_id` to short abbreviation for the diff poller |
| `awardCatalog.js` names | `client/app/uniformbuilder/modules/constants/awardCatalog.js` | award catalog for the uniform builder |
| `mos.js` | `client/app/uniformbuilder/modules/constants/mos.js` | 7CAV MOS codes and wiki link |
| `CavColor` / logo SVGs | `client/app/theme/cavColor.js`, `client/public/CavAppsLogo.svg` | branding assets |

<ref_file file="/home/bobbyswagg/research/DDI-API/DDIAPPS_INTEGRATION_AUDIT.md" />

## 3. Required Environment Variable Changes

### 3.1 Server (`.env`)

```dotenv
# New: DDI API host (no trailing slash)
# If absent, default could stay empty — but the code currently has no default.
API_BASE_URL=https://api.ddi.example

# Existing token names stay; values change to DDI-issued credentials
API_TOKEN=ddi-api-bearer-token
CLIENT_TOKEN=any-shared-secret-for-client

# Optional: if DDI still uses the same MariaDB schema, keep these.
# If DDI runs without the diff feature, these can remain unset.
# XENFORO_DB_HOST=
# XENFORO_DB_PORT=3306
# XENFORO_DB_USER=
# XENFORO_DB_PASSWORD=
# XENFORO_DB_NAME=xenforo
```

### 3.2 Client (`client/.env.local`)

No client-side env changes are strictly required. The client already points to the BFF via `NEXT_PUBLIC_*` URLs, and the BFF-to-upstream change is transparent. The only client-side value that changes is branding, not URLs.

```dotenv
NEXT_PUBLIC_CLIENT_TOKEN=any-shared-secret-for-client
COMBAT_API_URL=http://localhost:4000/roster/combat
RESERVE_API_URL=http://localhost:4000/roster/reserves
GROUP_API_URL=http://localhost:4000/roster/groups
CACHE_TIMESTAMP_URL=http://localhost:4000/cache-timestamp
NEXT_PUBLIC_INDIVIDUAL_API_URL=http://localhost:4000/roster/individual
NEXT_PUBLIC_DIFF_API_URL=http://localhost:4000
NEXT_PUBLIC_USERCACHE_API_URL=http://localhost:4000/userSearch
```

### 3.3 Docker / compose

`docker-compose.yml` and `docker-compose.dev.yml` reference `7cav/...` images in comments and the `edge` network. The runtime `server` service only needs `API_BASE_URL` and `API_TOKEN` injected.

## 4. Minimum Changes Required to Point DDIApps to DDI API

The goal is the smallest set of edits that makes the BFF consume the DDI API. Anything cosmetic can wait.

### 4.1 Must change for the app to work

1. **Upstream base URL**
   * Read `API_BASE_URL` from env in `server/controllers/cacheManager.js`.
   * Replace the four hardcoded `https://api.7cav.us` calls with `${API_BASE_URL}/api/v1/roster/1`, etc.
   * If `API_BASE_URL` is not present, fail fast on startup so the old upstream is never accidentally used.

2. **CORS allowlist**
   * Replace the two `apps.7cav.us` and `appsbeta.7cav.us` entries in `server/server.js` with DDI app domains (`apps.ddi.example`).
   * Keep `localhost` and `localhost:3000` for local development.

3. **Upstream token**
   * Set `API_TOKEN` to the DDI API bearer. No code change is required; `cacheManager.js` already reads `process.env.API_TOKEN`.

4. **Roster wire names (if roster history / diff feature is used)**
   * `server/routes/diffRoutes.js`: update `VALID_ROSTER_TYPES` to the DDI names: `ROSTER_TYPE_ACTIVE`, `ROSTER_TYPE_RESERVE`, `ROSTER_TYPE_EXTENDED_LEAVE`, `ROSTER_TYPE_HONOR_ROLL`, `ROSTER_TYPE_MEMORIAL`, `ROSTER_TYPE_FORMER_MEMBERS`.
   * `server/controllers/diffPoller.js`: update `ROSTER_TYPES` and `ROSTER_ID_MAP` similarly.
   * `client/app/rosterhistory/lib/constants.js`: update `ALL_ROSTER_TYPES`, `ROSTER_TYPE_LABELS`, and `ROSTER_TYPE_COLORS`.

5. **Rank short table**
   * Update `server/db/milpacDb.js` `RANK_SHORT` to match the DDI API's `RANK_SHORT_NAMES` default or override. This only matters if DDI overrides rank short names.

### 4.2 Should change for correctness / safety

6. **Rank image URL builder**
   * Replace `https://7cav.us` in `server/db/milpacDb.js` `rankImageUrl()` with the DDI asset host (which could be the `ASSET_BASE_URL` env). If DDI has the same path layout under a different base URL, the rest of the template can stay the same.

7. **Record type labels (only if DDI overrides `RECORD_TYPE_NAMES`)**
   * Update `client/app/rosterhistory/lib/constants.js` `RECORD_TYPE_LABELS` if DDI changes record wire names.

### 4.3 Cosmetic changes (can be deferred)

8. Replace 7CAV branding links in `client/app/page.jsx`, `client/app/not-found.jsx`, `client/app/adr/modules/ArrayMap.jsx`, and `client/app/uniformbuilder/page.jsx`.
9. Replace `CavColor` / `CavAppsLogo.svg` with DDI logos.
10. Update `awardCatalog.js` and `mos.js` only if DDI awards / MOS values differ.
11. Update README, `.env.example` comments, and workflow files.

## 5. Risks

### 5.1 Authentication

* **DDI API key format**: The DDI API accepts `Authorization: Bearer <key>` <ref_snippet file="/home/bobbyswagg/research/DDI-API/rest/tickets_test.go" lines="28-30" />, which matches the BFF's current header. The only risk is an incorrect or expired token.
* **Client token scheme**: The BFF compares `req.headers['authorization']` to `CLIENT_TOKEN` exactly, not as a `Bearer` value. If `NEXT_PUBLIC_CLIENT_TOKEN` and `CLIENT_TOKEN` do not match exactly, the client is rejected. This is unchanged and safe.
* **Token exposure**: `cacheManager.js` already strips the token from error logging; keep that behavior.

### 5.2 Response compatibility

* **Roster JSON shape**: The DDI API preserves `profiles: { [relationId]: Profile }` with `rank`, `primary`, `secondaries`, `records`, `awards`, `realName`, `user`, `roster`, etc. No known shape changes.
* **Roster wire names**: This is the biggest risk. `client/app/rosterhistory/lib/constants.js` and the BFF diff routes compare `roster` string values. After the DDI API changes, `ROSTER_TYPE_COMBAT` is now `ROSTER_TYPE_ACTIVE`. If the diff feature is enabled, those constants must be updated before the first snapshot or the diff engine will treat all profiles as transfers.
* **Rank short names**: `server/db/milpacDb.js` `RANK_SHORT` is used for rank ordering in the diff engine and the `/ranks` endpoint. If DDI overrides `RANK_SHORT_NAMES` to non-7CAV values, this table must match or the diff sort order will be wrong.
* **Record type names**: DDI keeps the 7CAV record wire names by default. If `RECORD_TYPE_NAMES` is overridden, the `RECORD_TYPE_LABELS` map must follow.

### 5.3 Uniform builder compatibility

* **Award names**: `getCanvasObject.jsx` matches `awardName` against `awardCatalog.js`. If the DDI API returns different award titles, unknown awards are already skipped silently, but the user will see missing decorations. The catalog must be updated to DDI award names if they differ.
* **MOS values**: `mos.js` expects 7CAV MOS codes. DDI can either keep the same `mos` custom-field values or update `mos.js` and the shoulder-cord/neck-pin logic.
* **Rank grades**: `GetUserInfo.jsx` maps `rankId` 1–31 to `O11`/`O10`/`E1` grades. As long as DDI keeps the same `rank_id` values, the local `skunkworks/uniformEpaulettes/{rankGrade}.png` images will load. If DDI changes `rank_id` values, the map and images must be updated.
* **Rank image URL**: The canvas does not use `rankImageUrl` from the API; it uses local epaulette PNGs. The `server/db/milpacDb.js` `rankImageUrl()` is only used for the `/ranks` endpoint payload and can be repointed safely.

## 6. Recommended Migration Order

1. **No-diff smoke test first**
   * Disable the roster-history feature (or ignore `/diffs`).
   * Set `API_BASE_URL` + `API_TOKEN` and repoint `cacheManager.js`.
   * Update CORS.
   * Verify ADR, Roster Statistics, and Uniform Builder render correctly.

2. **Roster history diff second**
   * Update `diffRoutes.js`, `diffPoller.js`, and `client/app/rosterhistory/lib/constants.js` to DDI roster wire names.
   * Clear the Postgres snapshot tables and let the diff engine bootstrap from DDI data.

3. **Branding and data third**
   * Replace 7CAV URLs and logos.
   * Update `awardCatalog.js` and `mos.js` if DDI data differs.

## 7. What Not to Do

* Do not rewrite the client as a direct consumer of the DDI API — the BFF layer is the intended abstraction.
* Do not change the BFF endpoint paths (`/roster/combat`, `/roster/reserves`, etc.) unless the client must be touched.
* Do not modify the diff database schema; the existing `xf_nf_rosters_*` shapes are already aligned.
* Do not touch the canvas sprite-sheet architecture unless DDI awards/ranks are materially different.