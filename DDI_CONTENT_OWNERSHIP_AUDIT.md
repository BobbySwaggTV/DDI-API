# DDI Content Ownership Audit

> Phase 5: identify all remaining 7CAV-specific content dependencies and decide what must become DDI-owned.  
> No code, database schema, API contracts, or assets were changed.

## 1. Scope

This audit covers content that still carries 7th Cavalry identity, images, or data in two repositories:

* **DDIApps/CavApps** (`/home/bobbyswagg/research/CavApps`) — Next.js frontend and Express BFF
* **DDI API** (`/home/bobbyswagg/research/DDI-API`) — Go backend

"Content" means: award catalogs, medal/ribbon definitions, rank assets, MOS constants, uniform-builder data and images, branding assets, and hardcoded 7CAV URLs.

## 2. CavApps / DDIApps Content

### 2.1 Award catalog

| What | Source file | Current ownership | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- | --- |
| Award names, placement, priority, type | `client/app/uniformbuilder/modules/constants/awardCatalog.js` | 7CAV (hundreds of named awards) | No | Yes | High — the uniform builder key-matches `awardName` from the API to this catalog. If DDI awards have different titles, the builder silently renders no medals/ribbons. |
| Award class taxonomy | `client/app/uniformbuilder/modules/constants/awardTypes.js` | Generic (Medal, Ribbon, etc.) | Yes | No | Low — the values are abstract classes, not 7CAV-specific. |
| Attachment-device taxonomy | `client/app/uniformbuilder/modules/constants/awardAttachmentTypes.js` | Generic (oak leaf, star, etc.) | Yes | No | Low — the device types are US-military standard. The `MAX_AWARD_COUNT` values are display constants. |
| Award-name string fragments | `client/app/uniformbuilder/modules/constants/awardNames.js` | Generic ("with Valor Device") | Yes | No | Low — only a few shared string fragments. |
| Award registry | `client/app/uniformbuilder/modules/AwardRegistry.jsx` | Code, not content | Yes | No | None — it is a lookup layer. |
| Badge image number mapping | `client/app/uniformbuilder/modules/constants/badgeImages.js` | 7CAV-ish (CIB/EIB/Aviator sequence) | Partial | Partial | Medium — the `combatBadges` PNG files are local; the number mapping must match the new images if DDI replaces them. |
| Award classes (render logic) | `client/app/uniformbuilder/modules/AwardClasses.jsx` | Code, mostly generic | Yes | No | Low — some donation-award thresholds are hardcoded, but the values are display rules, not identity. |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/constants/awardCatalog.js" />

### 2.2 Medal, ribbon, and badge images

| What | Source | Count | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- | --- |
| Sprite sheets and loose PNGs | `client/public/skunkworks/` (uniformRibbons, uniformMedals, uniformBadges, uniformTabs, uniformWeaponQuals, uniformSpecialMedals, uniformService, etc.) | ~200 PNGs | No | Yes | High — these are 7CAV-specific award, rank, and uniform art. Rebranding requires new image assets. |
| Sprite-sheet generator | `generateAwardSprites.js` | 1 script | N/A | N/A | Low — it is a build-time tool keyed to `awardCatalog.js` and `uniform-uploads/manifest.json`. It can consume DDI award art once the catalog is updated. |
| Manifest for uploads | `uniform-uploads/manifest.json` | 1 file | No | Yes | Medium — the upload queue for new award art. Needs DDI entries. |

<ref_file file="/home/bobbyswagg/research/CavApps/generateAwardSprites.js" />

### 2.3 Rank assets

| What | Source | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- |
| Rank → grade mapping (O11…E1, W5…) | `client/app/uniformbuilder/modules/GetUserInfo.jsx` | Partial | Partial | Medium — the mapping is generic US pay-grade logic and works for any army-style rank set. If DDI renumbers `rank_id` values, the map and the epaulette PNG names must change. |
| Rank epaulette PNGs | `client/public/skunkworks/uniformEpaulettes/{rankGrade}.png` | No | Yes | Medium — the canvas loads `O6.png`, `E5.png`, etc. These can stay if DDI adopts the same rank grades; otherwise, the mapping and/or images change. |
| Server rank short table | `server/db/milpacDb.js` `RANK_SHORT` | No | Yes | Low/Medium — used by the diff engine. If DDI overrides `RANK_SHORT_NAMES`, this table must match or diff sorting breaks. |

### 2.4 MOS constants

