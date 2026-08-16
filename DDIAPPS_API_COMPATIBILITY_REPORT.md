# DDIApps API Compatibility Report

> Phase 4.2 validation: confirm the DDIApps BFF can consume the DDI API.  
> No application code was changed.

## 1. Validation Environment

A live, end-to-end check was performed using the DDI API's own MariaDB test harness:

* DDI API source: `/home/bobbyswagg/research/DDI-API`
* DDIApps BFF source: `/home/bobbyswagg/research/CavApps`
* Database: MariaDB 11.5 from `testdb/compose.yaml`, seeded with `schema.sql` and `fixtures.sql`
* DDI API launched on `127.0.0.1:11000` with `DB_USERNAME=root DB_PASSWORD=harness DB_HOST=127.0.0.1 DB_PORT=<ephemeral>`
* BFF `cacheManager.js` exercised directly against `API_BASE_URL=http://127.0.0.1:11000` and `API_TOKEN=ddi_harness_active`
* Direct `curl` samples captured to `/tmp/ddi-validation/` for manual inspection

## 2. Required Environment Configuration

For the DDIApps BFF to reach the DDI API, the server needs these environment variables:

| Variable | Purpose | Example / Notes |
| --- | --- | --- |
| `API_BASE_URL` | DDI API host (no trailing slash) | `https://api.ddi.example` |
| `API_TOKEN` | Bearer token for upstream `Authorization` header | issued by the DDI API operator |
| `CLIENT_TOKEN` | Shared secret between Next.js client and BFF | any matching string on both sides |
| `ASSET_BASE_URL` | Base URL for rank image rendering in the diff poller | default `https://7cav.us` if unset |
| `CORS_ORIGINS` | Comma-separated extra allowed browser origins | `https://apps.ddi.example,https://appsbeta.ddi.example` |
| `FRONTEND_URL` | Used in the BFF health-check response and CORS | default `http://localhost:3000` |

<ref_file file="/home/bobbyswagg/research/CavApps/.env.example" />

## 3. BFF → DDI API Route Verification

| BFF upstream call | DDI API endpoint | Result | Notes |
| --- | --- | --- | --- |
| combat roster | `GET /api/v1/roster/1` | OK | 4 profiles returned, `ROSTER_TYPE_ACTIVE` |
| reserve roster | `GET /api/v1/roster/2` | OK | 2 profiles returned, `ROSTER_TYPE_RESERVE` |
| individual profile | `GET /api/v1/milpacs/profile/username/Trooper.A` | OK | full profile returned |
| position groups | `GET /api/v1/milpacs/position/groups` | OK | `{ "groups": [...] }` object returned |

### 3.1 `cacheManager` live check

Using the BFF's `cacheManager` against the running DDI API:

```text
combat keys: 4
reserve keys: 2
groups count: undefined (object is { "groups": [...] }, not a plain array)
individual username: Trooper.A
```

`undefined` is expected — the groups endpoint returns a wrapper object with a `groups` array, and the BFF's `getCachedGroups()` returns the whole wrapper. The client already consumes it as `rosterGroups.groups[selector]`.

The direct `curl` and the BFF `axios` calls both used `Authorization: Bearer ddi_harness_active` and completed successfully, confirming the bearer-token path and `API_BASE_URL` wiring are correct.

## 4. DDI API Response Shape vs. BFF Expectations

### 4.1 Roster response (`/api/v1/roster/1`)

Shape is `{ profiles: { [relationId]: Profile } }`.

The `Profile` object contains every field the BFF and client consume:

| Field | Present? | DDI value | BFF usage | Status |
| --- | --- | --- | --- | --- |
| `user.userId` | yes | `"100"` | unused | OK |
| `user.username` | yes | `"Trooper.A"` | ADR, uniform name tag | OK |
| `realName` | yes | `"Alpha Smith"` | ADR display | OK |
| `rank.rankId` | yes | `"5"` | `getRankGrade()` → epaulette image | OK |
| `rank.rankShort` | yes | `"BG"` | `GetUserInfo` rank display | OK if catalog matches |
| `rank.rankFull` | yes | `"Sergeant Major"` | ADR display | OK |
| `rank.rankImageUrl` | yes | `https://7cav.us/...` | server `rankImageUrl()` | OK |
| `primary.positionId` | yes | `"11"` | ADR billet matching | OK |
| `primary.positionTitle` | yes | `"Squad Leader"` | ADR display | OK |
| `secondaries` | yes | array of positions | ADR billet matching | OK |
| `records` | yes | array | diff engine, roster history | OK |
| `awards` | yes | array | uniform builder | OK |
| `mos` | yes | `"11B"` | uniform MOS logic | OK |
| `roster` | yes | `"ROSTER_TYPE_ACTIVE"` | diff engine, roster history | **Needs update for diff feature** |
| `uniformUrl` | yes | `https://7cav.us/...` | not consumed by canvas | OK |

### 4.2 Position groups (`/api/v1/milpacs/position/groups`)

```json
{
  "groups": [
    {
      "groupId": "1",
      "groupTitle": "Regimental HQ",
      "groupDisplayOrder": 10,
      "positions": [
        {
          "positionTitle": "Military Police",
          "positionId": "20",
          "positionDisplayOrder": 30,
          "positionPossibleSecondary": true
        }
      ]
    }
  ]
}
```

The client ADR expects `rosterGroups.groups[selector].positions` and `position.positionId`, both present. **No incompatibility.**

### 4.3 Individual profile (`/api/v1/milpacs/profile/username/{username}`)

Same `Profile` shape as the roster endpoint. The BFF's `GetIndividual` and the uniform builder's `GetCanvasObject` / `GetUserInfo` receive the same fields as the roster object.

### 4.4 Records

```json
{
  "recordDetails": "Promoted to Sergeant Major",
  "recordType": "RECORD_TYPE_PROMOTION",
  "recordDate": "2023-05-02",
  "recordUid": "1001"
}
```

The DDI API is currently configured with the default `RECORD_TYPE_NAMES` (`RECORD_TYPE_PROMOTION`, etc.), which matches the `client/app/rosterhistory/lib/constants.js` `RECORD_TYPE_LABELS`. **No incompatibility unless `RECORD_TYPE_NAMES` is overridden.**

### 4.5 Awards

```json
{
  "awardDetails": "For exemplary conduct",
  "awardName": "Good Conduct Medal",
  "awardDate": "2023-05-25",
  "awardImageUrl": "https://7cav.us/data/roster_awards/0/1.jpg?1",
  "awardUid": "2001"
}
```

The shape (`awardName`, `awardDetails`, `awardDate`, `awardImageUrl`) is exactly what `getCanvasObject.jsx` consumes. **The uniform builder will not break.**

However, the `AwardRegistry` in `client/app/uniformbuilder/modules/constants/awardCatalog.js` is keyed on `awardName`. The test fixtures use simple names (`Good Conduct Medal`, `Purple Heart`) that are not in the 7CAV catalog, so those awards are skipped by the canvas. In a real DDI deployment the catalog must be updated to DDI award titles, but this is a **data/content** task, not an API contract issue.

## 5. Identified Mismatches

### 5.1 `roster` wire-name value (`ROSTER_TYPE_ACTIVE` vs. `ROSTER_TYPE_COMBAT`)

* DDI API returns `ROSTER_TYPE_ACTIVE` for roster id `1`.
* The DDIApps BFF and client still use `ROSTER_TYPE_COMBAT` in `diffRoutes.js`, `diffPoller.js`, and `client/app/rosterhistory/lib/constants.js`.
* **Impact:** If the roster-history diff feature is used, the diff engine will not recognize the new roster type names and will treat profiles as transfers / unknown events.
* **Risk level:** High for the diff feature; zero for ADR, statistics, and uniform builder.
* **Recommendation:** update the diff constants in a follow-up phase, or configure the DDI API to emit the old 7CAV roster names temporarily.

### 5.2 Asset base URLs still point to `https://7cav.us`

* DDI API and BFF `rankImageUrl()` default to `https://7cav.us` when `ASSET_BASE_URL` is not set.
* The canvas uses local PNG assets, not these URLs, but the server-side diff poller and OpenAPI docs still reference 7CAV assets.
* **Impact:** Low for the uniform builder; medium for brand-correctness in the diff viewer / docs.
* **Recommendation:** set `ASSET_BASE_URL` to the DDI asset host when DDI rank/award images are available.

