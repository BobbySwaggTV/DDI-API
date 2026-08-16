# DDIApps Integration Audit

> Scope: CavApps / DDIApps frontend repository at `/home/bobbyswagg/research/CavApps`  
> Goal: determine the minimum changes a DDI-branded CavApps needs to consume the new DDI API.  
> No code was modified.

## 1. Frontend Framework and Structure

CavApps is a monorepo with two parts:

* **Client** — Next.js 16.3 (App Router, React 19) in `client/` <ref_file file="/home/bobbyswagg/research/CavApps/client/package.json" />
* **Server** — Express 5.2 "BFF" (backend-for-frontend) on port 4000 in `server/` <ref_file file="/home/bobbyswagg/research/CavApps/server/package.json" />

The client never calls the 7CAV/DDI API directly. It calls the BFF, which fetches upstream data and caches it in memory. The BFF also runs a cron-based roster-history diff poller against a local Postgres database and (optionally) the live XenForo MariaDB <ref_snippet file="/home/bobbyswagg/research/CavApps/README.md" lines="35-38" />.

### Client pages

* `client/app/adr/page.jsx` — Active Duty Roster (ADR)
* `client/app/rosterstatistics/page.jsx` — Roster statistics
* `client/app/uniformbuilder/page.jsx` — Uniform builder
* `client/app/rosterhistory/page.jsx` — Roster history / diff viewer

### Server controllers

* `server/controllers/cacheManager.js` — pulls upstream roster/profile/groups and keeps them in memory
* `server/controllers/diffPoller.js` — snapshots rosters and writes diff events to Postgres
* `server/controllers/diffEngine.js` — computes diffs between roster snapshots
* `server/db/milpacDb.js` — direct XenForo MariaDB reads for the diff poller

## 2. API Client Implementation

### 2.1 Upstream (DDI API) client

`server/controllers/cacheManager.js` uses `axios` to call the upstream API with a bearer token <ref_file file="/home/bobbyswagg/research/CavApps/server/controllers/cacheManager.js" />:

| Upstream call | DDI-API equivalent | Purpose |
| --- | --- | --- |
| `https://api.7cav.us/api/v1/roster/1` | `GET /api/v1/roster/ROSTER_TYPE_ACTIVE` or `/api/v1/roster/1` | full combat roster |
| `https://api.7cav.us/api/v1/roster/2` | `GET /api/v1/roster/ROSTER_TYPE_RESERVE` or `/api/v1/roster/2` | full reserve roster |
| `https://api.7cav.us/api/v1/milpacs/profile/username/{username}` | `GET /api/v1/milpacs/profile/username/{username}` | single profile |
| `https://api.7cav.us/api/v1/milpacs/position/groups` | `GET /api/v1/milpacs/position/groups` | position groups |

The upstream base URL and the numeric roster IDs `1` and `2` are hardcoded.

### 2.2 Client-to-BFF client

The Next.js client uses `fetch` with an `Authorization` header equal to `NEXT_PUBLIC_CLIENT_TOKEN` <ref_file file="/home/bobbyswagg/research/CavApps/client/app/reusableModules/getCombatRoster.jsx" />. The BFF matches `req.headers['authorization']` exactly to `CLIENT_TOKEN` in `server/server.js` <ref_snippet file="/home/bobbyswagg/research/CavApps/server/server.js" lines="35-46" />.

## 3. Current API Base URL Configuration

| File | Current value | Notes |
| --- | --- | --- |
| `server/controllers/cacheManager.js` | `https://api.7cav.us` | hardcoded in four `axios` calls |
| `server/db/milpacDb.js` | `https://7cav.us/data/roster_ranks/...` | hardcoded rank image URL builder |
| `client/.env` (via `README.md`) | `COMBAT_API_URL`, `RESERVE_API_URL`, `GROUP_API_URL`, `NEXT_PUBLIC_INDIVIDUAL_API_URL`, etc. | all point to `http://localhost:4000` BFF routes |
| `.env.example` | `API_TOKEN=your-7cav-api-token-here` | the server uses this as the upstream bearer |