| What | Source | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- |
| MOS code → shoulder cord / neck pin logic | `client/app/uniformbuilder/modules/constants/mos.js` | No | Yes | High — the file is explicitly 7CAV S1 MOS codes and a 7CAV wiki link. If the DDI API returns different `mos` strings, shoulder cords and neck pins will be wrong or missing. |
| MOS string in API | DDI API `mos` custom field | N/A | N/A | High — DDI can either keep the same `11B`, `68W` etc. values, or the client `mos.js` must be re-mapped. |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/constants/mos.js" />

### 2.5 Uniform builder data

| What | Source | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- |
| Name-tag generation | `client/app/uniformbuilder/modules/GetUserInfo.jsx` `generateNameTag` | Yes | No | None — username-prefix logic is generic. |
| Years-in-service, citation, tab coordinate math | `client/app/uniformbuilder/modules/get*CoordArray.jsx` | Yes | No | Low — layout math depends on tile geometry, not content. |
| Canvas image loading | `client/app/uniformbuilder/modules/canvas.jsx` | Partial | Partial | Medium — `skunkworks/` paths are hardcoded. The base paths stay; only the PNG contents need DDI art. |
| Uniform base / lapel / name-tag / cord / tab images | `client/public/skunkworks/uniformBase/`, `uniformLapelPins/`, `uniformNameTag/`, `uniformCords/`, `uniformTabs/` | No | Yes | High — these are the literal uniform artwork and DDI branding. |

### 2.6 Branding assets

| What | Source | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- |
| App logo SVG | `client/public/CavAppsLogo.svg` | No | Yes | High — 7CAV logo. |
| ADR logo component | `client/app/theme/adrLogo.js` | No | Yes | High — 7CAV text/logo. |
| CavApps text component | `client/app/theme/appsLogo.js` | No | Yes | High — 7CAV branding. |
| 7CAV color SVG | `client/app/theme/cavColor.js` | No | Yes | High — 7CAV shield used in uniform builder. |
| Skunkworks logo | `client/app/theme/skunkworksLogo.js` | No | Yes | Medium — 7CAV uniform builder logo. |
| Uniform builder title | `client/app/theme/uniformBuilderText.js` | No | Yes | Medium — 7CAV text. |
| Page links | `client/app/page.jsx` `https://7cav.us/forums/` | No | Yes | Low — one link; easy to replace. |
| Not-found support link | `client/app/not-found.jsx` `https://7cav.us/tickets/...` | No | Yes | Low — one link. |
| ADR profile links | `client/app/adr/modules/ArrayMap.jsx` `https://7cav.us/rosters/profile/` | No | Yes | Medium — every ADR name links to the 7CAV forum. |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/theme/cavColor.js" />
<ref_file file="/home/bobbyswagg/research/CavApps/client/app/adr/modules/ArrayMap.jsx" />

## 3. DDI API Content

### 3.1 Award data sources

| What | Source | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- |
| Award titles | `xf_nf_rosters_award.title` (MariaDB) | No | Yes (DB content) | High — the API returns `AwardName` from `awardRow.Award.Title`. If the database still has 7CAV award names, the uniform builder catalog must match them or nothing renders. |
| Award image URLs | `milpacs/awards.go` `ImageURL()` | Partial | Partial | Low — the URL shape is generic (`{ASSET_BASE_URL}/data/roster_awards/{group}/{awardId}.jpg?{image}`). The default base is `https://7cav.us`, overrideable via `ASSET_BASE_URL`. The actual image files must be DDI-owned. |
| Award JSON shape | `types/profiles.go` `Award` struct | Yes | No | None — contract is neutral. |

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />

### 3.2 Rank assets

| What | Source | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- |
| Rank short names | `internal/config/orgmap.go` `RANK_SHORT_NAMES` default | No | Yes | Medium — default 7CAV short names can be overridden via `RANK_SHORT_NAMES` env. If the client `GetUserInfo.jsx` uses the same pay grades, the short strings themselves are only display labels. |
| Rank image URLs | `milpacs/ranks.go` `ImageURL()` | Partial | Partial | Low — generic URL shape; default base is `https://7cav.us`. DDI can override with `ASSET_BASE_URL`. The actual image files matter more. |
| Rank wire names | `types/ranks.go` `rankTypeNames` | Yes | No | Low — the `RANK_TYPE_*` constants are internal; the API exposes `rankShort` and `rankFull` from the DB/config. |

