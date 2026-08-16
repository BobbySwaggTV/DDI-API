# DDI Rank Asset Migration Plan

> Phase 5.5 planning: a focused migration plan for rank image and epaulette assets.  
> No files were moved, no rank IDs were renamed, no API contracts were changed, and no uniform-rendering code was modified.

## 1. Current Rank Asset Locations

### 1.1 API-generated rank image URLs

The DDI API builds the `rankImageUrl` field from `ASSET_BASE_URL` and a path template. The rank assets live on the configured asset host, not inside the API binary.

| Layer | Source | Output |
| --- | --- | --- |
| Go template | `milpacs/ranks.go` `ImageURL()` | `{ASSET_BASE_URL}/data/roster_ranks/{group}/{rankId}.jpg?{rankImage}` |
| Group calculation | `math.Floor(float64(rank.RankId / 1000))` | `0` for current rank_id values |
| Cache buster | `rank.RankImage` | appended as `?{rankImage}` query string |
| Wire field | `types/profiles.go` `Rank.RankImageUrl` | consumed by BFF diff and roster-history views |

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />

### 1.2 DDIApps BFF rank image helper

`server/db/milpacDb.js` mirrors the Go `ImageURL()` method and returns the same URL shape. It is used by the diff engine and `rosterhistory` timeline cards.

```javascript
`${ASSET_BASE_URL}/data/roster_ranks/${group}/${rankId}.jpg?${rankImage}`
```

<ref_file file="/home/bobbyswagg/research/CavApps/server/db/milpacDb.js" />

### 1.3 Frontend rank short → grade mapping

`client/app/uniformbuilder/modules/GetUserInfo.jsx` converts the API `rankId` string into a `rankGrade` used for the canvas and epaulette filename.

| `rankId` | `rankGrade` | DDI API `rankShort` |
| --- | --- | --- |
| 1 | O11 | GOA |
| 2 | O10 | GEN |
| 3 | O9 | LTG |
| 4 | O8 | MG |
| 5 | O7 | BG |
| 6 | O6 | COL |
| 7 | O5 | LTC |
| 8 | O4 | MAJ |
| 9 | O3 | CPT |
| 10 | O2 | 1LT |
| 11 | O1 | 2LT |
| 12 | E11 | CSM |
| 13 | E10 | SGM |
| 14 | E9 | 1SG |
| 15 | E8 | MSG |
| 16 | E7 | SFC |
| 17 | E6 | SSG |
| 18 | E5 | SGT |
| 19 | E4 | CPL |
| 20 | E3 | SPC |
| 21 | E2 | PFC |
| 22 | E1 | PVT |
| 23 | E0 | RCT |
| 26 | W5 | CW5 |
| 27 | W4 | CW4 |
| 28 | W3 | CW3 |
| 29 | W2 | CW2 |
| 30 | W1 | WO1 |
| 31 | E0 | AR |

`rankGrade` is then used to load `client/public/skunkworks/uniformEpaulettes/{rankGrade}.png`.

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/GetUserInfo.jsx" />

### 1.4 Frontend epaulette image assets

The canvas loads local PNGs from `client/public/skunkworks/uniformEpaulettes/`.

| File | What it represents |
| --- | --- |
| `O11.png` – `O1.png` | officer epaulettes (GOA down to 2LT) |
| `E11.png` – `E1.png` | enlisted epaulettes (CSM down to PVT) |
| `W5.png` – `W1.png` | warrant-officer epaulettes |
| `E0.png` | fallback / unknown rank grade |

### 1.5 Rank short name catalog

The API's `rankShort` wire values come from `internal/config/orgmap.go` `RANK_SHORT_NAMES`. The default table is the 7CAV set, but it can be overridden with the `RANK_SHORT_NAMES` env.

The DDIApps BFF `server/db/milpacDb.js` has a static `RANK_SHORT` table for sorting. It must stay in sync with the API, but it does **not** control the API's `rankShort` on the wire.