The minimal change is to make the upstream base URL configurable via `process.env.API_BASE_URL` (or similar), defaulting to the DDI API host.

## 4. Authentication / Token Handling

### 4.1 Upstream API

The BFF sends `Authorization: Bearer {API_TOKEN}` to the upstream API. The DDI API already uses the same `Authorization: Bearer <key>` scheme <ref_snippet file="/home/bobbyswagg/research/DDI-API/rest/tickets_test.go" lines="28-30" />, so only the token value needs to change. No code change is required if `API_TOKEN` is supplied as an env var.

### 4.2 Client-to-BFF

The client sends `Authorization: {NEXT_PUBLIC_CLIENT_TOKEN}` and the BFF compares the raw header value to `CLIENT_TOKEN`. This is a shared-secret, not a bearer scheme. It is independent of the upstream API key.

### 4.3 CORS

`server/server.js` hardcodes an allowlist of `localhost` and `*.7cav.us` origins. DDI will need to replace the 7CAV origins with DDI app domains.

## 5. Required API Endpoints

The BFF currently requires these endpoints from the DDI API:

1. `GET /api/v1/roster/1` — full roster (combat)
2. `GET /api/v1/roster/2` — full roster (reserve)
3. `GET /api/v1/milpacs/profile/username/{username}` — full profile
4. `GET /api/v1/milpacs/position/groups` — position groups

Because the DDI API keeps the numeric roster IDs `1` and `2` and the same path templates, the BFF's upstream calls can keep the numbers and only swap the host. If DDI later wants to use the new name forms (`ROSTER_TYPE_ACTIVE`, `ROSTER_TYPE_RESERVE`), the BFF could also use those instead of numbers.

The BFF exposes these downstream endpoints to the client:

* `GET /roster/combat`
* `GET /roster/reserves`
* `GET /roster/individual?username=...`
* `GET /roster/groups`
* `GET /cache-timestamp`
* `GET /diffs`, `GET /diffs/:date`, `GET /diffs/range`
* `GET /ranks`
* `GET /userSearch?q=...`
* `POST /admin/snapshot`, `GET /admin/runs`

None of these need to change as long as the BFF's upstream data stays compatible.

## 6. JSON Response Shapes Consumed

### 6.1 Roster (`/api/v1/roster/{roster}`)

Shape: `{ profiles: { [relationId]: Profile } }`

`Profile` fields used by the client:

| Field | Used in | Notes |
| --- | --- | --- |
| `user.username` | ADR, uniform builder, diff | `Last.F` format expected |
| `user.userId` | — | not heavily used client-side |
| `realName` | ADR, uniform builder name tag | full real name |
| `rank.rankId` | everywhere | maps to rank grade / epaulette image |
| `rank.rankShort` | `GetUserInfo`, server `RANK_SHORT` table | DDI short names need to stay in sync |
| `rank.rankFull` | ADR display | full rank title |
| `rank.rankImageUrl` | — | not used by the canvas; server uses its own `rankImageUrl()` |
| `primary.positionId` / `positionTitle` | ADR, statistics | used for billet matching |
| `secondaries` | ADR | secondary positions |
| `records` | roster history diffs | `recordType` string, `recordDate`, `recordDetails` |
| `awards` | uniform builder | `awardName`, `awardDetails`, `awardDate`, `awardImageUrl` |
| `roster` | diff poller, roster history | `ROSTER_TYPE_*` string |
| `mos` | uniform builder | 7CAV MOS codes for shoulder cords / neck pins |
| `uniformUrl` | — | present but not consumed by CavApps canvas |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/adr/modules/MilpacParse.jsx" /> <ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/GetUserInfo.jsx" />

### 6.2 Position groups (`/api/v1/milpacs/position/groups`)