<ref_file file="/home/bobbyswagg/research/DDI-API/internal/config/orgmap.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" />

### 3.3 Profile assets

| What | Source | Can remain inherited? | Must become DDI? | Risk |
| --- | --- | --- | --- | --- |
| Uniform image URLs | `milpacs/profiles.go` `UniformUrl()` | Partial | Partial | Low — generic URL shape; default base `https://7cav.us`; overridable via `ASSET_BASE_URL`. |
| Asset base URL default | `milpacs/asset_url.go` `defaultAssetBaseURL = "https://7cav.us"` | No | Yes | Medium — the fallback still points to 7CAV. For a DDI deployment `ASSET_BASE_URL` must be set. |
| Forum ticket base URL | `datastores/tickets.go` `forumBaseURL()` | N/A | N/A | Low — uses `FORUM_BASE_URL` env; empty by default. If not set, `ForumUrl` is empty. |

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/asset_url.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" />

### 3.4 Remaining 7CAV URLs in DDI API

The 7CAV URLs in the DDI API fall into three buckets:

1. **Default asset host** (`milpacs/asset_url.go` `defaultAssetBaseURL`)
   - Always override with `ASSET_BASE_URL` in production.
2. **Hardcoded 7CAV URLs in tests**
   - `milpacs/asset_url_test.go`, `rest/*_test.go`, `datastores/*_test.go`, `contract/*_test.go`
   - These are test fixtures; they are acceptable to keep because they pin legacy behavior for unconfigured deployments.
3. **Test fixture `FORUM_BASE_URL` defaults**
   - `FORUM_BASE_URL` is empty by default; tests use `https://7cav.us` or `https://forum.example.test` explicitly.

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/asset_url.go" />

## 4. What Can Remain Inherited Temporarily

| Content | Why it can stay for now | Caveat |
| --- | --- | --- |
| Award type taxonomy (`awardTypes.js`) | Abstract class names, no 7CAV identity | None. |
| Attachment-device taxonomy (`awardAttachmentTypes.js`) | Generic device names and counts | DDI may want different max counts. |
| `MAX_AWARD_COUNT` | Display constants | Fine until award SOP changes. |
| `GetUserInfo.jsx` pay-grade mapping | Generic US rank grades (O11..E1, W5) | Breaks if DDI changes `rank_id` values. |
| `AwardClasses.jsx` render classes | Code, not content | Donation thresholds are hardcoded but not identity. |
| Coordinate-array math in uniform builder | Layout calculations | Tile geometry must match sprite sheets. |
| DDI API `Profile`/`Award`/`Rank` JSON shapes | Neutral contracts | None. |
| DDI API default rank wire-name catalog (`rankTypeNames`) | Internal constants | Not visible on the wire. |

## 5. What Must Become DDI-Owned

| Content | Why it must change | Recommended path |
| --- | --- | --- |
| `client/public/skunkworks/` images and SVGs | They are 7CAV-specific art | Commission / generate DDI art and replace files. |
| `client/app/uniformbuilder/modules/constants/awardCatalog.js` | Award names are 7CAV-specific | Build a DDI award catalog with the same schema (`name`, `awardPriority`, `medalPriority`, `awardType`, `awardAttachmentType`). |
| `client/app/uniformbuilder/modules/constants/mos.js` | 7CAV S1 MOS codes and wiki link | Map DDI MOS strings to the same shoulder-cord / neck-pin categories, or keep identical codes if DDI uses the same set. |
| `client/public/CavAppsLogo.svg` and `client/app/theme/*.js` | 7CAV branding | Replace SVG paths/logos with DDI versions. |
| `client/app/adr/modules/ArrayMap.jsx` | Links to `https://7cav.us/rosters/profile/` | Replace with DDI profile URL or remove the link. |
| `client/app/page.jsx` and `client/app/not-found.jsx` | 7CAV forum/ticket links | Replace with DDI URLs. |
| `client/app/uniformbuilder/page.jsx` placeholder | "Please enter a 7Cav Username" | Update text. |
| `README.md` and `.env.example` (CavApps) | 7CAV references | Already partially updated; finish branding. |
| DDI API database award titles (`xf_nf_rosters_award.title`) | Titles feed `AwardName` on the wire | Update DB content or ensure titles match the DDI catalog. |
| DDI API `ASSET_BASE_URL` default (`milpacs/asset_url.go`) | Default is `https://7cav.us` | Set `ASSET_BASE_URL` in production; optionally change the default once DDI assets are ready. |