### 5.3 Award catalog (`awardCatalog.js`) still lists 7CAV award names

* The API contract (`awardName`) is correct, but the client-side catalog keys do not match DDI award titles.
* **Impact:** Uniform builder will not render DDI-specific awards until the catalog is updated.
* **Risk level:** Medium for the uniform builder; low for ADR and statistics.
* **Recommendation:** update `client/app/uniformbuilder/modules/constants/awardCatalog.js` and the `/skunkworks/` sprite sheets to DDI awards.

### 5.4 Rank short names / `rankFull` may be out of sync

* The DDI API's `rankShort` values come from `RANK_SHORT_NAMES` env. The BFF's `server/db/milpacDb.js` `RANK_SHORT` table is also used for ordering.
* **Impact:** If DDI overrides rank shorts, the BFF diff `/ranks` endpoint and the rank ordering in the diff engine will diverge.
* **Risk level:** Low if DDI keeps 7CAV defaults; medium if it overrides.
* **Recommendation:** keep `RANK_SHORT_NAMES` in sync with `server/db/milpacDb.js`.

## 6. Authentication Verification

* DDI API requires `Authorization: Bearer <key>`.
* BFF sends `Authorization: Bearer ${API_TOKEN}` to the DDI API.
* Client-to-BFF uses `Authorization: ${CLIENT_TOKEN}` (raw shared secret), unchanged.
* Live calls succeeded with the fixture token `ddi_harness_active`.
* **No authentication incompatibility.**

## 7. Uniform Builder Compatibility

The uniform builder relies on:

* `data.user.username` → name tag
* `data.rank.rankId` → `getRankGrade()` → `skunkworks/uniformEpaulettes/{rankGrade}.png`
* `data.mos` → `mos.js` shoulder cord / neck pins
* `data.awards[i].awardName` → `AwardRegistry` → local sprite sheets
* `data.awards[i].awardDetails` → tier/valor logic

All of those fields are present in the DDI API response. The **only** uniform-builder risk is the `awardName` catalog (Section 5.3). The `rankId` → `rankGrade` mapping uses the existing 7CAV `rank_id` values, which the DDI API preserves, so the epaulette PNGs will load.

## 8. Summary

| Area | Status | Notes |
| --- | --- | --- |
| BFF → DDI API connectivity | PASS | `cacheManager` successfully fetched all four endpoints |
| Bearer auth | PASS | same `Authorization: Bearer` scheme |
| Roster shape (`/roster/1`, `/roster/2`) | PASS | `profiles` map with all required fields |
| Individual profile shape | PASS | same `Profile` object as roster |
| Position groups shape | PASS | `{ "groups": [...] }` matches client expectations |
| Records shape | PASS | default record wire names unchanged |
| Awards shape | PASS | all consumed fields present |
| Uniform builder data flow | PASS | all required fields present |
| Roster history / diff | **NEEDS UPDATE** | `ROSTER_TYPE_ACTIVE` vs. `ROSTER_TYPE_COMBAT` mismatch |
| Award rendering | **NEEDS CONTENT UPDATE** | catalog is still 7CAV, not API-contract-breaking |

## 9. Recommended Next Steps

1. **Minimum for ADR + Uniform Builder to work:** set `API_BASE_URL`, `API_TOKEN`, `CLIENT_TOKEN` in `.env` and deploy. No further code changes needed.
2. **For roster-history diff:** update `server/routes/diffRoutes.js`, `server/controllers/diffPoller.js`, and `client/app/rosterhistory/lib/constants.js` to DDI roster wire names.
3. **For DDI branding:** set `ASSET_BASE_URL` and update `client/app/uniformbuilder/modules/constants/awardCatalog.js` / sprite sheets.
4. **Before production:** re-run this validation against the live DDI database with real DDI `RANK_SHORT_NAMES`, `RECORD_TYPE_NAMES`, and `ORG_ROSTER_NAMES` to catch any catalog-level mismatches.