<ref_file file="/home/bobbyswagg/research/DDI-API/internal/config/orgmap.go" />

## 2. Rank Asset Requirements

### 2.1 rank_id dependencies

* The database `xf_nf_rosters_rank.rank_id` values are the authority.
* The API and frontend depend on `rankId` values 1–23 and 26–31.
* Gaps 24–25 are intentionally unassigned and must remain so.
* `GetUserInfo.jsx` `getRankGrade()` falls back to `E0` for unknown `rankId`.
* Any new DDI rank must either reuse an existing `rankId` or update the mapping in `GetUserInfo.jsx` and the epaulette PNG names.

### 2.2 Folder structure

**API host** (same as current): `/data/roster_ranks/{group}/{rankId}.jpg`

```
{ASSET_BASE_URL}/data/roster_ranks/0/1.jpg?{rankImage}
{ASSET_BASE_URL}/data/roster_ranks/0/22.jpg?{rankImage}
...
```

`group` is `floor(rankId / 1000)`. For current `rankId` values this is always `0`. Keeping the `group` subfolder lets DDI add new rank groups later (e.g. `1000–1999`) without code changes.

**Frontend bundle**: `client/public/skunkworks/uniformEpaulettes/`

```
client/public/skunkworks/uniformEpaulettes/
  O11.png ... O1.png
  E11.png ... E1.png
  W5.png ... W1.png
  E0.png
```

### 2.3 Naming conventions

| Location | Naming rule | Example |
| --- | --- | --- |
| API rank images | `{group}/{rankId}.jpg?{cacheBuster}` | `0/5.jpg?1741364618` |
| Frontend epaulettes | `{rankGrade}.png` | `O7.png` |
| Cache buster value | `rank.RankImage` integer from `xf_nf_rosters_rank` | `?5` |

Cache busters let DDI update rank art without renaming files or breaking CDN/browser caches.

### 2.4 Cache behavior

* The `rankImage` column in `xf_nf_rosters_rank` is the cache buster.
* When the art changes, increment (or reset) `rankImage` for the affected rows.
* Browsers and CDNs see a new `?{rankImage}` query string and request a fresh copy.
* The API itself does not cache the image bytes; it only emits URLs.

## 3. Recommended DDI Asset Layout

### 3.1 API-served rank images

Keep the existing `/data/roster_ranks/...` path shape. The migration is a bulk file copy plus a `rankImage` bump.

```
assets.ddi.example/data/roster_ranks/
  0/
    1.jpg     GOA
    2.jpg     GEN
    ...
    22.jpg    PVT
    23.jpg    RCT
    26.jpg    CW5
    27.jpg    CW4
    ...
    31.jpg    AR
```

### 3.2 Frontend epaulette images

Two options, in order of risk:

1. **Replace in place (lowest risk)**
   - Keep `client/public/skunkworks/uniformEpaulettes/`.
   - Replace the 30 PNGs with DDI epaulette art.
   - No code changes.

2. **Rename the folder (higher risk)**
   - Create `client/public/ddi/uniformEpaulettes/`.
   - Update `canvas.jsx` to load from the new path.
   - Requires code change and careful testing.

### 3.3 If DDI rank grades change

If DDI does not use the same O/E/W grade scheme, the new scheme must be mapped to the existing filenames or the following must be updated together:

* `client/app/uniformbuilder/modules/GetUserInfo.jsx` `getRankGrade()`
* `client/public/skunkworks/uniformEpaulettes/` PNG filenames
* The API `RANK_SHORT_NAMES` catalog

This is the highest-risk variant and is out of scope for a simple asset migration.

## 4. Migration Process

### 4.1 Development testing

1. **Set `ASSET_BASE_URL` to a staging bucket**
   - E.g. `https://assets-staging.ddi.example`.
   - Copy the 7CAV rank images there first.