## 6. Recommended Migration Order

1. **Phase 5.1 — text and links (low risk)**
   - Update placeholder text, `README.md`, `.env.example`, `page.jsx`, `not-found.jsx`, and `ArrayMap.jsx`.
   - Replace `client/app/theme/*.js` and `client/public/CavAppsLogo.svg` with DDI logos.

2. **Phase 5.2 — asset base URL (low risk)**
   - Set `ASSET_BASE_URL` to the DDI asset host in both DDI API and DDIApps server.
   - Keep the same URL path templates (`/data/roster_ranks/...`, `/data/roster_awards/...`, `/data/roster_uniforms/...`) so no code changes are needed.

3. **Phase 5.3 — rank art and rank ID stability (medium risk)**
   - Confirm DDI `rank_id` values 1–31 still map to the same pay grades.
   - Replace `client/public/skunkworks/uniformEpaulettes/` PNGs and rank image assets on the DDI asset host.

4. **Phase 5.4 — MOS mapping (medium risk)**
   - Decide if DDI will keep the same `11B`, `68W`, etc. codes.
   - If not, update `client/app/uniformbuilder/modules/constants/mos.js` and `setShoulderCord` / `setNeckPins`.

5. **Phase 5.5 — award catalog and sprite sheets (high risk)**
   - Build the DDI `awardCatalog.js` with DDI award names and priorities.
   - Add/replace award PNGs in `uniform-uploads/` and regenerate sprite sheets with `npm run sprites:generate`.
   - Ensure the DDI API database award titles match the catalog exactly, or the `AwardRegistry` will skip them.

6. **Phase 5.6 — remaining 7CAV test fixtures (low risk)**
   - Update DDI API test goldens and `defaultAssetBaseURL` once DDI assets are authoritative and tests are rewritten to the new defaults.

## 7. Compatibility Risks

| Risk | Area | Mitigation |
| --- | --- | --- |
| Uniform builder silently skips DDI awards because `awardName` does not match `awardCatalog.js` | `client/app/uniformbuilder/modules/constants/awardCatalog.js` | Ensure the DDI award catalog is the single source of truth and that the DB `xf_nf_rosters_award.title` values match the catalog keys. |
| MOS strings change and shoulder cords / neck pins disappear | `client/app/uniformbuilder/modules/constants/mos.js` | Either keep the same MOS codes in DDI or re-map the client MOS table. |
| `rank_id` values renumber and epaulette images break | `client/app/uniformbuilder/modules/GetUserInfo.jsx` / `client/public/skunkworks/uniformEpaulettes/` | Keep `rank_id` values stable and replace only the PNG art. If renumbering, update `getRankGrade()` and `rankGrade` PNG names. |
| `RANK_SHORT_NAMES` env overrides and BFF diff table gets out of sync | `server/db/milpacDb.js` `RANK_SHORT` | Keep `RANK_SHORT` in `milpacDb.js` in sync with the DDI API `RANK_SHORT_NAMES` default/override. |
| `ASSET_BASE_URL` not set and images still serve from 7CAV | DDI API `milpacs/asset_url.go`, DDIApps `server/db/milpacDb.js` | Always set `ASSET_BASE_URL` in production `.env`; do not rely on the 7CAV default. |
| DDI award images use different dimensions and sprite-sheet offsets break | `client/public/skunkworks/uniformRibbons/` and `uniformMedals/` | The `generateAwardSprites.js` tool enforces tile geometry; DDI art must match the existing 43x14 ribbon and 70x120 medal tile sizes, or the tile geometry must be updated in code. |
| DDI API roster wire names drift and diff constants need re-normalization | DDIApps diff engine | Already handled by `normalizeRosterType()` in Phase 4.3. |

## 8. Summary

The 7CAV-specific content falls into three groups:

* **Pure code / generic data** that can stay: award class taxonomy, attachment device names, rank grade mapping (if IDs stable), and coordinate math.
* **Configurable data** that is already externalized in the DDI API but still defaults to 7CAV: rank short names, roster names, record names, and asset base URL.
* **Hard content that must be replaced**: all `client/public/skunkworks/` and `client/app/theme/` artwork, the award catalog, the MOS catalog, and external 7CAV links.

The highest-priority replacements are the **award catalog + sprite sheets** and **MOS constants**, because those directly affect whether the uniform builder renders correctly for DDI members. Branding links and logos should come first because they are the most visible 7CAV references.