The ADR consumes a `groups` array of position group objects, each with `positionId`, `positionTitle`, and `positionDisplayOrder` <ref_snippet file="/home/bobbyswagg/research/CavApps/README.md" lines="159-168" />.

### 6.3 Roster history / diff shape

`diffPoller.js` and the client roster-history code consume these `roster_type` string values and event types:

* `ROSTER_TYPE_COMBAT`, `ROSTER_TYPE_RESERVE`, `ROSTER_TYPE_ELOA`, `ROSTER_TYPE_WALL_OF_HONOR`, `ROSTER_TYPE_ARLINGTON`, `ROSTER_TYPE_PAST_MEMBERS`
* `PROMOTION`, `NEW_MEMBER`, `RETURN_TO_ACTIVE`, `WALL_OF_HONOR_INDUCTION`, `ROSTER_TRANSFER`, `TRANSFER_RESERVE`, `TRANSFER_ELOA`, `DISCHARGE`, `FALLEN`, `REMOVED`, `NEW_AWARD`, `POSITION_CHANGE`, `NAME_CHANGE`, `NEW_RECORD`

These are hardcoded in `client/app/rosterhistory/lib/constants.js`, `server/routes/diffRoutes.js`, and `server/controllers/diffPoller.js` <ref_file file="/home/bobbyswagg/research/CavApps/client/app/rosterhistory/lib/constants.js" /> <ref_file file="/home/bobbyswagg/research/CavApps/server/controllers/diffPoller.js" />.

Because the DDI API has already renamed `ROSTER_TYPE_COMBAT` → `ROSTER_TYPE_ACTIVE`, `WALL_OF_HONOR` → `HONOR_ROLL`, etc., these three files must be updated to the new wire names or the diff viewer will break.

### 6.4 Record types

`client/app/rosterhistory/lib/constants.js` `RECORD_TYPE_LABELS` maps wire record names to human labels. The DDI API has kept `RECORD_TYPE_PROMOTION` etc. by default, so this is stable unless DDI reconfigures `RECORD_TYPE_NAMES`.

## 7. Uniform Generation Workflow

1. `GetCanvasObject(userName)` calls `GetIndividual(userName)` (BFF → upstream `/milpacs/profile/username/{userName}`).
2. It iterates `data.awards` and looks up each `awardName` in `AwardRegistry` (driven by `client/app/uniformbuilder/modules/constants/awardCatalog.js`).
3. `GetUserInfo(data, ...)` builds the top-level metadata object:
   * `nameTag` from `user.username`
   * `rank` / `rankId` / `rankGrade` from `data.rank`
   * `mosCheck` and `shoulderCord` / `neckPins` from `data.mos` using `client/app/uniformbuilder/modules/constants/mos.js`
4. `Canvas` loads local PNG assets from `/skunkworks/uniformEpaulettes/{rankGrade}.png`, `/skunkworks/uniformRibbons/...`, etc.
5. Awards are sorted by `awardPriority` from `awardCatalog.js` and rendered using hardcoded sprite sheets.

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/getCanvasObject.jsx" /> <ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/GetUserInfo.jsx" />

### Uniform-specific constraints

* `rankId` must stay in the 1–31 range and map to the `getRankGrade()` switch in `GetUserInfo.jsx` or the epaulette images won't load.
* `awardName` must match entries in `awardCatalog.js` or the award won't render (the builder already skips unknown awards silently).
* `mos` values are 7CAV MOS codes (`11B`, `68W`, etc.). DDI can either keep the same codes or update `mos.js` and `setShoulderCord`/`setNeckPins`.

## 8. Remaining 7CAV-Specific References

### 8.1 Hardcoded upstream URLs

* `https://api.7cav.us` in `server/controllers/cacheManager.js` (4 calls) <ref_file file="/home/bobbyswagg/research/CavApps/server/controllers/cacheManager.js" />
* `https://7cav.us/data/roster_ranks/...` in `server/db/milpacDb.js` <ref_snippet file="/home/bobbyswagg/research/CavApps/server/db/milpacDb.js" lines="37-42" />
* `https://7cav.us/rosters/profile/` in `client/app/adr/modules/ArrayMap.jsx` (profile links)