2. **Verify API URLs**
   - `GET /api/v1/milpacs/profile/username/{user}` and `GET /api/v1/roster/1` should return `rankImageUrl` using the staging host.
3. **Verify BFF diff views**
   - `server/db/milpacDb.js` `rankImageUrl()` should produce the same staging URL.
   - `rosterhistory` timeline cards should render rank images.
4. **Replace epaulettes locally**
   - Put DDI PNGs in `client/public/skunkworks/uniformEpaulettes/` (or a new tree).
   - Run the uniform builder and confirm every rank grade loads.

### 4.2 Production switch

1. **Upload DDI rank images to the production asset host**
   - Path: `{ASSET_BASE_URL}/data/roster_ranks/0/{rankId}.jpg?{rankImage}`
   - Use the same `rankImage` values for zero-downtime cutover, or bump them for cache invalidation.
2. **Set `ASSET_BASE_URL` to the DDI host**
   - DDI API `.env` / `docker-compose.yaml`
   - DDIApps BFF `.env` / `docker-compose.yaml`
3. **Deploy the DDIApps epaulettes**
   - Replace PNGs in `client/public/skunkworks/uniformEpaulettes/` (or the new tree) and rebuild.
4. **Smoke test**
   - ADR, roster history, and uniform builder for at least one profile per rank grade.

### 4.3 Rollback plan

1. **API/BFF rollback**
   - Revert `ASSET_BASE_URL` to `https://7cav.us` (or the previous DDI host).
   - The path templates are identical, so this is an immediate URL change.
2. **Frontend rollback**
   - Restore the old `uniformEpaulettes/` PNGs from version control.
3. **Cache recovery**
   - Browsers may have cached DDI images. If rolling back, bump `rankImage` or change `ASSET_BASE_URL` to a different path to force a fresh fetch.

## 5. Risks and Mitigations

| Risk | When it happens | Mitigation |
| --- | --- | --- |
| **Broken image paths** | DDI rank images are not at `{group}/{rankId}.jpg` under `ASSET_BASE_URL`. | Mirror the exact path template. Use a staging host and verify the API's `rankImageUrl` field before production. |
| **rank_id mismatch** | DDI uses different `rank_id` values or renumbers ranks. | Keep the legacy `rank_id` values until `GetUserInfo.jsx` `getRankGrade()` and the epaulette PNG names are remapped. |
| **Missing assets** | DDI does not have rank art for every `rank_id` used in the database. | Add a default `E0.png` fallback and ensure `GetUserInfo.jsx` returns `E0` for unknown IDs. |
| **Cache not invalidated** | New rank images show old 7CAV art because of CDN/browser caching. | Bump `rank.RankImage` values when replacing files, or add a query-string version to `ASSET_BASE_URL` if DDI's CDN does not respect `rankImage`. |
| **Mismatched rank short names** | The BFF `RANK_SHORT` table and the API `RANK_SHORT_NAMES` diverge. | Keep `RANK_SHORT` in `milpacDb.js` in sync with `RANK_SHORT_NAMES` or the env override. The diff engine uses the BFF table. |
| **Uniform builder shows wrong epaulette** | `rankGrade` is `E0` because `rankId` 24 or 25 is in use. | 24 and 25 are intentionally unmapped. Add explicit grades if those IDs become active, or keep them unused. |
| **Frontend path mismatch** | `canvas.jsx` loads `O7.png` but the new DDI file is `O7-new.png`. | Replace PNGs in place (same filenames) to avoid code changes. |

## 6. Summary

* The rank asset migration is a **path-safe, id-safe** change: keep `rankId` values and the `/data/roster_ranks/0/{rankId}.jpg` template.
* The lowest-risk approach is to replace the actual image files (API host + epaulette PNGs) while keeping all filenames and URL templates identical.
* Cache busters in `xf_nf_rosters_rank.rankImage` make the swap atomic and cache-friendly.
* The risk is not in the code; it is in the content (does DDI have correct art for every current `rank_id`?).