### 8.2 Roster and record wire-name constants

* `client/app/rosterhistory/lib/constants.js` — `ROSTER_TYPE_*` and `RECORD_TYPE_*` labels <ref_file file="/home/bobbyswagg/research/CavApps/client/app/rosterhistory/lib/constants.js" />
* `server/routes/diffRoutes.js` — `VALID_ROSTER_TYPES` still lists the old 7CAV names <ref_snippet file="/home/bobbyswagg/research/CavApps/server/routes/diffRoutes.js" lines="16-23" />
* `server/controllers/diffPoller.js` — `ROSTER_TYPES` and `ROSTER_ID_MAP` use old names <ref_snippet file="/home/bobbyswagg/research/CavApps/server/controllers/diffPoller.js" lines="15-31" />

### 8.3 Rank short-name tables

* `server/db/milpacDb.js` `RANK_SHORT` maps `rank_id` → short abbreviation. It must match whatever `RANK_SHORT_NAMES` the DDI API emits.
* `client/app/uniformbuilder/modules/GetUserInfo.jsx` `getRankGrade()` maps `rank_id` → `O11`…`E1` grades.

### 8.4 MOS, awards, and branding

* `client/app/uniformbuilder/modules/constants/mos.js` — 7CAV MOS codes and wiki link
* `client/app/uniformbuilder/modules/constants/awardCatalog.js` — 7CAV award names (e.g. `7th Cavalry Lifetime Dedication Award`, `Cavalry Centurion Medal`)
* `client/app/theme/cavColor.js` — 7CAV logo SVG
* `client/app/page.jsx` — links to `https://7cav.us/forums/`
* `client/app/not-found.jsx` — links to `https://7cav.us/tickets/categories/2/create`
* `client/app/uniformbuilder/page.jsx` — placeholder text "Please enter a 7Cav Username"
* `README.md`, `.env.example`, workflow files — 7CAV URLs and tokens

## 9. Recommended Minimum Changes for DDI

To make DDIApps consume the DDI API with the smallest surface area:

1. **Upstream host**
   * Add `API_BASE_URL` env to `server/controllers/cacheManager.js` and replace `https://api.7cav.us`.
2. **CORS allowlist**
   * Update `server/server.js` with DDI app domains.
3. **API token**
   * Generate a DDI API key and set `API_TOKEN` (same bearer format).
4. **Roster wire names**
   * Update `server/routes/diffRoutes.js`, `server/controllers/diffPoller.js`, and `client/app/rosterhistory/lib/constants.js` to the new DDI roster names (`ROSTER_TYPE_ACTIVE`, `ROSTER_TYPE_RESERVE`, `ROSTER_TYPE_EXTENDED_LEAVE`, `ROSTER_TYPE_HONOR_ROLL`, `ROSTER_TYPE_MEMORIAL`, `ROSTER_TYPE_FORMER_MEMBERS`).
5. **Rank short names**
   * Update `server/db/milpacDb.js` `RANK_SHORT` to match the DDI API's `RANK_SHORT_NAMES` env, or keep the 7CAV defaults if DDI has not overridden them.
6. **MOS and awards (only if DDI data differs)**
   * Update `client/app/uniformbuilder/modules/constants/mos.js` and `awardCatalog.js` to DDI values.
7. **Branding URLs**
   * Replace `https://7cav.us/...` links in `client/app/page.jsx`, `client/app/not-found.jsx`, `client/app/adr/modules/ArrayMap.jsx`, and `client/app/uniformbuilder/page.jsx`.
8. **Asset images**
   * The canvas uses local `/skunkworks/...` PNGs. These are not API-driven; replace them with DDI-branded assets separately.

If DDI does not use the roster-history diff feature, steps 4 and the Postgres/XenForo setup can be skipped entirely.